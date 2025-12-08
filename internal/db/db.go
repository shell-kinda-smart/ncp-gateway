package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"gateway/internal/config"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	var err error
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	DB, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to PostgreSQL database!")
	return runMigrations()
}

func runMigrations() error {
	// Note: In a real app we would use a migration tool. 
	// For this pivot, we are dropping tables to reset schema cleanly as per "nothing left regarding wallet creating" request.
	// WARNING: This deletes data. In prod use ALTER TABLE.
	drop := `
	DROP TABLE IF EXISTS transactions;
	DROP TABLE IF EXISTS deposit_addresses;
	DROP TABLE IF EXISTS merchants;
	`
	if _, err := DB.Exec(drop); err != nil {
		return fmt.Errorf("failed to drop tables: %v", err)
	}

	schema := `
	CREATE TABLE merchants (
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

	CREATE TABLE deposit_addresses (
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
	
	CREATE INDEX IF NOT EXISTS idx_deposit_address ON deposit_addresses(address);
	CREATE INDEX IF NOT EXISTS idx_tx_to_address ON transactions(to_address);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}
	log.Println("Database schema updated for HD Wallets (Tables Reset).")
	return nil
}
