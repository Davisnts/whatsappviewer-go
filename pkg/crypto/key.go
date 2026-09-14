package crypto

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// LoadKey loads the 32-byte AES key and 16-byte IV from a WhatsApp key file.
// The key file must be exactly 158 bytes.
// IV is located at offset 110 (16 bytes), and Key is located at offset 126 (32 bytes).
func LoadKey(filename string) (key []byte, iv []byte, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open key file: %w", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat key file: %w", err)
	}

	if fi.Size() != 158 {
		return nil, nil, fmt.Errorf("expected key filesize of 158 bytes, got %d", fi.Size())
	}

	if _, err := file.Seek(110, io.SeekStart); err != nil {
		return nil, nil, fmt.Errorf("failed to seek to IV offset: %w", err)
	}

	iv = make([]byte, 16)
	if _, err := io.ReadFull(file, iv); err != nil {
		return nil, nil, fmt.Errorf("failed to read IV: %w", err)
	}

	key = make([]byte, 32)
	if _, err := io.ReadFull(file, key); err != nil {
		return nil, nil, fmt.Errorf("failed to read key: %w", err)
	}

	return key, iv, nil
}

// ValidateSQLite validates that the decrypted stream starts with the standard SQLite header.
func ValidateSQLite(data []byte) error {
	const magic = "SQLite format 3"
	if len(data) < len(magic) {
		return errors.New("data too short to validate SQLite header")
	}
	if string(data[:len(magic)]) != magic {
		return errors.New("validation failed: SQLite format 3 magic header not found (invalid key or corrupt data)")
	}
	return nil
}
