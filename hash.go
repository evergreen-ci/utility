package utility

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
)

// Hash is a wrapper around a hashing algorithm.
type Hash struct {
	hash.Hash
}

// NewMD5Hash returns a Hash that uses the MD5 algorithm.
func NewMD5Hash() Hash {
	return Hash{Hash: md5.New()}
}

// NewSHA1Hash returns a Hash that uses the SHA1 algorithm.
func NewSHA1Hash() Hash {
	return Hash{Hash: sha1.New()}
}

// NewSHA256Hash returns a Hash that uses the SHA256 algorithm.
func NewSHA256Hash() Hash {
	return Hash{Hash: sha256.New()}
}

// Add adds data to the hasher.
func (h Hash) Add(data string) {
	// The hash.Hash interface says the io.Writer will never return an error, so
	// the returned error can be squashed.
	_, _ = io.WriteString(h, data)
}

// Sum returns the hash sum of the accumulated data.
func (h Hash) Sum() string {
	return fmt.Sprintf("%x", h.Hash.Sum(nil))
}
