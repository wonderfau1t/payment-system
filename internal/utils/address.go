package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateHexAddress() (string, error) {
	const fn = "internal.utils.generateHexAddress"

	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return "0x" + hex.EncodeToString(bytes), nil
}

func GenerateHexAddresses(count int) ([]string, error) {
	const fn = "internal.utils.generateHexAddresses"

	var hexAddresses []string
	for i := 0; i < count; i++ {
		address, err := GenerateHexAddress()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		hexAddresses = append(hexAddresses, address)
	}

	return hexAddresses, nil
}

func GenerateTransactionHash(sender, recipient string, amount float64, nonce int64) string {
	const fn = "internal.utils.generateTransactionHash"

	data := fmt.Sprintf("%s - %s - %f - %d", sender, recipient, amount, nonce)
	hash := sha256.Sum256([]byte(data))

	return "0x" + hex.EncodeToString(hash[:])
}
