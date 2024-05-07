package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/notional-labs/airdrop/internal/chains"
	"github.com/notional-labs/airdrop/internal/config"
	"github.com/notional-labs/airdrop/internal/logger"
	"github.com/notional-labs/airdrop/internal/queries"
	"github.com/notional-labs/airdrop/internal/query"
	"github.com/notional-labs/airdrop/internal/utils"
	"go.uber.org/zap"
)

func main() {
	// Capture start time
	startTime := time.Now()

	configPath := flag.String("c", "config.toml", "path to config file")
	flag.Parse()

	logger, err := logger.Setup()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	conn, err := utils.SetupGRPCConnection(cfg.GRPCServerAddress)
	if err != nil {
		logger.Fatal("Failed to connect to GRPC server", zap.Error(err))
	}

	client := query.NewQueryClient(conn)

	blockHeight, err := queries.GetLatestHeight(cfg.GRPCServerAddress)
	if err != nil {
		logger.Fatal("Failed to fetch latest height", zap.Error(err))
	}

	balanceInfo, err := chains.Composable(client.StakingClient, *configPath, blockHeight, logger)
	if err != nil {
		logger.Fatal("Failed to calculate airdrop for Composable", zap.Error(err))
	}

	fileBalance, _ := json.MarshalIndent(balanceInfo, "", " ")
	_ = os.WriteFile("balance.json", fileBalance, 0o600)

	// Calculate and print total time duration
	duration := time.Since(startTime)
	fmt.Printf("Total time taken: %v\n", duration)
}
