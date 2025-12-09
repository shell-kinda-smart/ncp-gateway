package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gateway/internal/config"
	"gateway/internal/crypto/btc"
	"gateway/internal/crypto/eth"
	"gateway/internal/crypto/generator"
	"gateway/internal/crypto/ltc"
	"gateway/internal/db"
	"gateway/internal/security"
)

// CreateMerchantRequest (Managed)
type CreateMerchantRequest struct {
	Name             string `json:"name"`
	PayoutAddressBTC string `json:"payout_address_btc"`
	PayoutAddressLTC string `json:"payout_address_ltc"`
	PayoutAddressETH string `json:"payout_address_eth"`
	WebhookURL       string `json:"webhook_url"` // Added Webhook Support
	Passphrase       string `json:"passphrase"` // Optional salt
}

type MerchantResponse struct {
	ID       string `json:"id"`
	APIKey   string `json:"api_key"`
	Mnemonic string `json:"mnemonic"` // ONE TIME VIEW
	Warning  string `json:"warning"`
}

// CreateMerchantHandler generates keys for the user (Pure Managed)
func CreateMerchantHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		var req CreateMerchantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		
		// 1. Generate Wallet (BIP39) - Enforces Standards (BIP84 BTC/LTC)
		keys, err := generator.GenerateNewWallet(cfg.Testnet)
		if err != nil {
			http.Error(w, "Key Generation Failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		

		// 2. Encrypt Mnemonic
		masterKey := cfg.MasterKey
		if masterKey == "" {
             log.Println("CRITICAL: MASTER_KEY not set")
             http.Error(w, "Configuration Error: Secure Storage Unavailable", http.StatusInternalServerError)
             return
        }

		encryptedMnemonic, err := security.Encrypt(keys.Mnemonic, masterKey)
		if err != nil {
			http.Error(w, "Encryption Failed", http.StatusInternalServerError)
			return
		}
		
		// 3. Save to DB (Explicitly set format to native/best)
		settings := map[string]string{
			"btc_format":  "native",
			"ltc_format":  "native",
			"webhook_url": req.WebhookURL,
		}
		settingsBytes, _ := json.Marshal(settings)
		
		// Using Custodial Creator
		m, err := db.CreateCustodialMerchant(db.DB, req.Name, encryptedMnemonic, "default_salt", 
			req.PayoutAddressBTC, req.PayoutAddressLTC, req.PayoutAddressETH,
			keys.BTCXPub, keys.LTCXPub, keys.ETHXPub, settingsBytes)
			
		if err != nil {
			http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		

		// 4. Import to Bitcoin Node (Watch Only) to track UTXOs
		// We fire and forget this logging error if it fails, or block?
		// Better to block to ensure it works.
		if keys.BTCXPub != "" {
			client := btc.NewClient(cfg.BTCRPCURL, cfg.BTCRPCUser, cfg.BTCRPCPass)
			// Construct Descriptor: wpkh(xpub.../0/*) and wpkh(xpub.../1/*) for change
			// Note: We use the raw key string. Ideally strictly typed.
			desc := fmt.Sprintf("wpkh(%s/0/*)", keys.BTCXPub)
			// We need checksum? bitcoind might require it or #checksum.
			// Actually active=true puts it in the wallet.
			req := btc.ImportDescriptorRequest{
				Desc:      desc,
				Timestamp: "now",
				Active:    true,
				Internal:  false,
				Range:     []int{0, 1000},
				Label:     "merchant:" + m.ID,
			}
			// Also import change chain? internal=true
			descChange := fmt.Sprintf("wpkh(%s/1/*)", keys.BTCXPub)
			reqChange := btc.ImportDescriptorRequest{
				Desc:      descChange,
				Timestamp: "now",
				Active:    true,
				Internal:  true,
				Range:     []int{0, 1000},
				Label:     "merchant:" + m.ID,
			}
			
			err := client.ImportDescriptors([]btc.ImportDescriptorRequest{req, reqChange})
			if err != nil {
				// Don't fail the request, but log it.
				// In PROD this is critical.
				// log.Printf("Failed to import descriptors: %v", err)
			}
		}

		// 5. Return Secrets (One Time)
		resp := MerchantResponse{
			ID:       m.ID,
			APIKey:   m.APIKey,
			Mnemonic: keys.Mnemonic,
			Warning:  "WRITE THIS DOWN. WE DO NOT DISPLAY IT AGAIN. YOU OWN THIS WALLET.",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
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
		// Authenticate
		apiKey := r.Header.Get("X-API-Key")
		merchant, err := db.GetMerchantByAPIKey(db.DB, apiKey)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			// GET: Fetch Wallets for a UserID (Query Param)
			userID := r.URL.Query().Get("user_id")
			if userID == "" {
				http.Error(w, "Missing user_id query parameter", http.StatusBadRequest)
				return
			}
			
			wallets, err := db.GetDepositAddressesByUserID(merchant.ID, userID)
			if err != nil {
				http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			if wallets == nil {
				wallets = []*db.DepositAddress{}
			}

			resp := GenerateWalletsResponse{
				UserID:     userID,
				MerchantID: merchant.ID,
				Wallets:    wallets,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return

		} else if r.Method == http.MethodPost {
			// POST: Generate (Idempotent)
			var req GenerateWalletsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if req.UserID == "" {
				http.Error(w, "user_id is required", http.StatusBadRequest)
				return
			}

			// 0. Check Existing Wallets (Idempotency)
			existingWallets, err := db.GetDepositAddressesByUserID(merchant.ID, req.UserID)
			if err == nil && len(existingWallets) > 0 {
				resp := GenerateWalletsResponse{
					UserID:     req.UserID,
					MerchantID: merchant.ID,
					Wallets:    existingWallets,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
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
				var settingsMap map[string]interface{}
				_ = json.Unmarshal(merchant.Settings, &settingsMap)
				
				btcWallet := btc.NewWallet(cfg.Testnet)
				
				// Enforce Native since we set it in creation
				btcType := "native"
				if val, ok := settingsMap["btc_format"].(string); ok && val != "" {
					btcType = val
				}

				addr, err := btcWallet.DeriveAddressByType(merchant.BTCXPub, uint32(index), btcType)
				if err != nil {
					// Fallback to Detect if needed, but for managed it should rely on settings
					http.Error(w, "BTC Derivation Error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				da, err := db.SaveDepositAddress(merchant.ID, "BTC", addr, req.UserID, index, btcType)
				if err == nil {
					wallets = append(wallets, da)
				}
			}

			// 3. Derive LTC
			if merchant.LTCXPub != "" {
				ltcWallet := ltc.NewWallet(cfg.Testnet)
				var settingsMap map[string]interface{}
				_ = json.Unmarshal(merchant.Settings, &settingsMap)
				
				addrType := "native"
				if val, ok := settingsMap["ltc_format"].(string); ok {
					addrType = val
				}

				addr, err := ltcWallet.DeriveAddressByType(merchant.LTCXPub, uint32(index), addrType)
				if err != nil {
					http.Error(w, "LTC Derivation Error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				da, err := db.SaveDepositAddress(merchant.ID, "LTC", addr, req.UserID, index, addrType)
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
				da, err := db.SaveDepositAddress(merchant.ID, "ETH", addr, req.UserID, index, "standard")
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

		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// WithdrawRequest
type WithdrawRequest struct {
	Coin    string `json:"coin"`   // BTC, LTC, ETH
	Address string `json:"address"` // Destination
	Amount  string `json:"amount"`  // "max" or specific float string
}

type WithdrawResponse struct {
	TxID   string `json:"tx_id"`
	Amount string `json:"amount"` 
	Fee    string `json:"fee"`
	Status string `json:"status"`
}

// WithdrawHandler triggers a payout
func WithdrawHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 1. Auth (using X-API-Key)
		apiKey := r.Header.Get("X-API-Key")
		merchant, err := db.GetMerchantByAPIKey(db.DB, apiKey)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		var req WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		log.Printf("Withdraw Request: Merchant=%s Coin=%s Amount=%s Dest=%s", merchant.Name, req.Coin, req.Amount, req.Address)


		// 2. Decrypt Mnemonic (HOT WALLET ACTIVATE)
		// We need the master key to decrypt the user's seed.
		masterKey := cfg.MasterKey 
		mnemonic, err := security.Decrypt(merchant.EncryptedMnemonic, masterKey)
		if err != nil {
			// Do not return exact error for security
			log.Printf("Decryption Failed for Merchant %s: %v", merchant.ID, err)
			http.Error(w, "Security Error: Failed to unlock wallet", http.StatusForbidden)
			return
		}
		
		// Validate that Mnemonic is real (sanity check)
		if mnemonic == "" {
             http.Error(w, "Wallet Corruption Error", http.StatusInternalServerError)
             return
        }

		// 3. Perform Withdrawal (Sign & Broadcast)
		var txID string
		var feeStr = "estimated"
		
		if req.Coin == "BTC" {
			client := btc.NewClient(cfg.BTCRPCURL, cfg.BTCRPCUser, cfg.BTCRPCPass)
			signer := btc.NewSigner(client, db.DB, cfg.Testnet)
			
			txID, err = signer.Send(mnemonic, req.Address, req.Amount, merchant.ID)
			if err != nil {
				http.Error(w, "Withdrawal Failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Stub for LTC/ETH
			http.Error(w, "Coin not implemented yet: "+req.Coin, http.StatusBadRequest)
			return
		}
		
		resp := WithdrawResponse{
			TxID:   txID,
			Amount: req.Amount,
			Fee:    feeStr,
			Status: "broadcasted",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
