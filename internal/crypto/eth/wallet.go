package eth

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
)

// DeriveAddress derives an Ethereum address from an XPUB at a specific index.
// Path: xpub -> internal/external (0/1) -> index
// ETH usually uses just "m/0/i" from the xpub if the xpub is already at account level.
// We assume the Merchant provided an XPUB derived at m/44'/60'/0'.
// So we just derive m/0/index.
func DeriveAddress(xpubStr string, index uint32) (string, error) {
	// 1. Parse XPUB
	key, err := hdkeychain.NewKeyFromString(xpubStr)
	if err != nil {
		return "", fmt.Errorf("invalid xpub: %v", err)
	}

	// 2. Derive External Chain (0)
	// We assume xpub is Account level.
	// We want: Account -> Internal/External (0) -> Index
	
	// Safety check: is it a private key? We only want public.
	if key.IsPrivate() {
		return "", fmt.Errorf("provided key is private, expected public xpub")
	}

	// Derive 0 (External)
	externalChain, err := key.Derive(0)
	if err != nil {
		return "", err
	}

	// Derive Index
	childKey, err := externalChain.Derive(index)
	if err != nil {
		return "", err
	}

	// 3. Get Public Key
	pubKey, err := childKey.ECPubKey()
	if err != nil {
		return "", err
	}

	// 4. Convert to ETH Address
	// ETH Address = Keccak256(UncompressedPubKey[1:])[12:]
	// btcec PubKey -> ECDSA PubKey logic or just use crypto.PubkeyToAddress
	
	// Convert btcec.PublicKey to ecdsa.PublicKey
	ecdsaPubKey := pubKey.ToECDSA()
	
	address := crypto.PubkeyToAddress(*ecdsaPubKey)
	return address.Hex(), nil
}

// Helper to validate xpub format
func IsValidXPub(xpub string) bool {
	_, err := hdkeychain.NewKeyFromString(xpub)
	return err == nil
}
