package ltc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

type Wallet struct {
	NetParams *chaincfg.Params
}

func NewWallet(testnet bool) *Wallet {
	params := &LitecoinParams
	if testnet {
		// Just reuse mainnet params for now or define TestnetParams if strictly needed.
		// For MVP, user said testnet=false usually.
		// But let's keep it safe.
		p := LitecoinParams
		params = &p
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

	// Double check network match
	if !key.IsForNet(w.NetParams) {
		return "", fmt.Errorf("xpub network mismatch")
	}

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

	// Convert to Address
	pubKey, err := childKey.ECPubKey()
	if err != nil {
		return "", err
	}

	// Default to Native Segwit (Bech32) for standard calls if unspecified
	// But clearer to have explicit methods.
	return w.DeriveNativeSegWit(pubKey)
}

func (w *Wallet) DeriveNativeSegWit(pubKey *btcec.PublicKey) (string, error) {
	addr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubKey.SerializeCompressed()), w.NetParams)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
}

func (w *Wallet) DeriveLegacy(pubKey *btcec.PublicKey) (string, error) {
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubKey.SerializeCompressed()), w.NetParams)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
}

// Helper to get PubKey from XPUB+Index
func (w *Wallet) GetChildPubKey(xpubStr string, index uint32) (*btcec.PublicKey, error) {
	key, err := hdkeychain.NewKeyFromString(xpubStr)
	if err != nil {
		return nil, fmt.Errorf("invalid xpub: %v", err)
	}
	if !key.IsForNet(w.NetParams) {
		return nil, fmt.Errorf("xpub network mismatch")
	}

	extKey, err := key.Derive(0)
	if err != nil {
		return nil, err
	}
	childKey, err := extKey.Derive(index)
	if err != nil {
		return nil, err
	}
	
	return childKey.ECPubKey()
}

// Unified call
func (w *Wallet) DeriveAddressByType(xpubStr string, index uint32, addrType string) (string, error) {
	pk, err := w.GetChildPubKey(xpubStr, index)
	if err != nil { return "", err }
	
	if addrType == "legacy" {
		return w.DeriveLegacy(pk)
	}
	// Default to Native
	return w.DeriveNativeSegWit(pk)
}
