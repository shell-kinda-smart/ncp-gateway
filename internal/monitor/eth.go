package monitor

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"encoding/json"

	"gateway/internal/crypto/eth"
	"gateway/internal/worker"
)

type ETHMonitor struct {
	Client *eth.Client
	DB     *sql.DB
}

func NewETHMonitor(client *eth.Client, db *sql.DB) *ETHMonitor {
	return &ETHMonitor{
		Client: client,
		DB:     db,
	}
}

func (m *ETHMonitor) Start() {
	go func() {
		// Run immediately
		if err := m.ProcessBlocks(); err != nil {
			log.Printf("Error processing ETH blocks: %v", err)
		}

		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if err := m.ProcessBlocks(); err != nil {
				log.Printf("Error processing ETH blocks: %v", err)
			}
		}
	}()
}

func (m *ETHMonitor) ProcessBlocks() error {
	chainHeight, err := m.Client.GetBlockNumber()
	if err != nil {
		return err
	}

	lastHeight := m.getLastHeight()
	if lastHeight == 0 {
		lastHeight = chainHeight - 1
	}

	for h := lastHeight + 1; h <= chainHeight; h++ {
		log.Printf("Processing ETH Block %d...", h)

		block, err := m.Client.GetBlockByNumber(h)
		if err != nil {
			log.Printf("Failed to get ETH block %d: %v", h, err)
			break
		}

		for _, tx := range block.Transactions {
			m.processTx(tx)
		}

		m.saveLastHeight(h)
	}

	return nil
}

func (m *ETHMonitor) getLastHeight() int64 {
	var val string
	err := m.DB.QueryRow("SELECT value FROM app_state WHERE key='eth_height'").Scan(&val)
	if err != nil {
		return 0
	}
	var h int64
	fmt.Sscanf(val, "%d", &h)
	return h
}

func (m *ETHMonitor) saveLastHeight(h int64) {
	_, err := m.DB.Exec("INSERT INTO app_state (key, value) VALUES ('eth_height', $1) ON CONFLICT (key) DO UPDATE SET value = $1", fmt.Sprintf("%d", h))
	if err != nil {
		log.Printf("Failed to save eth checkpoint: %v", err)
	}
}

func (m *ETHMonitor) processTx(tx eth.Transaction) {
	if tx.To == "" {
		return // Contract creation
	}

	// Normalize to lowercase for comparison
	toAddr := strings.ToLower(tx.To)

	if m.isDepositAddress(toAddr) {
		valBig := eth.HexToBigInt(tx.Value)
		// Convert Wei to ETH (float) for simpler DB storage? 
		// Or keep as Wei (Numeric). Let's store as Numeric (Wei) to be precise.
		// DB "amount" is NUMERIC.
		
		log.Printf("ETH DEPOSIT DETECTED: %s -> %s (%s Wei)", tx.Hash, toAddr, valBig.String())
		m.recordTransaction(tx.Hash, toAddr, valBig.String())
	}
}

func (m *ETHMonitor) isDepositAddress(addr string) bool {
	var id string
	// DB addresses should be stored lowercase or we use ILIKE? 
	// Let's assume we store them lowercase in DB for ETH.
	err := m.DB.QueryRow("SELECT id FROM deposit_addresses WHERE lower(address) = $1 AND coin='ETH'", addr).Scan(&id)
	return err == nil
}

func (m *ETHMonitor) recordTransaction(txID, toAddress string, amount string) {
	_, err := m.DB.Exec(`
		INSERT INTO transactions (tx_id, coin, amount, to_address, status)
		VALUES ($1, 'ETH', $2, $3, 'detected')
		ON CONFLICT (id) DO NOTHING
	`, txID, amount, toAddress)
	if err != nil {
		log.Printf("Failed to record eth tx: %v", err)
		return
	}

	// Trigger Webhook
	var settingsJSON []byte
	// ETH addresses are case insensitive in DB lookup usually, but we stored as is. 
	// We matched with lower(address) in isDepositAddress. We should match same way here.
	err = m.DB.QueryRow(`
		SELECT m.settings 
		FROM merchants m 
		JOIN deposit_addresses da ON da.merchant_id = m.id 
		WHERE lower(da.address) = $1
	`, strings.ToLower(toAddress)).Scan(&settingsJSON)

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
		Coin:      "ETH",
		Amount:    amount + " wei",
		ToAddress: toAddress,
		Status:    "detected",
	})
}
