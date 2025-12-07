package api

import (
	"encoding/json"
	"net/http"

	"gateway/internal/config"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/ltc"
	"gateway/internal/db"
)

// CreateMerchantRequest
type CreateMerchantRequest struct {
	Name    string `json:"name"`
	BTCXPub string `json:"btc_xpub"`
	LTCXPub string `json:"ltc_xpub"`
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
	if req.BTCXPub == "" && req.LTCXPub == "" {
		http.Error(w, "At least one XPUB is required", http.StatusBadRequest)
		return
	}

	m, err := db.CreateMerchant(req.Name, req.BTCXPub, req.LTCXPub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
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
		merchant, err := db.GetMerchantByAPIKey(apiKey)
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

		resp := GenerateWalletsResponse{
			UserID:     req.UserID,
			MerchantID: merchant.ID,
			Wallets:    wallets,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
