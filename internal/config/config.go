package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	Port          string
	BTCRPCURL     string
	BTCRPCUser    string
	BTCRPCPass    string
	LTCRPCURL     string
	LTCRPCUser    string
	LTCRPCPass    string
	Testnet       bool
}

func LoadConfig() *Config {
	// Load .env file if it exists, but don't crash if it doesn't (production might use real env vars)
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          getEnv("PORT", "8080"),
		BTCRPCURL:     os.Getenv("BTC_RPC_URL"),
		BTCRPCUser:    os.Getenv("BTC_RPC_USER"),
		BTCRPCPass:    os.Getenv("BTC_RPC_PASS"),
		LTCRPCURL:     os.Getenv("LTC_RPC_URL"),
		LTCRPCUser:    os.Getenv("LTC_RPC_USER"),
		LTCRPCPass:    os.Getenv("LTC_RPC_PASS"),
		Testnet:       os.Getenv("TESTNET") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
