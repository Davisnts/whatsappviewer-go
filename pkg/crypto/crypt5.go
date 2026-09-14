package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"fmt"
	"os"
)

var crypt5BaseKey = []byte{
	141, 75, 21, 92, 201, 255, 129, 229, 203, 246, 250, 120,
	25, 54, 106, 62, 198, 33, 166, 86, 65, 108, 215, 147,
}

var crypt5IV = []byte{
	0x1E, 0x39, 0xF3, 0x69, 0xE9, 0x0D, 0xB3, 0x3A,
	0xA7, 0x3B, 0x44, 0x2B, 0xBB, 0xB6, 0xB0, 0xB9,
}

// BuildCrypt5Key derives the 24-byte AES-192 key from the user account name.
func BuildCrypt5Key(accountName string) []byte {
	hash := md5.Sum([]byte(accountName))
	key := make([]byte, 24)
	for i := 0; i < 24; i++ {
		key[i] = crypt5BaseKey[i] ^ hash[i&0x0F]
	}
	return key
}

// DecryptCrypt5 decrypts a WhatsApp crypt5 database using the account name.
func DecryptCrypt5(inputFile, outputFile, accountName string) error {
	key := BuildCrypt5Key(accountName)
	return DecryptCrypt5WithKey(inputFile, outputFile, key)
}

// DecryptCrypt5WithKey decrypts a WhatsApp crypt5 database with a precomputed 24-byte key.
func DecryptCrypt5WithKey(inputFile, outputFile string, key []byte) error {
	ciphertext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read crypt5 input file: %w", err)
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return fmt.Errorf("crypt5 ciphertext size (%d) must be a non-zero multiple of AES block size (%d)", len(ciphertext), aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to initialize AES-192 cipher: %w", err)
	}

	iv := make([]byte, len(crypt5IV))
	copy(iv, crypt5IV)

	decrypter := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	decrypter.CryptBlocks(plaintext, ciphertext)

	if err := ValidateSQLite(plaintext); err != nil {
		return fmt.Errorf("crypt5 validation failed: %w", err)
	}

	if err := os.WriteFile(outputFile, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted database: %w", err)
	}

	return nil
}
