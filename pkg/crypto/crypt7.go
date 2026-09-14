package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"os"
)

const skipBytesCrypt7 = 67

// DecryptCrypt7WithKeyFile decrypts a WhatsApp crypt7 database using a key file.
func DecryptCrypt7WithKeyFile(inputFile, outputFile, keyFilename string) error {
	key, _, err := LoadKey(keyFilename)
	if err != nil {
		return err
	}
	return DecryptCrypt7(inputFile, outputFile, key)
}

// DecryptCrypt7 decrypts a WhatsApp crypt7 database using the 32-byte AES key.
func DecryptCrypt7(inputFile, outputFile string, key []byte) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read crypt7 file: %w", err)
	}

	if len(data) < skipBytesCrypt7 {
		return fmt.Errorf("crypt7 file too small (%d bytes), minimum is %d", len(data), skipBytesCrypt7)
	}

	iv := data[51:67]
	ciphertext := data[skipBytesCrypt7:]

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return fmt.Errorf("crypt7 ciphertext length (%d) is not a multiple of AES block size (%d)", len(ciphertext), aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	ivCopy := make([]byte, 16)
	copy(ivCopy, iv)

	decrypter := cipher.NewCBCDecrypter(block, ivCopy)
	plaintext := make([]byte, len(ciphertext))
	decrypter.CryptBlocks(plaintext, ciphertext)

	if err := ValidateSQLite(plaintext); err != nil {
		return fmt.Errorf("crypt7 validation failed: %w", err)
	}

	if err := os.WriteFile(outputFile, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted database: %w", err)
	}

	return nil
}
