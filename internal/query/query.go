package query

import (
	"google.golang.org/grpc"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Client struct {
	// Cosmos-sdk query clients
	StakingClient stakingtypes.QueryClient
}

func NewQueryClient(conn *grpc.ClientConn) *Client {
	client := &Client{
		StakingClient: stakingtypes.NewQueryClient(conn),
	}
	return client
}
