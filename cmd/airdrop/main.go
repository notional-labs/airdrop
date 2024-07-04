package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/notional-labs/airdrop/internal/chains"
	"github.com/notional-labs/airdrop/internal/config"
	"github.com/notional-labs/airdrop/internal/logger"
	"github.com/notional-labs/airdrop/internal/queries"
	"github.com/notional-labs/airdrop/internal/query"
	"github.com/notional-labs/airdrop/internal/utils"
)

func main() {
	// Capture start time
	startTime := time.Now()

	// Setup logger
	logger, err := logger.Setup()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}
	defer logger.Sync() // Flushes buffer, if any

	// Parse command line flags
	configPath := flag.String("c", "configs/config.toml", "path to config file")
	blockHeight := flag.String("h", "latest", "block height to query")
	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Determine block height
	if *blockHeight == "latest" {
		*blockHeight, err = queries.GetLatestHeight(cfg.RPCServerAddress + "/status")
		if err != nil {
			logger.Fatal("Failed to fetch latest height", zap.Error(err))
		}
	} else {
		if _, err := strconv.Atoi(*blockHeight); err != nil {
			logger.Fatal("Please provide the block height as an integer", zap.Error(err))
		}
	}

	logger.Info("Using block height", zap.String("Block height", *blockHeight))

	// Setup gRPC connection
	conn, err := utils.SetupGRPCConnection(cfg.GRPCServerAddress)
	if err != nil {
		logger.Fatal("Failed to connect to GRPC server", zap.Error(err))
	}
	defer conn.Close()

	client := query.NewQueryClient(conn)

	// Perform airdrop calculation
	balanceInfo, err := chains.Airdrop(client.StakingClient, *configPath, *blockHeight, logger)
	if err != nil {
		logger.Fatal("Failed to calculate airdrop", zap.Error(err))
	}

	// Write balance info to file
	fileBalance, err := json.MarshalIndent(balanceInfo, "", " ")
	if err != nil {
		logger.Fatal("Failed to marshal balance info", zap.Error(err))
	}

	err = os.WriteFile("balance.json", fileBalance, 0o600)
	if err != nil {
		logger.Fatal("Failed to write balance info to file", zap.Error(err))
	}

	// Calculate and log total time duration
	duration := time.Since(startTime)
	logger.Info("Total time taken", zap.String("Duration", duration.String()))
}
