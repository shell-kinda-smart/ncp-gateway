package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/ltc"
)

func main() {
	_ = godotenv.Load()

	btcURL := os.Getenv("BTC_RPC_URL")
	ltcURL := os.Getenv("LTC_RPC_URL")

	fmt.Println("--- Testing RPC Connectivity ---")

	// 1. Test BTC
	if btcURL != "" {
		fmt.Printf("Testing BTC URL: %s...\n", btcURL)
		client := btc.NewClient(btcURL, "", "")
		count, err := client.GetBlockCount()
		if err != nil {
			log.Printf("❌ BTC Error: %v\n", err)
		} else {
			fmt.Printf("✅ BTC SUCCESS! Current Block Height: %d\n", count)
		}
	} else {
		fmt.Println("⚠️ BTC_RPC_URL not set")
	}

	// 2. Test LTC
	if ltcURL != "" {
		fmt.Printf("Testing LTC URL: %s...\n", ltcURL)
		client := ltc.NewClient(ltcURL, "", "")
		count, err := client.GetBlockCount()
		if err != nil {
			log.Printf("❌ LTC Error: %v\n", err)
		} else {
			fmt.Printf("✅ LTC SUCCESS! Current Block Height: %d\n", count)
		}
	} else {
		fmt.Println("⚠️ LTC_RPC_URL not set")
	}
}
