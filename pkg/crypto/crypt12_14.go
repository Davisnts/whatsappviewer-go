package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"os"
)

// DecryptCrypt12WithKeyFile decrypts a WhatsApp crypt12 database using a key file.
func DecryptCrypt12WithKeyFile(inputFile, outputFile, keyFilename string) error {
	key, _, err := LoadKey(keyFilename)
	if err != nil {
		return err
	}
	return DecryptCrypt12(inputFile, outputFile, key)
}

// DecryptCrypt12 decrypts a WhatsApp crypt12 database using a 32-byte key.
// IV offset: 51, Data offset: 67, Footer size: 20.
func DecryptCrypt12(inputFile, outputFile string, key []byte) error {
	return decryptCrypt12_14(inputFile, outputFile, key, 51, 67, 20)
}

// DecryptCrypt14WithKeyFile decrypts a WhatsApp crypt14 database using a key file.
func DecryptCrypt14WithKeyFile(inputFile, outputFile, keyFilename string) error {
	key, _, err := LoadKey(keyFilename)
	if err != nil {
		return err
	}
	return DecryptCrypt14(inputFile, outputFile, key)
}

// DecryptCrypt14 decrypts a WhatsApp crypt14 database using a 32-byte key.
// IV offset: 67, Data offset: 191, Footer size: 0.
func DecryptCrypt14(inputFile, outputFile string, key []byte) error {
	return decryptCrypt12_14(inputFile, outputFile, key, 67, 191, 0)
}

// decryptCrypt12_14 handles the shared AES-GCM stream decryption + Gzip/Zlib decompression
// for crypt12 and crypt14 formats.
func decryptCrypt12_14(inputFile, outputFile string, key []byte, ivOffset, dataOffset, footerSize int) error {
	fileData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	filesize := len(fileData)
	if filesize < dataOffset+footerSize {
		return fmt.Errorf("input file is too small (%d bytes), minimum expected %d bytes", filesize, dataOffset+footerSize)
	}

	if filesize < ivOffset+16 {
		return fmt.Errorf("input file is too small to read IV at offset %d", ivOffset)
	}

	iv := fileData[ivOffset : ivOffset+16]
	ciphertext := fileData[dataOffset : filesize-footerSize]

	if len(ciphertext) == 0 {
		return fmt.Errorf("ciphertext is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	ctrCounter, err := DeriveGCMInitialCounter(key, iv)
	if err != nil {
		return fmt.Errorf("failed to derive GCM initial counter: %w", err)
	}

	ctr := cipher.NewCTR(block, ctrCounter)
	decryptedCompressed := make([]byte, len(ciphertext))
	ctr.XORKeyStream(decryptedCompressed, ciphertext)

	uncompressed, err := DecompressGzipOrZlib(decryptedCompressed)
	if err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	if err := ValidateSQLite(uncompressed); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := os.WriteFile(outputFile, uncompressed, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted output file: %w", err)
	}

	return nil
}
