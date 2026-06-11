package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func mustKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := mustKey(t)
	plaintext := []byte(`{"name":"github-2fa","type":"totp"}`)

	blob, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Contains(blob, plaintext) {
		t.Fatal("ciphertext leaked plaintext bytes")
	}

	got, err := Decrypt(blob, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round trip mismatch: got %q want %q", got, plaintext)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	blob, err := Encrypt([]byte("secret"), mustKey(t))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := Decrypt(blob, mustKey(t)); err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptTamperedFails(t *testing.T) {
	key := mustKey(t)
	blob, err := Encrypt([]byte("secret"), key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	blob[len(blob)-1] ^= 0xFF // flip a bit in the tag
	if _, err := Decrypt(blob, key); err == nil {
		t.Fatal("expected error decrypting tampered ciphertext")
	}
}

func TestInvalidKeySize(t *testing.T) {
	if _, err := Encrypt([]byte("x"), []byte("short")); err != ErrInvalidKeySize {
		t.Fatalf("expected ErrInvalidKeySize, got %v", err)
	}
}

func TestCiphertextTooShort(t *testing.T) {
	if _, err := Decrypt([]byte("abc"), mustKey(t)); err != ErrCiphertextTooShort {
		t.Fatalf("expected ErrCiphertextTooShort, got %v", err)
	}
}
