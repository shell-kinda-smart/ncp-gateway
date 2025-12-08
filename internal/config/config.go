package config

import (
	"log"
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

	ETHRPCURL string
	Testnet       bool
}

func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it")
	} else {
		log.Println("Loaded .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try to construct from components
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		pass := getEnv("DB_PASS", "postgres")
		name := getEnv("DB_NAME", "gateway_db")
		// Assume disable ssl for local dev flexibility if not specified in URL
		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=disable"
	}

	return &Config{
		DatabaseURL:   dbURL,
		Port:          getEnv("PORT", "8080"),
		BTCRPCURL:     os.Getenv("BTC_RPC_URL"),
		BTCRPCUser:    os.Getenv("BTC_RPC_USER"),
		BTCRPCPass:    os.Getenv("BTC_RPC_PASS"),
		LTCRPCURL:     getEnv("LTC_RPC_URL", "http://localhost:9332"),
		LTCRPCUser:    getEnv("LTC_RPC_USER", ""),
		LTCRPCPass:    getEnv("LTC_RPC_PASS", ""),

		ETHRPCURL:     getEnv("ETH_RPC_URL", ""),
		Testnet:       os.Getenv("TESTNET") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
