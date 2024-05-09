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

	logger, err := logger.Setup()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	configPath := flag.String("c", "configs/config.toml", "path to config file")
	blockHeight := flag.String("h", "latest", "block height to query")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	if *blockHeight == "latest" {
		var err error
		*blockHeight, err = queries.GetLatestHeight(cfg.RPCServerAddress + "/status")
		if err != nil {
			log.Fatal("Failed to fetch latest height", zap.Error(err))
		}
	} else {
		if _, err := strconv.Atoi(*blockHeight); err != nil {
			log.Fatal("Please provide the block height as an integer")
		}
	}

	logger.Info("", zap.String("Block height", *blockHeight))

	conn, err := utils.SetupGRPCConnection(cfg.GRPCServerAddress)
	if err != nil {
		logger.Fatal("Failed to connect to GRPC server", zap.Error(err))
	}
	defer conn.Close()

	client := query.NewQueryClient(conn)

	balanceInfo, err := chains.Airdrop(client.StakingClient, *configPath, *blockHeight, logger)
	if err != nil {
		logger.Fatal("Failed to calculate airdrop", zap.Error(err))
	}

	fileBalance, _ := json.MarshalIndent(balanceInfo, "", " ")
	_ = os.WriteFile("balance.json", fileBalance, 0o600)

	// Calculate and print total time duration
	duration := time.Since(startTime)
	logger.Info("", zap.String("Total time taken", duration.String()))
}
