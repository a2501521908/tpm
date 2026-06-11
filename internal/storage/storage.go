// Package storage defines the persistence abstraction for TPM. Each named entry
// is stored as an independent, encrypted .enc file so that two devices editing
// different entries never produce a sync conflict. A single entry can hold
// multiple credentials keyed by "kind" (e.g. "code" and "password"), letting one
// name (e.g. "tencent") carry both a dynamic code and a static password.
package storage

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
)

// Credential kinds. A kind is what the user types as the command/type, and also
// selects which provider generates the value.
const (
	KindCode     = "code"     // dynamic one-time code (TOTP)
	KindPassword = "password" // static password / token
)

// ErrNotFound is returned when an entry does not exist.
var ErrNotFound = errors.New("storage: entry not found")

// ErrNoCredential is returned when an entry exists but lacks the requested kind.
var ErrNoCredential = errors.New("storage: entry has no credential of that type")

// validName restricts entry names to filesystem-safe characters so the name can
// be used directly as a file name without risk of path traversal.
var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// Credential is a single secret of a given kind belonging to an entry.
type Credential struct {
	Secret string                 `json:"secret"`
	TTL    int                    `json:"ttl,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// Entry is the decrypted, in-memory representation of a named item. It holds one
// or more credentials keyed by kind.
type Entry struct {
	Name        string                 `json:"name"`
	Credentials map[string]*Credential `json:"credentials,omitempty"`
}

// Kinds returns the sorted list of credential kinds present in the entry.
func (e *Entry) Kinds() []string {
	kinds := make([]string, 0, len(e.Credentials))
	for k := range e.Credentials {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}

// Get returns the credential of the given kind, or ErrNoCredential.
func (e *Entry) Get(kind string) (*Credential, error) {
	cred, ok := e.Credentials[kind]
	if !ok {
		return nil, ErrNoCredential
	}
	return cred, nil
}

// Set adds or replaces a credential of the given kind.
func (e *Entry) Set(kind string, cred *Credential) {
	if e.Credentials == nil {
		e.Credentials = map[string]*Credential{}
	}
	e.Credentials[kind] = cred
}

// StorageProvider abstracts where and how encrypted entries are persisted.
type StorageProvider interface {
	Init(dataDir string) error
	SaveEntry(entry *Entry) error
	GetEntry(name string) (*Entry, error)
	DeleteEntry(name string) error
	ListNames() ([]string, error)
	Exists(name string) (bool, error)
}

// ValidateName ensures an entry name is safe to use as a file name.
func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("storage: invalid entry name %q (allowed: letters, digits, '.', '_', '-')", name)
	}
	return nil
}
