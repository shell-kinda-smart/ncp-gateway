package main

import (
	"log"
	"net/http"
	"time"

	"gateway/internal/api"
	"gateway/internal/config"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/ltc"
	"gateway/internal/crypto/eth"
	"gateway/internal/db"
	"gateway/internal/monitor"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Initialize DB
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize Crypto Clients & Start Monitors
	// 3. Initialize Crypto Clients & Start Monitors (Async to not block API startup)
	go func() {
		btcClient := btc.NewClient(cfg.BTCRPCURL, cfg.BTCRPCUser, cfg.BTCRPCPass)

		// Start Monitors with Staggering (10s gap to avoid rate limits/load spike)
		log.Println("Starting BTC Monitor...")
		btcMonitor := monitor.NewBTCMonitor(btcClient, db.DB, cfg.Testnet)
		btcMonitor.Start()

		time.Sleep(10 * time.Second)

		ltcClient := ltc.NewClient(cfg.LTCRPCURL, cfg.LTCRPCUser, cfg.LTCRPCPass)
		log.Println("Starting LTC Monitor...")
		ltcMonitor := monitor.NewLTCMonitor(ltcClient, db.DB, cfg.Testnet)
		ltcMonitor.Start()

		if cfg.ETHRPCURL != "" {
			time.Sleep(10 * time.Second)
			log.Println("Starting ETH Monitor...")
			ethClient := eth.NewClient(cfg.ETHRPCURL)
			ethMonitor := monitor.NewETHMonitor(ethClient, db.DB)
			ethMonitor.Start()
		}
	}()

	// 4. Setup Router
	http.HandleFunc("/merchants", api.CreateMerchantHandler(cfg))
	http.HandleFunc("/wallets", api.GenerateWalletsHandler(cfg))
	http.HandleFunc("/withdraw", api.WithdrawHandler(cfg))

	// 5. Start Server
	log.Printf("Starting Payment Gateway on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
