package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/notional-labs/airdrop/internal/types"
)

func LoadConfig(configPath string) (*types.Config, error) {
	var config types.Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadPriceSourceAPI(configPath string) (string, error) {
	var priceSource types.PriceSource
	if _, err := toml.DecodeFile(configPath, &priceSource); err != nil {
		return "", err
	}
	return priceSource.PriceSourceAPI, nil
}

func LoadCoinID(configPath string) (string, error) {
	var coinID types.CoinID
	if _, err := toml.DecodeFile(configPath, &coinID); err != nil {
		return "", err
	}
	return coinID.CoinID, nil
}

func LoadAirdropTokenDenom(configPath string) (string, error) {
	var tokenDenom types.AirdropTokenDenom
	if _, err := toml.DecodeFile(configPath, &tokenDenom); err != nil {
		return "", err
	}
	return tokenDenom.Denom, nil
}

func LoadMinimumStakingTokensWorth(configPath string) (string, error) {
	var minimumStakingTokensWorth types.MinimumStakingTokensWorth
	if _, err := toml.DecodeFile(configPath, &minimumStakingTokensWorth); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", minimumStakingTokensWorth.USD), nil
}

func LoadTotalAirdropTokens(configPath string) (string, error) {
	var airdropDistributionTokens types.AirdropDistributionTokens
	if _, err := toml.DecodeFile(configPath, &airdropDistributionTokens); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", airdropDistributionTokens.TotalTokens), nil
}
