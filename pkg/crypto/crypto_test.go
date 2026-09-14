package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCrypt5Key(t *testing.T) {
	key := BuildCrypt5Key("test@example.com")
	if len(key) != 24 {
		t.Fatalf("expected 24 bytes key, got %d", len(key))
	}
}

func TestValidateSQLite(t *testing.T) {
	valid := []byte("SQLite format 3\x00extra data")
	if err := ValidateSQLite(valid); err != nil {
		t.Errorf("expected valid sqlite, got error: %v", err)
	}

	invalid := []byte("Not a valid database")
	if err := ValidateSQLite(invalid); err == nil {
		t.Errorf("expected error for invalid database, got nil")
	}
}

func TestDecompressGzipOrZlib(t *testing.T) {
	original := []byte("SQLite format 3 - Test decompression")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(original)
	gw.Close()

	decompressed, err := DecompressGzipOrZlib(buf.Bytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(decompressed, original) {
		t.Errorf("expected %s, got %s", original, decompressed)
	}
}

func TestCrypt5RoundTrip(t *testing.T) {
	account := "user@gmail.com"
	key := BuildCrypt5Key(account)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "test.crypt5")
	outputPath := filepath.Join(tmpDir, "test.db")

	plaintext := []byte("SQLite format 3\x001234567890123456") // 32 bytes (multiple of 16)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	iv := make([]byte, 16)
	copy(iv, crypt5IV)

	ciphertext := make([]byte, len(plaintext))
	enc := cipher.NewCBCEncrypter(block, iv)
	enc.CryptBlocks(ciphertext, plaintext)

	if err := os.WriteFile(inputPath, ciphertext, 0644); err != nil {
		t.Fatal(err)
	}

	if err := DecryptCrypt5(inputPath, outputPath, account); err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	decrypted, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data mismatch")
	}
}
