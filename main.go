package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
    stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakerInfo struct {
    Address      string `json:"address"`
    BondedTokens uint64 `json:"bonded_tokens"`
}

func main() {
    // Parse command-line arguments
    nodeURL := flag.String("node", "http://localhost:26657", "Node URL")
    flag.Parse()

    // Get the block height from the command-line argument
    if len(os.Args) < 2 {
        log.Fatal("Please provide the block height as an argument")
    }
    blockHeight, err := strconv.ParseInt(os.Args[1], 10, 64)
    if err != nil {
        log.Fatalf("Invalid block height: %v", err)
    }

    // Set up the Cosmos SDK client
    ctx := context.Background()
    clientCtx := client.Context{}.WithNodeURI(*nodeURL)
    clientCtx = clientCtx.WithClient(tmservice.New(clientCtx))

    // Query the stakers at the specified block height
    queryClient := stakingtypes.NewQueryClient(clientCtx)
    var stakerInfoList []StakerInfo
    var totalBondedTokens uint64
    pageReq := &stakingtypes.PageRequest{Limit: 100}

    for {
        res, err := queryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
            Status:     stakingtypes.BondStatusBonded,
            Pagination: pageReq,
        })
        if err != nil {
            log.Fatal(err)
        }

        for _, validator := range res.Validators {
            stakerInfo := StakerInfo{
                Address:      validator.OperatorAddress,
                BondedTokens: validator.Tokens.Uint64(),
            }
            stakerInfoList = append(stakerInfoList, stakerInfo)
            totalBondedTokens += validator.Tokens.Uint64()
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

    // Query the total staked supply at the specified block height
    supplyRes, err := queryClient.TotalSupply(ctx, &stakingtypes.QueryTotalSupplyRequest{})
    if err != nil {
        log.Fatal(err)
    }

    // Check the total bonded tokens against the staked supply
    if totalBondedTokens != supplyRes.Supply.Amount.Uint64() {
        fmt.Printf("Warning: Total bonded tokens (%d) does not match the staked supply (%d)\n", totalBondedTokens, supplyRes.Supply.Amount.Uint64())
    } else {
        fmt.Println("Total bonded tokens matches the staked supply")
    }
}
