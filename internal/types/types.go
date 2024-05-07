package types

import (
	"encoding/json"
)

type Config struct {
	AddressPrefix     string `toml:"address_prefix"`
	GRPCServerAddress string `toml:"grpc_server_address"`
	RPCServerAddress  string `toml:"rpc_server_address"`
	CoinID            string `toml:"coin_id"`
}

type PriceSource struct {
	PriceSourceAPI string `toml:"price_source_api"`
}

type CoinID struct {
	CoinID string `toml:"coin_id"`
}

type AirdropTokenDenom struct {
	Denom string `toml:"airdrop_token_denom"`
}

type MinimumStakingTokensWorth struct {
	USD int `toml:"minimum_staking_tokens_worth"`
}

type AirdropDistributionTokens struct {
	TotalTokens int64 `toml:"airdrop_distribution"`
}

type NodeResponse struct {
	ID      json.Number `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  Result      `json:"result"`
}

type Result struct {
	SyncInfo SyncInfo `json:"sync_info"`
}

type SyncInfo struct {
	CatchingUp           bool   `json:"catching_up"`
	EarlieastAppHash     string `json:"earliest_app_hash"`
	EarlieastBlockHash   string `json:"earliest_block_hash"`
	EarlieastBlockHeight string `json:"earliest_block_height"`
	EarlieastBlockTime   string `json:"earliest_block_time"`
	LatestAppHash        string `json:"latest_app_hash"`
	LatestBlockHash      string `json:"latest_block_hash"`
	LatestBlockHeight    string `json:"latest_block_height"`
	LatestBlockTime      string `json:"latest_block_time"`
}
