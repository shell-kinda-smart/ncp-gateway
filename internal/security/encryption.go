package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// Encrypt encrypts plain text string into a hex encoded string using a 32-byte key
func Encrypt(stringToEncrypt string, keyString string) (string, error) {
	// Simple validation
	if len(keyString) < 32 {
		return "", fmt.Errorf("encryption key must be at least 32 chars")
	}
	// Use first 32 bytes of key
	key := []byte(keyString[:32])
	plaintext := []byte(stringToEncrypt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	
	// combine nonce + ciphertext
	finalBlob := make([]byte, len(nonce)+len(ciphertext))
	copy(finalBlob, nonce)
	copy(finalBlob[len(nonce):], ciphertext)

	return hex.EncodeToString(finalBlob), nil
}

// Decrypt decrypts hex encoded string into plain text string
func Decrypt(encryptedString string, keyString string) (string, error) {
	if len(keyString) < 32 {
		return "", fmt.Errorf("encryption key must be at least 32 chars")
	}
	key := []byte(keyString[:32])
	
	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	if len(enc) < 12 {
		return "", fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := enc[:12]
	ciphertext := enc[12:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
