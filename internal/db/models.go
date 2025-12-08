package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type Merchant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	APIKey       string    `json:"api_key"` // In real app, dont return this in all responses
	BTCXPub      string    `json:"btc_xpub"`
	LTCXPub      string    `json:"ltc_xpub"`
	ETHXPub      string    `json:"eth_xpub"` // Added for ETH support
	Settings     []byte    `json:"settings"` // Added for merchant settings
	CurrentIndex int       `json:"current_index"`
	CreatedAt    time.Time `json:"created_at"`
}

type DepositAddress struct {
	ID              string    `json:"id"`
	Address         string    `json:"address"`
	Coin            string    `json:"coin"`
	MerchantID      string    `json:"merchant_id"`
	DerivationIndex int       `json:"derivation_index"`
	UserID          string    `json:"user_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// GenerateAPIKey creates a random API Key
func GenerateAPIKey() string { // Changed return signature to only string
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// In a real application, you might want to log this error or panic
		// For this context, we'll return an empty string or a known error value
		// but the original intent was to return a string without error handling here.
		// For simplicity, assuming rand.Read will not fail in typical scenarios.
		return ""
	}
	return hex.EncodeToString(bytes)
}

func CreateMerchant(db *sql.DB, name, btcXPub, ltcXPub, ethXPub string, settings []byte) (*Merchant, error) { // Added db, ethXPub, settings parameters
	apiKey := GenerateAPIKey() // No error return from GenerateAPIKey now

	// Default settings to {} if nil
	if settings == nil {
		settings = []byte("{}")
	}

	merchant := &Merchant{ // Changed variable name from m to merchant
		Name:     name,
		APIKey:   apiKey,
		BTCXPub:  btcXPub,
		LTCXPub:  ltcXPub,
		ETHXPub:  ethXPub,  // Added ETHXPub
		Settings: settings, // Added Settings
	}

	// Updated query to include eth_xpub and settings, and only returning id
	query := `INSERT INTO merchants (name, api_key, btc_xpub, ltc_xpub, eth_xpub, settings) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at` // Added created_at back to returning

	err := db.QueryRow(query, merchant.Name, merchant.APIKey, merchant.BTCXPub, merchant.LTCXPub, merchant.ETHXPub, merchant.Settings).Scan(&merchant.ID, &merchant.CreatedAt) // Used db and merchant variable
	if err != nil {
		return nil, err
	}
	return merchant, nil // Return merchant
}

func GetMerchantByAPIKey(db *sql.DB, apiKey string) (*Merchant, error) { // Added db parameter
	m := &Merchant{}
	// Updated query to select eth_xpub, settings, and current_index
	query := `SELECT id, name, api_key, btc_xpub, ltc_xpub, eth_xpub, settings, current_index, created_at FROM merchants WHERE api_key = $1`
	err := db.QueryRow(query, apiKey).Scan(&m.ID, &m.Name, &m.APIKey, &m.BTCXPub, &m.LTCXPub, &m.ETHXPub, &m.Settings, &m.CurrentIndex, &m.CreatedAt) // Scan into new fields
	if err != nil {
		return nil, err
	}
	return m, nil
}

// IncrementMerchantIndex atomically increments and returns the new index
func IncrementMerchantIndex(merchantID string) (int, error) {
	var newIndex int
	// We increment first, so the first derived address is at index 1 (or 0 if we want 0-indexed, let's do 1-indexed for cleanliness usually, or 0. Let's do 0 based, so we return OLD and inc. 
	// Actually typical is: get current (0), return 0, set to 1.
	// Or: UPDATE merchants SET current_index = current_index + 1 RETURNING current_index
	// Let's use: return value used for this derivation.
	
	// We want to return the index TO BE USED.
	// Let's say we use current_index. Then we increment.
	err := DB.QueryRow(`UPDATE merchants SET current_index = current_index + 1 WHERE id = $1 RETURNING current_index`, merchantID).Scan(&newIndex)
	// If it was 0, it becomes 1. We can use 'newIndex - 1' as the path.
	return newIndex - 1, err
}

func SaveDepositAddress(merchantID, coin, address, userID string, index int) (*DepositAddress, error) {
	da := &DepositAddress{
		Address:         address,
		Coin:            coin,
		MerchantID:      merchantID,
		UserID:          userID,
		DerivationIndex: index,
		Status:          "active",
	}

	query := `INSERT INTO deposit_addresses (address, coin, merchant_id, derivation_index, user_id, status)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`

	err := DB.QueryRow(query, da.Address, da.Coin, da.MerchantID, da.DerivationIndex, da.UserID, da.Status).Scan(&da.ID, &da.CreatedAt)
	if err != nil {
		return nil, err
	}
	return da, nil
}
