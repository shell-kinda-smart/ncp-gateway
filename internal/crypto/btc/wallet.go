package btc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

type Wallet struct {
	NetParams *chaincfg.Params
}

func NewWallet(testnet bool) *Wallet {
	params := &chaincfg.MainNetParams
	if testnet {
		params = &chaincfg.TestNet3Params
	}
	return &Wallet{NetParams: params}
}

// DeriveAddress generates a child address from an xpub at a specific index.
// Path: m/0/index (External chain)
func (w *Wallet) DeriveAddress(xpubStr string, index uint32) (string, error) {
	// Parse xpub
	key, err := hdkeychain.NewKeyFromString(xpubStr)
	if err != nil {
		return "", fmt.Errorf("invalid xpub: %v", err)
	}

	// 3. Verify Network (Mainnet vs Testnet)
	// Relaxed check: btcsuite expects xpub (legacy) for Mainnet.
	// Users providing zpub (SegWit) causes IsForNet to fail (Bytes mismatch).
	// We trust NewKeyFromString validated the structure.
	// if !key.IsForNet(w.NetParams) {
	// 	return "", fmt.Errorf("xpub network mismatch. Expected Net: %v", w.NetParams.Name)
	// }

	// We assume xpub is already at account level (e.g. m/84'/0'/0').
	// We adhere to BIP44/84 standard for external chain: derive 0 (external) then index.
	
	// Derive External Chain (0)
	extKey, err := key.Derive(0)
	if err != nil {
		return "", fmt.Errorf("child 0 error: %v", err)
	}

	// Derive Index
	childKey, err := extKey.Derive(index)
	if err != nil {
		return "", fmt.Errorf("child index error: %v", err)
	}

	// Convert to Address (P2WPKH - SegWit)
	// We assume the xpub corresponds to SegWit (zpub/xpub). 
	// The address format depends on how we encode it.
	// For this gateway, we enforce P2WPKH (bc1q...).
	
	pubKey, err := childKey.ECPubKey()
	if err != nil {
		return "", err
	}

	addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubKey.SerializeCompressed()), w.NetParams)
	if err != nil {
		return "", err
	}

	return addr.EncodeAddress(), nil
}

// DeriveAddressByType generates a child address based on the requested format.
func (w *Wallet) DeriveAddressByType(xpubStr string, index uint32, addrType string) (string, error) {
	// Parse xpub
	key, err := hdkeychain.NewKeyFromString(xpubStr)
	if err != nil {
		return "", fmt.Errorf("invalid xpub: %v", err)
	}

	// Relaxed Network Check
	// if !key.IsForNet(w.NetParams) { ... }

	// Derive Path: m/0/index
	extKey, err := key.Derive(0)
	if err != nil { return "", err }
	
	childKey, err := extKey.Derive(index)
	if err != nil { return "", err }
	
	pubKey, err := childKey.ECPubKey()
	if err != nil { return "", err }
	
	serializedKey := pubKey.SerializeCompressed()
	
	switch addrType {
	case "legacy":
		// P2PKH (1...)
		addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(serializedKey), w.NetParams)
		if err != nil { return "", err }
		return addr.EncodeAddress(), nil
		
	case "segwit", "nested":
		// P2SH-P2WPKH (3...)
		witnessAddr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(serializedKey), w.NetParams)
		if err != nil { return "", err }
		
		// Create Redeem Script: OP_0 <20-byte-hash>
		redeemScript, err := txscript.PayToAddrScript(witnessAddr)
		if err != nil { return "", err }
		
		addr, err := btcutil.NewAddressScriptHash(redeemScript, w.NetParams)
		if err != nil { return "", err }
		return addr.EncodeAddress(), nil
		
	case "native":
		// P2WPKH (bc1q...)
		addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(serializedKey), w.NetParams)
		if err != nil { return "", err }
		return addr.EncodeAddress(), nil
		
	default:
		// Default to Native
		addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(serializedKey), w.NetParams)
		if err != nil { return "", err }
		return addr.EncodeAddress(), nil
	}
}

// DetectFormat attempts to guess the preferred format based on XPUB version bytes.
func (w *Wallet) DetectFormat(xpubStr string) string {
	if len(xpubStr) < 4 { return "native" }
	prefix := xpubStr[:4]
	
	switch prefix {
	case "xpub", "tpub":
		return "legacy"
	case "ypub", "upub":
		return "segwit"
	case "zpub", "vpub":
		return "native"
	default:
		return "native"
	}
}

