package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

func TestGCMDerivationMatchesStandardGCM(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	iv := make([]byte, 16)
	rand.Read(iv)

	plaintext := []byte("Hello, WhatsApp Crypt12 and Crypt14! Testing GCM counter derivation.")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("NewCipher failed: %v", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		t.Fatalf("NewGCMWithNonceSize failed: %v", err)
	}

	// Standard Go GCM produces ciphertext + 16-byte tag
	sealed := gcm.Seal(nil, iv, plaintext, nil)
	expectedCiphertext := sealed[:len(plaintext)]

	// Our CTR implementation
	initialCounter, err := DeriveGCMInitialCounter(key, iv)
	if err != nil {
		t.Fatalf("DeriveGCMInitialCounter failed: %v", err)
	}

	ctrStream := cipher.NewCTR(block, initialCounter)
	actualCiphertext := make([]byte, len(plaintext))
	ctrStream.XORKeyStream(actualCiphertext, plaintext)

	if !bytes.Equal(expectedCiphertext, actualCiphertext) {
		t.Fatalf("Derived CTR ciphertext does not match standard GCM keystream!\nExpected: %x\nActual:   %x", expectedCiphertext, actualCiphertext)
	}
}
