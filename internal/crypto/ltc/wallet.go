package ltc

import (
	"fmt"

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

	// Convert to Address (Witness PubKey Hash - ltc1...)
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
