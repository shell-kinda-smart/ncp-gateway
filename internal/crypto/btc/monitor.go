package btc

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type Monitor struct {
	Client    *Client
	DB        *sql.DB
	NetParams *chaincfg.Params
}

func NewMonitor(client *Client, db *sql.DB, testnet bool) *Monitor {
	params := &chaincfg.MainNetParams
	if testnet {
		params = &chaincfg.TestNet3Params
	}
	return &Monitor{
		Client:    client,
		DB:        db,
		NetParams: params,
	}
}

func (m *Monitor) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30s
		for range ticker.C {
			if err := m.ProcessBlocks(); err != nil {
				log.Printf("Error processing BTC blocks: %v", err)
			}
		}
	}()
}

func (m *Monitor) ProcessBlocks() error {
	// 1. Get current chain height
	height, err := m.Client.GetBlockCount()
	if err != nil {
		return err
	}

	// 2. Get last processed height from DB (We need a table for this, or store in memory/redis. 
	// For now, let's assume we start from 'height - 1' or use a KV store.
	// To keep it simple for this MVP, we will just scan the tip, 
	// but a PROD system needs a 'params' table to store 'last_btc_block_height'.
	// TODO: Add 'system_state' table.
	// For now, let's just log the height.
	log.Printf("Current BTC Height: %d", height)

	// 3. Fetch Block
	hash, err := m.Client.GetBlockHash(height)
	if err != nil {
		return err
	}

	rawHex, err := m.Client.GetBlockRaw(hash)
	if err != nil {
		return err
	}

	blockBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		return err
	}

	msgBlock := wire.NewMsgBlock(wire.NewBlockHeader(0, nil, nil, 0, 0))
	if err := msgBlock.Deserialize(bytes.NewReader(blockBytes)); err != nil {
		return err
	}
	
	block := btcutil.NewBlock(msgBlock)

	// 4. Iterate Transactions
	for _, tx := range block.Transactions() {
		m.processTx(tx)
	}

	return nil
}

func (m *Monitor) processTx(tx *btcutil.Tx) {
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
func (m *Monitor) isDepositAddress(addr string) bool {
	var id string
	err := m.DB.QueryRow("SELECT id FROM deposit_addresses WHERE address = $1 AND coin='BTC'", addr).Scan(&id)
	return err == nil
}

func (m *Monitor) recordTransaction(txID, toAddress string, amount int64) {
	// Convert sats to BTC? Or store as Numeric/Sats. The DB schema uses Numeric.
	// Let's store as Numeric (Sats) for precision or Float. 
	// Best practice: Store as string or big int. PostgreSQL NUMERIC is good.
	
	_, err := m.DB.Exec(`
		INSERT INTO transactions (tx_id, coin, amount, to_address, status)
		VALUES ($1, 'BTC', $2, $3, 'detected')
		ON CONFLICT (id) DO NOTHING -- UUIDs, but mainly we want to avoid dups based on txid? 
		-- Logic needs unique constraint on tx_id+to_address?
	`, txID, amount, toAddress)
	if err != nil {
		log.Printf("Failed to record tx: %v", err)
	}
}
