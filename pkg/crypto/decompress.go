package crypto

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
)

// DecompressGzipOrZlib decompresses a buffer that may be formatted with Gzip, Zlib, or raw Deflate.
func DecompressGzipOrZlib(compressed []byte) ([]byte, error) {
	if len(compressed) < 2 {
		return nil, fmt.Errorf("compressed data too short (%d bytes)", len(compressed))
	}

	// Check for Gzip magic bytes (0x1F, 0x8B)
	if compressed[0] == 0x1F && compressed[1] == 0x8B {
		r, err := gzip.NewReader(bytes.NewReader(compressed))
		if err == nil {
			defer r.Close()
			decompressed, err := io.ReadAll(r)
			if err == nil {
				return decompressed, nil
			}
		}
	}

	// Check for Zlib magic header (typically 0x78)
	if compressed[0] == 0x78 {
		r, err := zlib.NewReader(bytes.NewReader(compressed))
		if err == nil {
			defer r.Close()
			decompressed, err := io.ReadAll(r)
			if err == nil {
				return decompressed, nil
			}
		}
	}

	// Try Gzip as fallback
	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err == nil {
		defer gzReader.Close()
		decompressed, err := io.ReadAll(gzReader)
		if err == nil {
			return decompressed, nil
		}
	}

	// Try Zlib as fallback
	zlReader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err == nil {
		defer zlReader.Close()
		decompressed, err := io.ReadAll(zlReader)
		if err == nil {
			return decompressed, nil
		}
	}

	// Try raw Deflate
	flReader := flate.NewReader(bytes.NewReader(compressed))
	defer flReader.Close()
	decompressed, err := io.ReadAll(flReader)
	if err == nil {
		return decompressed, nil
	}

	return nil, fmt.Errorf("decompression failed (neither gzip, zlib, nor raw deflate succeeded - invalid key?)")
}
