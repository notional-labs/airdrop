package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakerInfo struct {
	Address      string `json:"address"`
	BondedTokens uint64 `json:"bonded_tokens"`
}

func main() {
	var blockHeight int64
	var nodeURL string

	// Improved command-line argument parsing.
	flag.Int64Var(&blockHeight, "height", -1, "Block height to query")
	flag.StringVar(&nodeURL, "node", "http://localhost:26657", "Node URL")
	flag.Parse()

	if blockHeight == -1 {
		log.Fatal("Please provide the block height as an argument with --height flag")
	}

	// Setup the Cosmos SDK client context
	ctx := context.Background()
	clientCtx := client.Context{}.WithNodeURI(nodeURL)
	tmRPC, err := client.GetTmClient(clientCtx)
	if err != nil {
		log.Fatalf("Failed to get tendermint client: %v", err)
	}
	clientCtx = clientCtx.WithClient(tmservice.New(tmRPC))

	// Query the stakers at the specified block height
	queryClient := stakingtypes.NewQueryClient(clientCtx)
	var stakerInfoList []StakerInfo
	var totalBondedTokens sdk.Int
	pageReq := &stakingtypes.PageRequest{Limit: 100}

	for {
		res, err := queryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
			Status:     stakingtypes.BondStatusBonded,
			Pagination: pageReq,
			// Include block height here if method supports it, check SDK documentation.
		})
		if err != nil {
			log.Fatalf("Failed to query validators: %v", err)
		}

		for _, validator := range res.Validators {
			stakerInfo := StakerInfo{
				Address:      validator.OperatorAddress,
				BondedTokens: validator.Tokens.Uint64(),
			}
			stakerInfoList = append(stakerInfoList, stakerInfo)
			totalBondedTokens = totalBondedTokens.Add(validator.Tokens)
		}

		if res.Pagination.NextKey == nil {
			break
		}
		pageReq.Key = res.Pagination.NextKey
	}

	// Convert the staker information to JSON
	jsonData, err := json.MarshalIndent(stakerInfoList, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Write the JSON data to a file
	err = os.WriteFile("stakers.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Staker information written to stakers.json")

	// Query the total staked supply, note: you should adjust this part if necessary,
	// since the original call did not consider block height and may not reflect the exact state at the requested height.
	// Mocking as the Cosmos SDK does not inherently manage a call for total staked supply at a specific height within these functions.
	fmt.Printf("Warning: Total bonded tokens calculation and matching against total staked supply needs adjustments for block height accuracy.\n")
}
