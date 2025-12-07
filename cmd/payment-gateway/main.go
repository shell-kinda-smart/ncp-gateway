package main

import (
	"log"
	"net/http"

	"gateway/internal/api"
	"gateway/internal/config"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/ltc"
	"gateway/internal/db"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Initialize DB
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize Crypto Clients & Monitors
	btcClient := btc.NewClient(cfg.BTCRPCURL, cfg.BTCRPCUser, cfg.BTCRPCPass)
	btcMonitor := btc.NewMonitor(btcClient, db.DB, cfg.Testnet)
	btcMonitor.Start()

	ltcClient := ltc.NewClient(cfg.LTCRPCURL, cfg.LTCRPCUser, cfg.LTCRPCPass)
	ltcMonitor := ltc.NewMonitor(ltcClient, db.DB, cfg.Testnet)
	ltcMonitor.Start()

	// 4. Setup Router
	http.HandleFunc("/merchants", api.CreateMerchantHandler)
	http.HandleFunc("/wallets", api.GenerateWalletsHandler(cfg))

	// 5. Start Server
	log.Printf("Starting Payment Gateway on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
