package main

import (
	"fmt"
	"log"

	"gateway/internal/crypto/ltc"
)

func main() {
	// User's LTC XPUB
	xpubStr := "xpub6DDn5pSVZKEuWmeJ8p84mpgVubu5bcu76EPp5Ue4irjBf7TDMr8zTkxtaFKL87f4sjK1MPArHgvRVeiwpXnX8GMtjthNRtMsaRXchTGZgc7"
	
	// Parse Key
	// key, err := hdkeychain.NewKeyFromString(xpubStr)
	// if err != nil {
	// 	log.Fatalf("Failed to parse XPUB: %v", err)
	// }

	// Use Internal LTC Wallet Package to Verify
	// This uses the EXACT same logic as the main app.
	wallet := ltc.NewWallet(false)
	
	addr, err := wallet.DeriveAddress(xpubStr, 0)
	if err != nil {
		log.Fatalf("Failed to derive address: %v", err)
	}

	fmt.Printf("XPUB: %s\n", xpubStr)
	fmt.Printf("Derived m/0/0 Address (Legacy P2PKH): %s\n", addr)
}
