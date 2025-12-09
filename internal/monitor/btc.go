package monitor

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"gateway/internal/crypto/btc"
	"gateway/internal/price"
	"gateway/internal/worker"
)

type BTCMonitor struct {
	Client    *btc.Client
	DB        *sql.DB
	NetParams *chaincfg.Params
}

func NewBTCMonitor(client *btc.Client, db *sql.DB, testnet bool) *BTCMonitor {
	params := &chaincfg.MainNetParams
	if testnet {
		params = &chaincfg.TestNet3Params
	}
	return &BTCMonitor{
		Client:    client,
		DB:        db,
		NetParams: params,
	}
}

func (m *BTCMonitor) Start() {
	go func() {
		// Run immediately on start
		if err := m.ProcessBlocks(); err != nil {
			log.Printf("Error processing BTC blocks: %v", err)
		}

		ticker := time.NewTicker(60 * time.Second) // Check every 60s
		for range ticker.C {
			if err := m.ProcessBlocks(); err != nil {
				log.Printf("Error processing BTC blocks: %v", err)
			}
		}
	}()
}

func (m *BTCMonitor) ProcessBlocks() error {
	// 1. Get Chain Height
	chainHeight, err := m.Client.GetBlockCount()
	if err != nil {
		return err
	}

	// 2. Get Last Processed Height
	lastHeight := m.getLastHeight()
	if lastHeight == 0 {
		lastHeight = chainHeight - 1 // First run: Start from tip
	}

	// 3. Loop from Last+1 to ChainHeight (Catch up)
	for h := lastHeight + 1; h <= chainHeight; h++ {
		log.Printf("Processing BTC Block %d...", h)
		
		hash, err := m.Client.GetBlockHash(h)
		if err != nil {
			log.Printf("Failed to get hash for %d: %v", h, err)
			break // Stop processing to retry next tick
		}

		rawHex, err := m.Client.GetBlockRaw(hash)
		if err != nil {
			log.Printf("Failed to get block %s: %v", hash, err)
			break
		}

		blockBytes, err := hex.DecodeString(rawHex)
		if err != nil {
			log.Printf("Decode error: %v", err)
			break
		}

		// Prepare empty msgBlock to deserialize into
		// Use empty hashes, not nil, to avoid panic in NewBlockHeader
		emptyHash := &chainhash.Hash{}
		header := wire.NewBlockHeader(0, emptyHash, emptyHash, 0, 0)
		msgBlock := wire.NewMsgBlock(header)

		if err := msgBlock.Deserialize(bytes.NewReader(blockBytes)); err != nil {
			log.Printf("Deserialize error: %v", err)
			break
		}
		
		block := btcutil.NewBlock(msgBlock)

		for _, tx := range block.Transactions() {
			m.processTx(tx)
		}


		// 4. Update Checkpoint
		m.saveLastHeight(h)
	}

	return nil
}

func (m *BTCMonitor) getLastHeight() int64 {
	var val string
	err := m.DB.QueryRow("SELECT value FROM app_state WHERE key='btc_height'").Scan(&val)
	if err != nil {
		return 0
	}
	var h int64
	fmt.Sscanf(val, "%d", &h)
	return h
}

func (m *BTCMonitor) saveLastHeight(h int64) {
	_, err := m.DB.Exec("INSERT INTO app_state (key, value) VALUES ('btc_height', $1) ON CONFLICT (key) DO UPDATE SET value = $1", fmt.Sprintf("%d", h))
	if err != nil {
		log.Printf("Failed to save checkpoint: %v", err)
	}
}

func (m *BTCMonitor) processTx(tx *btcutil.Tx) {
	// Check Outputs for our addresses
	for _, txOut := range tx.MsgTx().TxOut {
		// Extract address from script
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, m.NetParams)
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			addrStr := addr.EncodeAddress()
			
			// Check DB if this address belongs to us
			// TODO: Optimise with cache
			if m.isDepositAddress(addrStr) {
				log.Printf("BTC DEPOSIT DETECTED: %s -> %s (%d sats)", tx.Hash(), addrStr, txOut.Value)
				// Insert into Transactions table
				m.recordTransaction(tx.Hash().String(), addrStr, txOut.Value)
			}
		}
	}
}

// Stub helpers for DB interaction to keep this file compile-able without full DB updates yet
func (m *BTCMonitor) isDepositAddress(addr string) bool {
	var id string
	err := m.DB.QueryRow("SELECT id FROM deposit_addresses WHERE address = $1 AND coin='BTC'", addr).Scan(&id)
	return err == nil
}

func (m *BTCMonitor) recordTransaction(txID, toAddress string, amount int64) {
	// Convert sats to BTC? Or store as Numeric/Sats. The DB schema uses Numeric.
	// Let's store as Numeric (Sats) for precision or Float. 
	// Best practice: Store as string or big int. PostgreSQL NUMERIC is good.
	
	_, err := m.DB.Exec(`
		INSERT INTO transactions (tx_id, coin, amount, to_address, status)
		VALUES ($1, 'BTC', $2, $3, 'detected')
		ON CONFLICT (id) DO NOTHING
		-- UUIDs, but mainly we want to avoid dups based on txid? 
		-- Logic needs unique constraint on tx_id+to_address?
	`, txID, amount, toAddress)
	if err != nil {
		log.Printf("Failed to record tx: %v", err)
		return
	}

	// Trigger Webhook
	// 1. Get Merchant Webhook URL
	var settingsJSON []byte
	err = m.DB.QueryRow(`
		SELECT m.settings 
		FROM merchants m 
		JOIN deposit_addresses da ON da.merchant_id = m.id 
		WHERE da.address = $1
	`, toAddress).Scan(&settingsJSON)

	url := ""
	if err == nil && settingsJSON != nil {
		var settings map[string]interface{}
		if json.Unmarshal(settingsJSON, &settings) == nil {
			if u, ok := settings["webhook_url"].(string); ok {
				url = u
			}
		}
	}

	// Trigger sending
	// Calculate BTC amount (Sats -> BTC)
	btcAmount := float64(amount) / 100000000.0

	// Fetch USD Price
	ps := price.NewPriceService()
	usdPrice, _ := ps.GetUSDPrice("BTC")
	usdAmount := btcAmount * usdPrice
	
	ts := time.Now().Format(time.RFC3339)

	worker.SendWebhook(url, worker.WebhookPayload{
		Event:     "deposit_detected",
		TxID:      txID,
		Coin:      "BTC",
		Address:   toAddress,
		UserID:    "", // TODO: Fetch UserID
		Amount:    btcAmount,
		AmountStr: fmt.Sprintf("%.8f BTC", btcAmount),
		USDPrice:  usdPrice,
		USDAmount: usdAmount,
		Status:    "detected",
		Timestamp: ts,
	})
}
