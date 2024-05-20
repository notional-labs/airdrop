package utils

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func MakeGetRequest(uri string) (*http.Response, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new HTTP GET request: %w", err)
	}

	// Send the HTTP request and get the response
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP GET request: %w", err)
	}

	return res, nil
}

func FindValidatorInfo(validators []stakingtypes.Validator, address string) int {
	for key, v := range validators {
		if v.OperatorAddress == address {
			return key
		}
	}
	return -1
}

func SetupGRPCConnection(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func ConvertBech32Address(dstChainAddress, srcDenom string) (string, error) {
	_, bz, err := bech32.DecodeAndConvert(dstChainAddress)
	if err != nil {
		return "", fmt.Errorf("error decoding address: %w", err)
	}
	newBech32DelAddr, err := bech32.ConvertAndEncode("pica", bz)
	if err != nil {
		return "", fmt.Errorf("error converting address: %w", err)
	}
	return newBech32DelAddr, nil
}
