package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Parse Flags
	// Defaults empty to make them optional
	userBTC := flag.String("btc", "", "User BTC Address")
	userLTC := flag.String("ltc", "", "User LTC Address")
	userETH := flag.String("eth", "", "User ETH Address")
	
	// XPUB Flags (Defaults to Test Vector, but LTC updated to User Request)
	xpubBTC := flag.String("btc-xpub", "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8", "BTC XPUB")
	xpubLTC := flag.String("ltc-xpub", "xpub6DDn5pSVZKEuWmeJ8p84mpgVubu5bcu76EPp5Ue4irjBf7TDMr8zTkxtaFKL87f4sjK1MPArHgvRVeiwpXnX8GMtjthNRtMsaRXchTGZgc7", "LTC XPUB")
	xpubETH := flag.String("eth-xpub", "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8", "ETH XPUB")

	webhook := flag.String("webhook", "https://discord.com/api/webhooks/...", "Webhook URL")
	flag.Parse()

	_ = godotenv.Load()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		if host == "" { host = "localhost" }
		port := os.Getenv("DB_PORT")
		if port == "" { port = "5432" }
		user := os.Getenv("DB_USER")
		if user == "" { user = "eth_boi" }
		pass := os.Getenv("DB_PASS")
		if pass == "" { pass = "eth_boi" }
		name := os.Getenv("DB_NAME")
		if name == "" { name = "eth_boi" }
		if name == "" { name = "eth_boi" }
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	}

	// FORCE OVERRIDE for Test Script
	// To ensure we use the user's new DB even if env vars are sticky/old
	user := "eth_boi"
	pass := "eth_boi" 
	name := "eth_boi"
	host := "localhost"
	port := "5432"
	dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	
	fmt.Printf("DEBUG: Connecting to %s\n", dbURL)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	fmt.Println("✅ Connected to Database")

	fmt.Println("✅ Connected to Database")

	// 0. Ensure Tables Exist (Migration)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS merchants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			api_key TEXT UNIQUE NOT NULL,
			btc_xpub TEXT,
			ltc_xpub TEXT,
			eth_xpub TEXT,
			settings JSONB DEFAULT '{}',
			current_index INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS deposit_addresses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			address TEXT UNIQUE NOT NULL,
			coin TEXT NOT NULL,
			merchant_id UUID REFERENCES merchants(id),
			derivation_index INT NOT NULL,
			user_id TEXT,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tx_id TEXT NOT NULL,
			coin TEXT NOT NULL,
			amount NUMERIC NOT NULL,
			from_address TEXT,
			to_address TEXT,
			confirmations INT DEFAULT 0,
			status TEXT DEFAULT 'detected',
			merchant_id UUID REFERENCES merchants(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS app_state (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Failed to migrate tables: %v", err)
	}
	fmt.Println("✅ Tables Migrated/Verified")

	// 1. Create Merchant
	merchantName := "Test Merchant (Eth Boi)"
	webhookURL := *webhook
	
	// Use provided XPUBs
	btcXPub := *xpubBTC
	ltcXPub := *xpubLTC 
	ethXPub := *xpubETH

	// Auto-Detect Wallet Format Preferences
	settingsMap := map[string]string{"webhook_url": webhookURL}
	
	// Check LTC Format
	if *userLTC != "" {
		if (*userLTC)[0] == 'L' || (*userLTC)[0] == 'M' {
			fmt.Println("ℹ️  Detected Legacy LTC Address (L/M). Configuring Generator to Legacy Mode.")
			settingsMap["ltc_format"] = "legacy"
		} else if len(*userLTC) > 4 && (*userLTC)[:4] == "ltc1" {
			fmt.Println("ℹ️  Detected Native LTC Address (ltc1). Configuring Generator to Native Mode.")
			settingsMap["ltc_format"] = "native"
		}
	}

	// Check BTC Format
	if *userBTC != "" {
		if (*userBTC)[0] == '1' || (*userBTC)[0] == '3' {
			settingsMap["btc_format"] = "legacy"
		} else if len(*userBTC) > 3 && (*userBTC)[:3] == "bc1" {
			settingsMap["btc_format"] = "native"
		}
	}

	settingsBytes, _ := json.Marshal(settingsMap)

	// Clean up previous test merchant if exists
	// _, _ = db.Exec("DELETE FROM merchants WHERE name = $1", name)

	var merchantID string
	var apiKey string
	
	// Insert
	// We generate a random API Key
	apiKey = "test_api_key_123456789" // Fixed for easy testing
	
	err = db.QueryRow(`
		INSERT INTO merchants (name, api_key, btc_xpub, ltc_xpub, eth_xpub, settings, current_index)
		VALUES ($1, $2, $3, $4, $5, $6, 5)
		ON CONFLICT (api_key) DO UPDATE SET 
			name = EXCLUDED.name,
			btc_xpub = EXCLUDED.btc_xpub,
			ltc_xpub = EXCLUDED.ltc_xpub,
			eth_xpub = EXCLUDED.eth_xpub,
			settings = EXCLUDED.settings,
			current_index = 5
		RETURNING id
	`, merchantName, apiKey, btcXPub, ltcXPub, ethXPub, settingsBytes).Scan(&merchantID)

	if err != nil {
		log.Fatalf("Failed to insert merchant: %v", err)
	}

	fmt.Printf("✅ Merchant Created/Updated!\nID: %s\nAPI Key: %s\n", merchantID, apiKey)

	// 2. Insert the User's "Base Addresses" IF provided
	// This allows testing monitoring on an existing wallet without generating new ones.
	// If empty, we skip.
	
	if *userBTC != "" { insertAddr(db, merchantID, "BTC", *userBTC) }
	if *userLTC != "" { insertAddr(db, merchantID, "LTC", *userLTC) }
	if *userETH != "" { insertAddr(db, merchantID, "ETH", *userETH) }

	fmt.Println("✅ Merchant Setup Complete!")
}

func insertAddr(db *sql.DB, mID, coin, addr string) {
	_, err := db.Exec(`
		INSERT INTO deposit_addresses (merchant_id, coin, address, derivation_index, user_id, status)
		VALUES ($1, $2, $3, -1, 'static_base', 'active')
		ON CONFLICT (address) DO NOTHING
	`, mID, coin, addr)
	if err != nil {
		log.Printf("Failed to insert %s addr: %v", coin, err)
	} else {
		fmt.Printf("   -> %s Address: %s\n", coin, addr)
	}
}
