package api

import (
	"encoding/json"
	"net/http"

	"gateway/internal/config"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/eth"
	"gateway/internal/crypto/ltc"
	"gateway/internal/db"
)

// CreateMerchantRequest
type CreateMerchantRequest struct {
	Name    string `json:"name"`
	BTCXPub string `json:"btc_xpub"`
	LTCXPub string `json:"ltc_xpub"`
	ETHXPub string `json:"eth_xpub"`
	Settings json.RawMessage `json:"settings"`
}

type MerchantResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	APIKey  string `json:"api_key"`
	BTCXPub string `json:"btc_xpub"`
	LTCXPub string `json:"ltc_xpub"`
	ETHXPub string `json:"eth_xpub"`
}

func CreateMerchantHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic Validation
	if req.BTCXPub == "" && req.LTCXPub == "" && req.ETHXPub == "" {
		http.Error(w, "At least one XPUB is required", http.StatusBadRequest)
		return
	}

	m, err := db.CreateMerchant(db.DB, req.Name, req.BTCXPub, req.LTCXPub, req.ETHXPub, req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := MerchantResponse{
		ID:      m.ID,
		Name:    m.Name,
		APIKey:  m.APIKey,
		BTCXPub: m.BTCXPub,
		LTCXPub: m.LTCXPub,
		ETHXPub: m.ETHXPub,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GenerateWalletsRequest
type GenerateWalletsRequest struct {
	UserID string `json:"user_id"`
}

type GenerateWalletsResponse struct {
	UserID     string              `json:"user_id"`
	MerchantID string              `json:"merchant_id"`
	Wallets    []*db.DepositAddress `json:"wallets"`
}

func GenerateWalletsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		merchant, err := db.GetMerchantByAPIKey(db.DB, apiKey)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req GenerateWalletsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 1. Get New Index
		index, err := db.IncrementMerchantIndex(merchant.ID)
		if err != nil {
			http.Error(w, "Failed to increment index: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		var wallets []*db.DepositAddress

		// 2. Derive BTC
		if merchant.BTCXPub != "" {
			btcWallet := btc.NewWallet(cfg.Testnet)
			addr, err := btcWallet.DeriveAddress(merchant.BTCXPub, uint32(index))
			if err != nil {
				http.Error(w, "BTC Derivation Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			da, err := db.SaveDepositAddress(merchant.ID, "BTC", addr, req.UserID, index)
			if err == nil {
				wallets = append(wallets, da)
			}
		}

		// 3. Derive LTC
		if merchant.LTCXPub != "" {
			ltcWallet := ltc.NewWallet(cfg.Testnet)
			addr, err := ltcWallet.DeriveAddress(merchant.LTCXPub, uint32(index))
			if err != nil {
				http.Error(w, "LTC Derivation Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			da, err := db.SaveDepositAddress(merchant.ID, "LTC", addr, req.UserID, index)
			if err == nil {
				wallets = append(wallets, da)
			}
		}

		// 4. Derive ETH
		if merchant.ETHXPub != "" {
			addr, err := eth.DeriveAddress(merchant.ETHXPub, uint32(index))
			if err != nil {
				http.Error(w, "ETH Derivation Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			da, err := db.SaveDepositAddress(merchant.ID, "ETH", addr, req.UserID, index)
			if err == nil {
				wallets = append(wallets, da)
			}
		}

		resp := GenerateWalletsResponse{
			UserID:     req.UserID,
			MerchantID: merchant.ID,
			Wallets:    wallets,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
