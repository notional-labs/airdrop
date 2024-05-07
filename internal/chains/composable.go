package chains

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"go.uber.org/zap"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/notional-labs/airdrop/internal/config"
	"github.com/notional-labs/airdrop/internal/queries"
	"github.com/notional-labs/airdrop/internal/utils"
)

func Composable(stakingClient stakingtypes.QueryClient, configPath, blockHeight string, logger *zap.Logger) (
	[]banktypes.Balance, error,
) {
	// Load config
	priceSource, err := config.LoadPriceSourceAPI(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load price source: %w", err)
	}

	coinID, err := config.LoadCoinID(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load coin id: %w", err)
	}

	totalAirdropTokens, err := config.LoadTotalAirdropTokens(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load total airdrop tokens: %w", err)
	}

	tokenDenom, err := config.LoadAirdropTokenDenom(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load airdrop token denom: %w", err)
	}

	minimumStakingTokensWorth, err := config.LoadMinimumStakingTokensWorth(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load minimum staking tokens worth: %w", err)
	}
	//

	delegators := []stakingtypes.DelegationResponse{}

	validators, err := queries.GetValidators(stakingClient, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}
	logger.Info("Composable", zap.Int("Total validator", len(validators)))

	for validatorIndex, validator := range validators {
		delegationsResponse, err := queries.GetValidatorDelegations(stakingClient, validator.OperatorAddress, blockHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to query delegate info for validator: %w", err)
		}
		total := delegationsResponse.Pagination.Total
		logger.Info(fmt.Sprintf("Validator index: %d", validatorIndex), zap.Uint64("Total delegators", total))
		delegators = append(delegators, delegationsResponse.DelegationResponses...)
	}

	usd := sdkmath.LegacyMustNewDecFromStr(minimumStakingTokensWorth)
	priceSourceURL := priceSource + coinID + "&vs_currencies=usd"
	tokenPriceInUsd, err := queries.FetchTokenPrice(priceSourceURL, coinID)
	logger.Info("Composable", zap.String("Token price in usd", tokenPriceInUsd.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token price: %w", err)
	}
	amountTokensWorth20Usd := usd.QuoTruncate(tokenPriceInUsd)
	logger.Info("", zap.String(fmt.Sprintf("Amount tokens worth $%s:", minimumStakingTokensWorth), amountTokensWorth20Usd.String()))

	// Caculate total delegated tokens
	totalDelegatedTokens := sdkmath.LegacyMustNewDecFromStr("0")
	for _, delegator := range delegators {
		validatorIndex := utils.FindValidatorInfo(validators, delegator.Delegation.ValidatorAddress)
		validatorInfo := validators[validatorIndex]
		token := (delegator.Delegation.Shares.MulInt(validatorInfo.Tokens)).QuoTruncate(validatorInfo.DelegatorShares)
		// Remove account staking tokens worth less than $20
		if token.LT(amountTokensWorth20Usd) {
			continue
		}
		totalDelegatedTokens = totalDelegatedTokens.Add(token)
	}

	logger.Debug("", zap.String("Total delegated tokens", totalDelegatedTokens.String()))

	airdropTokens := sdkmath.LegacyMustNewDecFromStr(totalAirdropTokens)
	logger.Debug("", zap.String("Total tokens for airdrop", airdropTokens.String()))

	airdropMap := make(map[string]int)
	checkAmount := 0
	balanceInfo := []banktypes.Balance{}
	for _, delegator := range delegators {
		validatorIndex := utils.FindValidatorInfo(validators, delegator.Delegation.ValidatorAddress)
		validatorInfo := validators[validatorIndex]
		token := (delegator.Delegation.Shares.MulInt(validatorInfo.Tokens)).QuoTruncate(validatorInfo.DelegatorShares)

		logger.Debug(
			fmt.Sprintf("Delegator address: %s", delegator.Delegation.DelegatorAddress),
			zap.String("Staking tokens", token.String()),
		)

		tokenAirdrop := airdropTokens.Mul(token).QuoTruncate(totalDelegatedTokens)
		bech32Address, err := utils.ConvertBech32Address(delegator.Delegation.DelegatorAddress, tokenDenom)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Bech32Address: %w", err)
		}

		logger.Debug(fmt.Sprintf("Address: %s", bech32Address), zap.String("Tokens airdrop", tokenAirdrop.String()))

		// Aggregate the tokens staked by the same address across multiple validators
		amount := airdropMap[bech32Address]
		coins := sdktypes.NewCoins(sdktypes.NewCoin(tokenDenom, tokenAirdrop.TruncateInt()))
		airdropMap[bech32Address] = amount + int(coins.AmountOf(tokenDenom).Int64())
	}
	for address, amount := range airdropMap {
		// Skip addresses that receive less than 1 token
		if amount == 0 {
			continue
		}
		checkAmount += amount
		balanceInfo = append(balanceInfo, banktypes.Balance{
			Address: address,
			Coins:   sdktypes.NewCoins(sdktypes.NewCoin(tokenDenom, sdkmath.NewInt(int64(amount)))),
		})
	}
	logger.Info("Total balance: ", zap.Int("total balance", checkAmount))
	return balanceInfo, nil
}
