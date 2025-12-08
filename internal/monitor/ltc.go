package monitor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"gateway/internal/crypto/ltc"
	"gateway/internal/worker"
)

type LTCMonitor struct {
	Client    *ltc.Client
	DB        *sql.DB
	NetParams *chaincfg.Params
}

func NewLTCMonitor(client *ltc.Client, db *sql.DB, testnet bool) *LTCMonitor {
	params := &ltc.LitecoinParams
	if testnet {
		// Use mainnet params for now or define testnet
		p := ltc.LitecoinParams
		params = &p
	}
	return &LTCMonitor{
		Client:    client,
		DB:        db,
		NetParams: params,
	}
}

func (m *LTCMonitor) Start() {
	go func() {
		// Run immediately
		if err := m.ProcessBlocks(); err != nil {
			log.Printf("Error processing LTC blocks: %v", err)
		}

		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if err := m.ProcessBlocks(); err != nil {
				log.Printf("Error processing LTC blocks: %v", err)
			}
		}
	}()
}

func (m *LTCMonitor) ProcessBlocks() error {
	chainHeight, err := m.Client.GetBlockCount()
	if err != nil {
		return err
	}

	lastHeight := m.getLastHeight()
	if lastHeight == 0 {
		lastHeight = chainHeight - 1
	}

	for h := lastHeight + 1; h <= chainHeight; h++ {
		log.Printf("Processing LTC Block %d...", h)

		hash, err := m.Client.GetBlockHash(h)
		if err != nil {
			log.Printf("Failed to get hash %d: %v", h, err)
			break
		}

		// Use JSON Verbose instead of Raw Byte parsing to avoid decoding issues (MWEB/SegWit flags)
		block, err := m.Client.GetBlockVerbose(hash)
		if err != nil {
			log.Printf("Failed to get block verbose %s: %v", hash, err)
			break
		}

		for _, tx := range block.Tx {
			m.processTx(tx)
		}

		m.saveLastHeight(h)
	}

	return nil
}

func (m *LTCMonitor) getLastHeight() int64 {
	var val string
	err := m.DB.QueryRow("SELECT value FROM app_state WHERE key='ltc_height'").Scan(&val)
	if err != nil {
		return 0
	}
	var h int64
	fmt.Sscanf(val, "%d", &h)
	return h
}

func (m *LTCMonitor) saveLastHeight(h int64) {
	_, err := m.DB.Exec("INSERT INTO app_state (key, value) VALUES ('ltc_height', $1) ON CONFLICT (key) DO UPDATE SET value = $1", fmt.Sprintf("%d", h))
	if err != nil {
		log.Printf("Failed to save ltc checkpoint: %v", err)
	}
}

func (m *LTCMonitor) processTx(tx ltc.TxJSON) {
	for _, vout := range tx.Vout {
		// Amount is float64 in JSON (LTC), convert to Lits (int64)
		// 1 LTC = 100,000,000 Lits
		amount := int64(vout.Value * 100000000)

		for _, addrStr := range vout.ScriptPubKey.Addresses {
			if m.isDepositAddress(addrStr) {
				log.Printf("LTC DEPOSIT DETECTED: %s -> %s (%d lits)", tx.TxID, addrStr, amount)
				m.recordTransaction(tx.TxID, addrStr, amount)
			}
		}
	}
}

func (m *LTCMonitor) isDepositAddress(addr string) bool {
	var id string
	err := m.DB.QueryRow("SELECT id FROM deposit_addresses WHERE address = $1 AND coin='LTC'", addr).Scan(&id)
	return err == nil
}

func (m *LTCMonitor) recordTransaction(txID, toAddress string, amount int64) {
	_, err := m.DB.Exec(`
		INSERT INTO transactions (tx_id, coin, amount, to_address, status)
		VALUES ($1, 'LTC', $2, $3, 'detected')
		ON CONFLICT (id) DO NOTHING
	`, txID, amount, toAddress)
	if err != nil {
		log.Printf("Failed to record LTC tx: %v", err)
		return
	}

	// Trigger Webhook
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

	worker.SendWebhook(url, worker.WebhookPayload{
		TxID:      txID,
		Coin:      "LTC",
		Amount:    fmt.Sprintf("%d lits", amount),
		ToAddress: toAddress,
		Status:    "detected",
	})
}
