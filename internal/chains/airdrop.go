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

func Airdrop(stakingClient stakingtypes.QueryClient, configPath, blockHeight string, logger *zap.Logger) (
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

	// Initialize delegators slice
	delegators := []stakingtypes.DelegationResponse{}

	// Get validators
	validators, err := queries.GetValidators(stakingClient, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}
	logger.Info("Fetched validators", zap.Int("totalValidators", len(validators)))

	// Get delegations for each validator
	for validatorIndex, validator := range validators {
		delegationsResponse, err := queries.GetValidatorDelegations(stakingClient, validator.OperatorAddress, blockHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to query delegate info for validator: %w", err)
		}
		total := delegationsResponse.Pagination.Total
		logger.Info("Fetched delegations", zap.Int("validatorIndex", validatorIndex), zap.Uint64("totalDelegators", total))
		delegators = append(delegators, delegationsResponse.DelegationResponses...)
	}

	// Fetch token price in USD
	priceSourceURL := priceSource + coinID + "&vs_currencies=usd"
	tokenPriceInUsd, err := queries.FetchTokenPrice(priceSourceURL, coinID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token price: %w", err)
	}
	logger.Info("Fetched token price", zap.String("priceSourceURL", priceSourceURL), zap.String("tokenPriceInUSD", tokenPriceInUsd.String()))

	// Calculate minimum tokens threshold
	usd := sdkmath.LegacyMustNewDecFromStr(minimumStakingTokensWorth)
	minimumTokensThreshold := usd.QuoTruncate(tokenPriceInUsd)
	logger.Info("Calculated minimum tokens threshold", zap.String("minimumStakingTokensWorth", minimumStakingTokensWorth), zap.String("minimumTokensThreshold", minimumTokensThreshold.String()))

	// Calculate total delegated tokens
	totalDelegatedTokens := sdkmath.LegacyMustNewDecFromStr("0")
	for _, delegator := range delegators {
		validatorIndex := utils.FindValidatorInfo(validators, delegator.Delegation.ValidatorAddress)
		validatorInfo := validators[validatorIndex]
		token := delegator.Delegation.Shares.MulInt(validatorInfo.Tokens).QuoTruncate(validatorInfo.DelegatorShares)
		totalDelegatedTokens = totalDelegatedTokens.Add(token)
	}
	logger.Debug("Calculated total delegated tokens", zap.String("totalDelegatedTokens", totalDelegatedTokens.String()))

	// Calculate airdrop tokens
	airdropTokens := sdkmath.LegacyMustNewDecFromStr(totalAirdropTokens)
	logger.Debug("Total tokens for airdrop", zap.String("airdropTokens", airdropTokens.String()))

	airdropMap := make(map[string]int64)
	checkAmount := int64(0)
	balanceInfo := []banktypes.Balance{}

	for _, delegator := range delegators {
		validatorIndex := utils.FindValidatorInfo(validators, delegator.Delegation.ValidatorAddress)
		validatorInfo := validators[validatorIndex]
		token := delegator.Delegation.Shares.MulInt(validatorInfo.Tokens).QuoTruncate(validatorInfo.DelegatorShares)
		// Remove account staking tokens worth less than threshold
		if token.LT(minimumTokensThreshold) {
			continue
		}

		logger.Debug("Delegator staking tokens", zap.String("delegatorAddress", delegator.Delegation.DelegatorAddress), zap.String("stakingTokens", token.String()))

		tokenAirdrop := airdropTokens.Mul(token).QuoTruncate(totalDelegatedTokens)
		bech32Address, err := utils.ConvertBech32Address(delegator.Delegation.DelegatorAddress, tokenDenom)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Bech32Address: %w", err)
		}

		logger.Debug("Airdrop tokens", zap.String("address", bech32Address), zap.String("tokensAirdrop", tokenAirdrop.String()))

		// Aggregate the tokens staked by the same address across multiple validators
		amount := airdropMap[bech32Address]
		airdropMap[bech32Address] = amount + tokenAirdrop.TruncateInt().Int64()
	}

	for address, amount := range airdropMap {
		// Skip addresses that receive less than 1 token
		if amount == 0 {
			continue
		}
		checkAmount += amount
		balanceInfo = append(balanceInfo, banktypes.Balance{
			Address: address,
			Coins:   sdktypes.NewCoins(sdktypes.NewCoin(tokenDenom, sdkmath.NewInt(amount))),
		})
	}

	logger.Info("Airdrop calculation complete", zap.Int64("totalAirdroppedTokens", checkAmount))
	return balanceInfo, nil
}
