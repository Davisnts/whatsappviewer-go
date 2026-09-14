package crypto

import (
	"crypto/aes"
	"encoding/binary"
)

// ghash computes GHASH over blocks using hash key H
func ghash(H []byte, data []byte) []byte {
	// H is 16 bytes
	// data must be multiple of 16 bytes
	y := make([]byte, 16)
	for i := 0; i < len(data); i += 16 {
		for j := 0; j < 16; j++ {
			y[j] ^= data[i+j]
		}
		y = gfMul(y, H)
	}
	return y
}

// gfMul multiplies two elements in GF(2^128) using standard polynomial x^128 + x^7 + x^2 + x + 1 (0xE100000000000000)
func gfMul(x, y []byte) []byte {
	v := make([]byte, 16)
	copy(v, y)
	z := make([]byte, 16)

	for i := 0; i < 128; i++ {
		bit := (x[i/8] >> (7 - (i % 8))) & 1
		if bit != 0 {
			for j := 0; j < 16; j++ {
				z[j] ^= v[j]
			}
		}
		// v = v >> 1 (in GF(2^128) representation, LSB-first bit representation)
		lsb := v[15] & 1
		for j := 15; j > 0; j-- {
			v[j] = (v[j] >> 1) | ((v[j-1] & 1) << 7)
		}
		v[0] >>= 1
		if lsb != 0 {
			v[0] ^= 0xE1
		}
	}
	return z
}

// DeriveGCMInitialCounter computes the initial J0 counter block for GCM when IV length != 12 (specifically 16 bytes).
// J0 = GHASH_H(IV || 0^64 || [len(IV) in bits]_64)
// The first data block uses counter = inc32(J0).
func DeriveGCMInitialCounter(key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Compute H = AES_K(0^16)
	H := make([]byte, 16)
	block.Encrypt(H, H)

	var ghashInput []byte
	if len(iv) == 12 {
		j0 := make([]byte, 16)
		copy(j0, iv)
		j0[15] = 1
		return j0, nil
	}

	// For 16-byte IV:
	// Block 1: IV (16 bytes)
	// Block 2: 8 zero bytes followed by 64-bit length in bits (16 * 8 = 128 bits)
	ghashInput = make([]byte, 32)
	copy(ghashInput[:len(iv)], iv)
	binary.BigEndian.PutUint64(ghashInput[24:32], uint64(len(iv)*8))

	j0 := ghash(H, ghashInput)

	// Increment 32-bit counter for data encryption: inc32(J0)
	ctr := binary.BigEndian.Uint32(j0[12:16])
	binary.BigEndian.PutUint32(j0[12:16], ctr+1)

	return j0, nil
}
