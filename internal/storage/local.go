package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zhangshuaike/tpm/internal/crypto"
)

const (
	entriesDir = "entries"
	metaFile   = ".tpm_meta"
	encExt     = ".enc"

	// metaVersion is bumped when the on-disk format changes.
	metaVersion = 1
)

// meta is the non-sensitive global metadata persisted in plaintext.
type meta struct {
	Version   int    `json:"version"`
	Crypto    string `json:"crypto"`
	CreatedAt string `json:"created_at"`
}

// LocalStorage persists encrypted entries as individual files under
// <dataDir>/entries/<name>.enc. It is safe to place dataDir inside a cloud sync
// folder.
type LocalStorage struct {
	dataDir string
	key     []byte
}

// NewLocalStorage creates a LocalStorage bound to the given master key.
func NewLocalStorage(key []byte) *LocalStorage {
	return &LocalStorage{key: key}
}

// Init creates the directory layout and writes .tpm_meta if absent.
func (s *LocalStorage) Init(dataDir string) error {
	s.dataDir = dataDir
	if err := os.MkdirAll(filepath.Join(dataDir, entriesDir), 0o700); err != nil {
		return fmt.Errorf("storage: create entries dir: %w", err)
	}
	metaPath := filepath.Join(dataDir, metaFile)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		m := meta{
			Version:   metaVersion,
			Crypto:    "AES-256-GCM",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Errorf("storage: marshal meta: %w", err)
		}
		if err := os.WriteFile(metaPath, data, 0o600); err != nil {
			return fmt.Errorf("storage: write meta: %w", err)
		}
	}
	return nil
}

// Open binds an existing store at dataDir without (re)writing metadata.
func (s *LocalStorage) Open(dataDir string) error {
	s.dataDir = dataDir
	if _, err := os.Stat(filepath.Join(dataDir, entriesDir)); err != nil {
		return fmt.Errorf("storage: data dir not initialized at %s: %w", dataDir, err)
	}
	return nil
}

func (s *LocalStorage) entryPath(name string) string {
	return filepath.Join(s.dataDir, entriesDir, name+encExt)
}

// SaveEntry encrypts the entry JSON and writes it atomically.
func (s *LocalStorage) SaveEntry(entry *Entry) error {
	if err := ValidateName(entry.Name); err != nil {
		return err
	}
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("storage: marshal entry: %w", err)
	}
	blob, err := crypto.Encrypt(plaintext, s.key)
	if err != nil {
		return err
	}
	path := s.entryPath(entry.Name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o600); err != nil {
		return fmt.Errorf("storage: write entry: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("storage: finalize entry: %w", err)
	}
	return nil
}

// GetEntry reads and decrypts an entry.
func (s *LocalStorage) GetEntry(name string) (*Entry, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	blob, err := os.ReadFile(s.entryPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: read entry: %w", err)
	}
	plaintext, err := crypto.Decrypt(blob, s.key)
	if err != nil {
		return nil, err
	}
	var entry Entry
	if err := json.Unmarshal(plaintext, &entry); err != nil {
		return nil, fmt.Errorf("storage: unmarshal entry: %w", err)
	}
	return &entry, nil
}

// DeleteEntry removes an entry file.
func (s *LocalStorage) DeleteEntry(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	err := os.Remove(s.entryPath(name))
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("storage: delete entry: %w", err)
	}
	return nil
}

// ListNames returns all entry names sorted alphabetically.
func (s *LocalStorage) ListNames() ([]string, error) {
	dir := filepath.Join(s.dataDir, entriesDir)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("storage: list entries: %w", err)
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, encExt) {
			continue
		}
		names = append(names, strings.TrimSuffix(name, encExt))
	}
	sort.Strings(names)
	return names, nil
}

// Exists reports whether an entry file is present.
func (s *LocalStorage) Exists(name string) (bool, error) {
	if err := ValidateName(name); err != nil {
		return false, err
	}
	_, err := os.Stat(s.entryPath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("storage: stat entry: %w", err)
}

// compile-time assertion that LocalStorage satisfies the interface.
var _ StorageProvider = (*LocalStorage)(nil)
