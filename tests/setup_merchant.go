package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	
	// Import models to use GenerateAPIKey if exported or just raw SQL
	// "gateway/internal/db" 
)

// We define a simplified process here to avoid importing the whole app if db package is complex
// But we should use the internal db package if possible.
// Given imports, let's just use raw SQL for simplicity and robustness in a script.

func main() {
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
	webhookURL := "https://discord.com/api/webhooks/1416434601303343267/sQ5v-PjR34udmz6JNu9fF9idX1h3QHKwnmqWpDcTBOdZs8T40CvpPSMOO4n__JEhWybS"
	
	// User provided Addresses used as XPUBs?? NO.
	// We MUST use real XPUBs for derivation to work.
	// I will use standard test vector xpubs.
	// BTC: xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8
	// LTC: Ltub288szl... (using BTC xpub format usually works for LTC logic if using btcsuite, but technically version bytes differ. Let's use a standard BTC one for LTC too since our logic mostly cares about the key data, or use a known LTC xpub).
	// ETH: xpub... (ETH doesn't strictly use xpub in standard clients but our code expects it).
	// I will use the SAME xpub for all 3 for simplicity of testing derivation, or distinct ones.
	
	btcXPub := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"
	ltcXPub := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8" 
	ethXPub := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"

	settings := map[string]string{"webhook_url": webhookURL}
	settingsBytes, _ := json.Marshal(settings)

	// Clean up previous test merchant if exists
	// _, _ = db.Exec("DELETE FROM merchants WHERE name = $1", name)

	var merchantID string
	var apiKey string
	
	// Insert
	// We generate a random API Key
	apiKey = "test_api_key_123456789" // Fixed for easy testing
	
	err = db.QueryRow(`
		INSERT INTO merchants (name, api_key, btc_xpub, ltc_xpub, eth_xpub, settings)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (api_key) DO UPDATE SET 
			name = EXCLUDED.name,
			btc_xpub = EXCLUDED.btc_xpub,
			ltc_xpub = EXCLUDED.ltc_xpub,
			eth_xpub = EXCLUDED.eth_xpub,
			settings = EXCLUDED.settings
		RETURNING id
	`, merchantName, apiKey, btcXPub, ltcXPub, ethXPub, settingsBytes).Scan(&merchantID)

	if err != nil {
		log.Fatalf("Failed to insert merchant: %v", err)
	}

	fmt.Printf("✅ Merchant Created/Updated!\nID: %s\nAPI Key: %s\n", merchantID, apiKey)

	// 2. Insert the User's "Base Addresses" as a reference? 
	// The user asked to "add public adress is going to according to base adress". 
	// Maybe they want these specifically tracked? 
	// In our HD system, we track *derived* addresses.
	// I will insert these specific addresses into `deposit_addresses` so the monitor watches them!
	// This is a "hack" to support their static addresses in our HD system.
	// We will give them index -1 or something special.
	
	userBTC := "bc1q2ykugly7f0x6fxdu9pegckzhzgk3lpu3u6ad0c"
	userLTC := "LapN9jvkQMXmmYoj5XCBV7MRWuDrnWXb2V"
	userETH := "0x1189f4AdB7D2A79E5595CD5CfE0fa22DF6caa50D"

	insertAddr(db, merchantID, "BTC", userBTC)
	insertAddr(db, merchantID, "LTC", userLTC)
	insertAddr(db, merchantID, "ETH", userETH)

	fmt.Println("✅ Static Addresses injected for monitoring!")
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
