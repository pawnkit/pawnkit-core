package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

const sha256Prefix = "sha256:"

// Content returns a canonical, versioned hash of data, suitable for use in
// a cache key.
func Content(data []byte) string {
	sum := sha256.Sum256(data)

	return sha256Prefix + hex.EncodeToString(sum[:])
}

// String returns the hash of s.
func String(s string) string {
	return Content([]byte(s))
}

// Reader hashes all bytes read from r.
func Reader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hash: reading input: %w", err)
	}

	return sha256Prefix + hex.EncodeToString(h.Sum(nil)), nil
}

// JSON hashes v's JSON encoding. Equal values always produce the same
// encoding (struct fields in declaration order, sorted map keys), so the
// hash is stable across runs and processes.
func JSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("hash: marshaling value: %w", err)
	}

	return Content(data), nil
}

// Combine deterministically combines one or more hash strings (or other
// short, stable identifiers) into a single hash. The result depends on the
// order of parts.
func Combine(parts ...string) string {
	h := sha256.New()

	for _, p := range parts {
		// NUL separator prevents ambiguous concatenation, e.g. ("ab","c")
		// vs. ("a","bc").
		h.Write([]byte(p))
		h.Write([]byte{0})
	}

	return sha256Prefix + hex.EncodeToString(h.Sum(nil))
}
