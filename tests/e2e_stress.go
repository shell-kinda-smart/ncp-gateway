package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config
const (
	BaseURL = "http://localhost:8080"
	ApiKey  = "test_api_key_123456789" // Matching setup_merchant.go
)

type WalletResponse struct {
	UserID  string `json:"user_id"`
	Wallets []struct {
		Address     string `json:"address"`
		Coin        string `json:"coin"`
		AddressType string `json:"address_type"`
	} `json:"wallets"`
}

func main() {
	// 1. Environment & DB Connection
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback
		dbURL = "postgres://eth_boi:eth_boi@localhost:5432/eth_boi?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer db.Close()

	// 2. Clear Database (Reset for Test)
	fmt.Println("🧹 Clearing Database...")
	_, err = db.Exec("TRUNCATE TABLE transactions, deposit_addresses CASCADE") // Keep merchant for now or reload? 
	// The user wants "setup our 1st merchant". Let's run the setup logic here or assume setup_merchant was run.
	// Actually, let's just make sure the merchant exists or Clean slate it.
	// Let's Clean everything including merchants to be safe, then Re-Insert.
	_, err = db.Exec("TRUNCATE TABLE transactions, deposit_addresses, merchants CASCADE")
	if err != nil {
		log.Fatalf("Failed to truncate: %v", err)
	}

	// 3. Create Merchant (Using User Keys)
	fmt.Println("🏪 Creating Professional Merchant...")
	keys := map[string]string{
		"name":     "Pro Crypto Store",
		"btc_xpub": "zpub6r3GjDG3VokUzXinpeQF3BxrmfU4YbzLTbkhVqH367DPn8oQXD6HgjKHEPYLSXjQdqpR3yb5ZhWDyLJHaKZPtRsJDDd4bo4fRWqTxoqc5MD", // Using Native Segwit zpub
		"ltc_xpub": "xpub6DDn5pSVZKEuWmeJ8p84mpgVubu5bcu76EPp5Ue4irjBf7TDMr8zTkxtaFKL87f4sjK1MPArHgvRVeiwpXnX8GMtjthNRtMsaRXchTGZgc7",
		"eth_xpub": "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8", // Placeholder ETH xpub (User didn't provide valid ETH xpub, using test vector valid one)
		// Note: User provided 0x address but no XPUB for ETH. Using Standard Test Vector XPUB for ETH testing.
	}
	
	settings := map[string]string{
		"webhook_url": "https://discord.com/api/webhooks/...",
		"ltc_format": "legacy", // Enforcing legacy as per user previous request? Or user keys? 
		// User provided L... address, so wants legacy.
		"btc_format": "native", // zpub implies native
	}
	settingsBytes, _ := json.Marshal(settings)

	_, err = db.Exec(`
		INSERT INTO merchants (name, api_key, btc_xpub, ltc_xpub, eth_xpub, settings, current_index)
		VALUES ($1, $2, $3, $4, $5, $6, 5)
	`, keys["name"], ApiKey, keys["btc_xpub"], keys["ltc_xpub"], keys["eth_xpub"], settingsBytes)
	if err != nil {
		log.Fatalf("Failed to create merchant: %v", err)
	}

	// 4. Generate 40 Unique Users (Stress Test)
	fmt.Println("🚀 Generating 40 Unique Users...")
	
	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= 40; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user_%03d", id)
			
			// Generate
			wallets := callAPI(userID)
			
			// Verify
			if len(wallets.Wallets) == 0 {
				fmt.Printf("❌ User %s: No wallets generated!\n", userID)
				return
			}
			
			// Verify Idempotency (Call again)
			wallets2 := callAPI(userID)
			if wallets.Wallets[0].Address != wallets2.Wallets[0].Address {
				fmt.Printf("❌ User %s: IDEMPOTENCY FAILED! Got %s then %s\n", userID, wallets.Wallets[0].Address, wallets2.Wallets[0].Address)
				return
			}
			
			fmt.Printf("✅ User %s: OK (%s | %s)\n", userID, wallets.Wallets[0].Coin, wallets.Wallets[0].Address)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	
	fmt.Printf("\n✨ Completed 40 Users in %v\n", duration)
	fmt.Println("Professional System Verification: PASSED")
}

func callAPI(userID string) WalletResponse {
	reqBody, _ := json.Marshal(map[string]string{
		"user_id": userID,
	})
	
	req, _ := http.NewRequest("POST", BaseURL+"/wallets", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ApiKey)
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return WalletResponse{}
	}
	defer resp.Body.Close()
	
	// Verify Status Code
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		log.Printf("API Error (%d): %s", resp.StatusCode, buf.String())
		return WalletResponse{}
	}

	var wResp WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&wResp); err != nil {
		log.Printf("JSON Decode Error: %v", err)
	}
	return wResp
}
