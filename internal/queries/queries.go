package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sdkmath "cosmossdk.io/math"
	"github.com/cenkalti/backoff/v4"
	"github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	airdrop "github.com/notional-labs/airdrop/internal/backoff"
	"google.golang.org/grpc/metadata"

	"github.com/notional-labs/airdrop/internal/types"
	"github.com/notional-labs/airdrop/internal/utils"
)

const (
	LimitPerPage = 100000000
)

func GetValidators(stakingClient stakingtypes.QueryClient, blockHeight string) ([]stakingtypes.Validator, error) {
	// Get validator
	ctx := metadata.AppendToOutgoingContext(context.Background(), grpc.GRPCBlockHeightHeader, blockHeight)
	req := &stakingtypes.QueryValidatorsRequest{
		Pagination: &query.PageRequest{
			Limit: LimitPerPage,
		},
	}

	var resp *stakingtypes.QueryValidatorsResponse
	var err error

	exponentialBackoff := airdrop.NewBackoff(ctx)

	retryableRequest := func() error {
		resp, err = stakingClient.Validators(ctx, req)
		return err
	}

	if err := backoff.Retry(retryableRequest, exponentialBackoff); err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	if resp == nil || resp.Validators == nil {
		return nil, fmt.Errorf("validators response is nil")
	}

	return resp.Validators, nil
}

func GetValidatorDelegations(stakingClient stakingtypes.QueryClient, validatorAddr string, blockHeight string) (
	*stakingtypes.QueryValidatorDelegationsResponse, error,
) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), grpc.GRPCBlockHeightHeader, blockHeight)
	req := &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validatorAddr,
		Pagination: &query.PageRequest{
			CountTotal: true,
			Limit:      LimitPerPage,
		},
	}

	var resp *stakingtypes.QueryValidatorDelegationsResponse
	var err error

	exponentialBackoff := airdrop.NewBackoff(ctx)

	retryableRequest := func() error {
		resp, err = stakingClient.ValidatorDelegations(ctx, req)
		return err
	}

	if err := backoff.Retry(retryableRequest, exponentialBackoff); err != nil {
		return nil, fmt.Errorf("failed to get validator delegations: %w", err)
	}

	return resp, nil
}

func GetLatestHeight(apiURL string) (string, error) {
	ctx := context.Background()

	var response *http.Response
	var err error

	exponentialBackoff := airdrop.NewBackoff(ctx)

	retryableRequest := func() error {
		// Make a GET request to the API
		response, err = utils.MakeGetRequest(apiURL)
		return err
	}

	if err := backoff.Retry(retryableRequest, exponentialBackoff); err != nil {
		return "", fmt.Errorf("error making GET request to get latest height: %w", err)
	}

	defer response.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the response body into a NodeResponse struct
	var data types.NodeResponse
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	// Extract the latest block height from the response
	latestBlockHeight := data.Result.SyncInfo.LatestBlockHeight
	return latestBlockHeight, nil
}

func FetchTokenPrice(apiURL, coinID string) (sdkmath.LegacyDec, error) {
	ctx := context.Background()

	var response *http.Response
	var err error

	exponentialBackoff := airdrop.NewBackoff(ctx)

	retryableRequest := func() error {
		// Make a GET request to the API
		response, err = utils.MakeGetRequest(apiURL)
		return err
	}

	if err := backoff.Retry(retryableRequest, exponentialBackoff); err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("error making GET request to fetch token price: %w", err)
	}

	defer response.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("error reading response body for token price: %w", err)
	}

	var data interface{}
	// Unmarshal the JSON byte slice into the defined struct
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("error unmarshalling JSON for token price: %w", err)
	}
	tokenPrice := data.(map[string]interface{})
	priceInUsd := fmt.Sprintf("%v", tokenPrice[coinID].(map[string]interface{})["usd"])

	tokenPriceInUsd := sdkmath.LegacyMustNewDecFromStr(priceInUsd)
	return tokenPriceInUsd, nil
}
