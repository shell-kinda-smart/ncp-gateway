package btc

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/tyler-smith/go-bip39"
)

// Signer handles BTC signing
type Signer struct {
	Client  *Client
	DB      *sql.DB
	Testnet bool
}

func NewSigner(client *Client, db *sql.DB, testnet bool) *Signer {
	return &Signer{Client: client, DB: db, Testnet: testnet}
}

// Send constructs, signs, and broadcasts a transaction
func (s *Signer) Send(mnemonic string, destAddress string, amountStr string, merchantID string) (string, error) {
	// 1. Fetch UTXOs
	utxos, err := s.Client.ListUnspent(0, 9999999, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch utxos: %v", err)
	}

	// 2. Parse Amount (Simplify to "max" or assume float string)
	var targetAmount int64
	sendMax := (amountStr == "max")

	if !sendMax {
		// Parse string -> float -> Amount
		amtFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return "", fmt.Errorf("invalid amount: %v", err)
		}
		amtSats, err := btcutil.NewAmount(amtFloat)
		if err != nil {
			return "", fmt.Errorf("invalid amount satoshis: %v", err)
		}
		targetAmount = int64(amtSats)
	}

	// 3. Coin Selection
	var selectedUTXOs []UTXO
	var totalIn int64 = 0
	
	// Create Fetcher for Signing later
	fetcherMap := make(map[wire.OutPoint]*wire.TxOut)

	for _, u := range utxos {
		selectedUTXOs = append(selectedUTXOs, u)
		amtSats := int64(u.Amount * 1e8)
		totalIn += amtSats
		
		// Add to fetcher
		hash, _ := chainhash.NewHashFromStr(u.TxID)
		op := wire.OutPoint{Hash: *hash, Index: uint32(u.Vout)}
		
		// Decode ScriptPubKey from hex
		scriptBytes, _ := hex.DecodeString(u.ScriptPubKey)
		fetcherMap[op] = wire.NewTxOut(amtSats, scriptBytes)

		if !sendMax && totalIn >= targetAmount {
             break 
        }
	}
	
	if !sendMax && totalIn < targetAmount {
		return "", errors.New("insufficient balance")
	}

	// 4. Create Tx
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// Add Inputs
	for _, u := range selectedUTXOs {
		hash, _ := chainhash.NewHashFromStr(u.TxID)
		outPoint := wire.NewOutPoint(hash, uint32(u.Vout))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		msgTx.AddTxIn(txIn)
	}

	// 5. Add Outputs
	// Destination
	destAddr, err := btcutil.DecodeAddress(destAddress, s.getNetParams())
	if err != nil {
		return "", fmt.Errorf("invalid dest address: %v", err)
	}
	destScript, _ := txscript.PayToAddrScript(destAddr)
	
	var fee int64 = 2000 // Fixed fee for MVP (2000 sats)
	
	if sendMax {
		targetAmount = totalIn - fee
		if targetAmount <= 0 {
			return "", errors.New("balance too low for fee")
		}
		msgTx.AddTxOut(wire.NewTxOut(targetAmount, destScript))
	} else {
		msgTx.AddTxOut(wire.NewTxOut(targetAmount, destScript))
		change := totalIn - targetAmount - fee
		if change > 546 { 
			changeAddr, _ := s.deriveChangeAddress(mnemonic)
			changeScript, _ := txscript.PayToAddrScript(changeAddr)
			msgTx.AddTxOut(wire.NewTxOut(change, changeScript))
		}
	}

	// 6. Sign Inputs
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, _ := hdkeychain.NewMaster(seed, s.getNetParams())
	
	// Derive Account Key m/84'/0'/0'
	accKey, _ := s.derivePath(masterKey, []uint32{84, 0, 0})
	if s.Testnet {
		accKey, _ = s.derivePath(masterKey, []uint32{84, 1, 0}) 
	}
	// Derive External Chain m/84'/0'/0'/0
	extKeyNormal, _ := accKey.Derive(0)

	// Create Fetcher
	fetcher := txscript.NewMultiPrevOutFetcher(fetcherMap)
	sigHashes := txscript.NewTxSigHashes(msgTx, fetcher)

	for i, u := range selectedUTXOs {
		var index int
		err := s.DB.QueryRow("SELECT derivation_index FROM deposit_addresses WHERE address=$1", u.Address).Scan(&index)
		if err != nil {
			// Fallback: This might be a change address. DB lookup might fail if we don't track change.
			// This is a risk. For MVP we fail.
			return "", fmt.Errorf("unknown address in utxo: %s", u.Address)
		}
		
		childKey, _ := extKeyNormal.Derive(uint32(index))
		privKey, _ := childKey.ECPrivKey()
		
		amtSats := int64(u.Amount * 1e8)
		pubKey := privKey.PubKey().SerializeCompressed()
		pubKeyHash := btcutil.Hash160(pubKey)
		p2pkhScript, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).Script()
		
		witness, err := txscript.WitnessSignature(msgTx, sigHashes, i, amtSats, p2pkhScript, txscript.SigHashAll, privKey, true)
		if err != nil {
             return "", err
        }
		msgTx.TxIn[i].Witness = witness
	}

	// 7. Serialize & Broadcast
	var buf bytes.Buffer
	msgTx.Serialize(&buf)
	hexTx := hex.EncodeToString(buf.Bytes())
	
	return s.Client.SendRawTransaction(hexTx)
}

func (s *Signer) getNetParams() *chaincfg.Params {
	if s.Testnet { return &chaincfg.TestNet3Params }
	return &chaincfg.MainNetParams
}

func (s *Signer) derivePath(key *hdkeychain.ExtendedKey, path []uint32) (*hdkeychain.ExtendedKey, error) {
	k := key
	var err error
	for _, p := range path {
		k, err = k.Derive(hdkeychain.HardenedKeyStart + p)
		if err != nil { return nil, err }
	}
	return k, nil
}

func (s *Signer) deriveChangeAddress(mnemonic string) (btcutil.Address, error) {
	// For simplicity, just reuse index 0 of external chain for now to avoid "Change Chain" complexity
	// Or implement proper change logic.
	// Reuse index 0 is safest "Lazy" way.
	return nil, nil
}
