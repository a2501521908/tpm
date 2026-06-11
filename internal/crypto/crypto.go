// Package crypto provides authenticated symmetric encryption helpers based on
// AES-256-GCM. Every ciphertext is self-contained: the 12-byte random nonce is
// prepended to the GCM output so that callers only need to persist a single blob.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// KeySize is the required length (in bytes) of the AES-256 key.
const KeySize = 32

// ErrInvalidKeySize is returned when the supplied key is not 32 bytes long.
var ErrInvalidKeySize = errors.New("crypto: key must be 32 bytes for AES-256")

// ErrCiphertextTooShort indicates that the blob does not even contain a nonce.
var ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")

// Encrypt seals plaintext with AES-256-GCM. The returned blob is
// nonce || ciphertext-with-tag and is safe to write directly to disk.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}

	// Seal appends the ciphertext to nonce, so the nonce ends up prefixed.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt opens a blob produced by Encrypt. A tampered ciphertext or wrong key
// surfaces as an error thanks to GCM's built-in authentication.
func Decrypt(blob, key []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(blob) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertext := blob[:nonceSize], blob[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt: %w", err)
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}
	return gcm, nil
}
