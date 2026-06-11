package storage

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhangshuaike/tpm/internal/crypto"
)

func newTestStore(t *testing.T) (*LocalStorage, string, []byte) {
	t.Helper()
	dir := t.TempDir()
	key := make([]byte, crypto.KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("key: %v", err)
	}
	s := NewLocalStorage(key)
	if err := s.Init(dir); err != nil {
		t.Fatalf("init: %v", err)
	}
	return s, dir, key
}

func TestSaveGetMultiCredential(t *testing.T) {
	s, _, _ := newTestStore(t)
	in := &Entry{Name: "tencent"}
	in.Set(KindCode, &Credential{Secret: "JBSWY3DPEHPK3PXP", TTL: 30})
	in.Set(KindPassword, &Credential{Secret: "s3cr3t!"})
	if err := s.SaveEntry(in); err != nil {
		t.Fatalf("save: %v", err)
	}

	out, err := s.GetEntry("tencent")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	code, err := out.Get(KindCode)
	if err != nil || code.Secret != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("code credential mismatch: %+v err=%v", code, err)
	}
	pw, err := out.Get(KindPassword)
	if err != nil || pw.Secret != "s3cr3t!" {
		t.Fatalf("password credential mismatch: %+v err=%v", pw, err)
	}
	if got := out.Kinds(); len(got) != 2 || got[0] != KindCode || got[1] != KindPassword {
		t.Fatalf("unexpected kinds: %v", got)
	}
}

func TestFileIsCiphertext(t *testing.T) {
	s, dir, _ := newTestStore(t)
	secret := "SUPER-SECRET-VALUE"
	e := &Entry{Name: "db"}
	e.Set(KindPassword, &Credential{Secret: secret})
	if err := s.SaveEntry(e); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, entriesDir, "db.enc"))
	if err != nil {
		t.Fatalf("read enc: %v", err)
	}
	if bytes.Contains(data, []byte(secret)) {
		t.Fatal("plaintext secret leaked into .enc file")
	}
	if bytes.Contains(data, []byte("password")) {
		t.Fatal("plaintext kind leaked into .enc file")
	}
}

func TestGetMissing(t *testing.T) {
	s, _, _ := newTestStore(t)
	if _, err := s.GetEntry("nope"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListAndDelete(t *testing.T) {
	s, _, _ := newTestStore(t)
	for _, n := range []string{"b-entry", "a-entry"} {
		e := &Entry{Name: n}
		e.Set(KindPassword, &Credential{Secret: "x"})
		if err := s.SaveEntry(e); err != nil {
			t.Fatalf("save %s: %v", n, err)
		}
	}
	names, err := s.ListNames()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(names) != 2 || names[0] != "a-entry" || names[1] != "b-entry" {
		t.Fatalf("unexpected names (want sorted): %v", names)
	}
	if err := s.DeleteEntry("a-entry"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := s.DeleteEntry("a-entry"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestInvalidName(t *testing.T) {
	s, _, _ := newTestStore(t)
	e := &Entry{Name: "../escape"}
	e.Set(KindPassword, &Credential{Secret: "x"})
	if err := s.SaveEntry(e); err == nil {
		t.Fatal("expected error for path-traversal name")
	}
}
