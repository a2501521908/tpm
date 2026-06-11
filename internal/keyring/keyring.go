// Package keyring manages the 32-byte AES master key inside the operating
// system's native secret store (macOS Keychain, Windows Credential Manager,
// Linux Secret Service) via go-keyring. The master key is never written to the
// sync directory; it only ever lives in the OS keyring and in process memory.
package keyring

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	gokeyring "github.com/zalando/go-keyring"

	"github.com/zhangshuaike/tpm/internal/crypto"
)

const (
	// service is the keyring service identifier used for all TPM secrets.
	service = "tpm"
	// account is the keyring account name that holds the master key.
	account = "master-key"
)

// ErrNotFound is returned when no master key exists in the keyring yet.
var ErrNotFound = errors.New("keyring: master key not found")

// Exists reports whether a master key is already stored in the keyring.
func Exists() (bool, error) {
	_, err := gokeyring.Get(service, account)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gokeyring.ErrNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("keyring: lookup: %w", err)
}

// Generate creates a new random 32-byte master key, stores it in the keyring,
// and returns the raw key bytes.
func Generate() ([]byte, error) {
	key := make([]byte, crypto.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("keyring: generate key: %w", err)
	}
	if err := store(key); err != nil {
		return nil, err
	}
	return key, nil
}

// Import stores an externally supplied master key (base64-encoded, as exported
// from a primary device) into the local keyring. It validates length to catch
// paste mistakes early.
func Import(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("keyring: decode imported key: %w", err)
	}
	if len(key) != crypto.KeySize {
		return nil, fmt.Errorf("keyring: imported key must be %d bytes, got %d", crypto.KeySize, len(key))
	}
	if err := store(key); err != nil {
		return nil, err
	}
	return key, nil
}

// Get returns the raw master key bytes from the keyring.
func Get() ([]byte, error) {
	encoded, err := gokeyring.Get(service, account)
	if err != nil {
		if errors.Is(err, gokeyring.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("keyring: get: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("keyring: decode stored key: %w", err)
	}
	if len(key) != crypto.KeySize {
		return nil, fmt.Errorf("keyring: stored key has invalid length %d", len(key))
	}
	return key, nil
}

// Export returns the master key as a base64 string for secure transfer to
// another device.
func Export() (string, error) {
	key, err := Get()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Delete removes the master key from the keyring. Mainly useful for tests and
// for a future `tpm reset` command.
func Delete() error {
	err := gokeyring.Delete(service, account)
	if err != nil && !errors.Is(err, gokeyring.ErrNotFound) {
		return fmt.Errorf("keyring: delete: %w", err)
	}
	return nil
}

func store(key []byte) error {
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := gokeyring.Set(service, account, encoded); err != nil {
		return fmt.Errorf("keyring: store: %w", err)
	}
	return nil
}
