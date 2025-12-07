package db

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Merchant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	APIKey       string    `json:"api_key"` // In real app, dont return this in all responses
	BTCXPub      string    `json:"btc_xpub"`
	LTCXPub      string    `json:"ltc_xpub"`
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
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateMerchant(name, btcXPub, ltcXPub string) (*Merchant, error) {
	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	m := &Merchant{
		Name:    name,
		APIKey:  apiKey,
		BTCXPub: btcXPub,
		LTCXPub: ltcXPub,
	}

	query := `INSERT INTO merchants (name, api_key, btc_xpub, ltc_xpub) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	
	err = DB.QueryRow(query, m.Name, m.APIKey, m.BTCXPub, m.LTCXPub).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func GetMerchantByAPIKey(apiKey string) (*Merchant, error) {
	m := &Merchant{}
	query := `SELECT id, name, api_key, btc_xpub, ltc_xpub, current_index, created_at FROM merchants WHERE api_key = $1`
	err := DB.QueryRow(query, apiKey).Scan(&m.ID, &m.Name, &m.APIKey, &m.BTCXPub, &m.LTCXPub, &m.CurrentIndex, &m.CreatedAt)
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
