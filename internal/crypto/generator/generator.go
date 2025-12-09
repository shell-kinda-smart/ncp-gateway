package generator

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

type GeneratedKeys struct {
	Mnemonic string
	BTCXPub  string // zpub (native)
	LTCXPub  string // xpub (legacy/standard)
	ETHXPub  string // xpub (standard BIP44)
}

// GenerateNewWallet creates a fresh 12-word seed and derives all necessary XPUBs.
// It returns the secrets (Mnemonic) and the public keys.
func GenerateNewWallet(testnet bool) (*GeneratedKeys, error) {
	// 1. Generate Entropy
	entropy, err := bip39.NewEntropy(128) // 128 bits = 12 words
	if err != nil {
		return nil, fmt.Errorf("entropy error: %v", err)
	}

	// 2. Encode to Mnemonic
	mnemonic, _ := bip39.NewMnemonic(entropy)

	// 3. Generate Seed from Mnemonic (No passphrase for simplicity)
	seed := bip39.NewSeed(mnemonic, "")

	// 4. Create Master Key
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}

	masterKey, err := hdkeychain.NewMaster(seed, netParams)
	if err != nil {
		return nil, fmt.Errorf("master key error: %v", err)
	}

	// 5. Derive Coins (BIP44/84)
	
	// BTC: BIP84 (Native SegWit) -> m/84'/0'/0'
	// Note: For Mainnet usage 0. For Testnet usage 1 is typical but CoinType is tricky.
	// We will stick to Mainnet paths or standard testnet coin type (1).
	btcCoinType := uint32(0)
	if testnet { btcCoinType = 1 }
	
	btcKey, err := deriveAccount(masterKey, 84, btcCoinType, 0)
	if err != nil { return nil, err }
	btcPub, _ := btcKey.Neuter() // extended public key

	// LTC: BIP44 (Legacy) -> m/44'/2'/0'
	ltcCoinType := uint32(2)
	ltcKey, err := deriveAccount(masterKey, 44, ltcCoinType, 0)
	if err != nil { return nil, err }
	ltcPub, _ := ltcKey.Neuter()

	// ETH: BIP44 (Standard) -> m/44'/60'/0'
	ethCoinType := uint32(60)
	ethKey, err := deriveAccount(masterKey, 44, ethCoinType, 0)
	if err != nil { return nil, err }
	ethPub, _ := ethKey.Neuter()

	return &GeneratedKeys{
		Mnemonic: mnemonic,
		BTCXPub:  btcPub.String(), 
		LTCXPub:  ltcPub.String(),
		ETHXPub:  ethPub.String(),
	}, nil
}

// deriveAccount helper: m / purpose' / coinType' / account'
func deriveAccount(master *hdkeychain.ExtendedKey, purpose, coinType, account uint32) (*hdkeychain.ExtendedKey, error) {
	// Purpose
	k, err := master.Derive(hdkeychain.HardenedKeyStart + purpose)
	if err != nil { return nil, err }
	
	// Coin
	k, err = k.Derive(hdkeychain.HardenedKeyStart + coinType)
	if err != nil { return nil, err }

	// Account
	k, err = k.Derive(hdkeychain.HardenedKeyStart + account)
	if err != nil { return nil, err }

	return k, nil
}
