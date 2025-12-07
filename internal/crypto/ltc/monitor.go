package ltc

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/ltcsuite/ltcd/txscript"
	"github.com/ltcsuite/ltcd/wire"
)

type Monitor struct {
	Client    *Client
	DB        *sql.DB
	NetParams *chaincfg.Params
}

func NewMonitor(client *Client, db *sql.DB, testnet bool) *Monitor {
	params := &chaincfg.MainNetParams
	if testnet {
		params = &chaincfg.TestNet4Params
	}
	return &Monitor{
		Client:    client,
		DB:        db,
		NetParams: params,
	}
}

func (m *Monitor) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if err := m.ProcessBlocks(); err != nil {
				log.Printf("Error processing LTC blocks: %v", err)
			}
		}
	}()
}

func (m *Monitor) ProcessBlocks() error {
	height, err := m.Client.GetBlockCount()
	if err != nil {
		return err
	}

	log.Printf("Current LTC Height: %d", height)

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
		// LTC RPC might return non-hex? No, standard is hex.
		return err
	}

	// LTC Wire Parsing
	// Note: LTC uses same wire structure as BTC usually
	msgBlock := wire.NewMsgBlock(wire.NewBlockHeader(0, nil, nil, 0, 0))
	if err := msgBlock.Deserialize(bytes.NewReader(blockBytes)); err != nil {
		return err
	}
	
	block := ltcutil.NewBlock(msgBlock)

	for _, tx := range block.Transactions() {
		m.processTx(tx)
	}

	return nil
}

func (m *Monitor) processTx(tx *ltcutil.Tx) {
	for _, txOut := range tx.MsgTx().TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, m.NetParams)
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			addrStr := addr.EncodeAddress()
			if m.isDepositAddress(addrStr) {
				log.Printf("LTC DEPOSIT DETECTED: %s -> %s (%d lits)", tx.Hash(), addrStr, txOut.Value)
				m.recordTransaction(tx.Hash().String(), addrStr, txOut.Value)
			}
		}
	}
}

func (m *Monitor) isDepositAddress(addr string) bool {
	var id string
	err := m.DB.QueryRow("SELECT id FROM deposit_addresses WHERE address = $1 AND coin='LTC'", addr).Scan(&id)
	return err == nil
}

func (m *Monitor) recordTransaction(txID, toAddress string, amount int64) {
	_, err := m.DB.Exec(`
		INSERT INTO transactions (tx_id, coin, amount, to_address, status)
		VALUES ($1, 'LTC', $2, $3, 'detected')
		ON CONFLICT (id) DO NOTHING
	`, txID, amount, toAddress)
	if err != nil {
		log.Printf("Failed to record LTC tx: %v", err)
	}
}
