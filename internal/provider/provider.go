// Package provider turns a stored credential into an actual value to print.
// Each credential "kind" (code, password, ...) maps to a PasswordProvider
// registered in a small registry, keeping the CLI unaware of how a given value
// is produced.
package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/zhangshuaike/tpm/internal/storage"
)

// Result is the outcome of generating a value for a credential.
type Result struct {
	// Code is the generated value (a TOTP code, a static password, ...).
	Code string
	// ExpiresIn is the seconds until Code rotates/expires; 0 when not applicable.
	ExpiresIn int
}

// PasswordProvider produces the current value for a credential of its kind.
type PasswordProvider interface {
	// Kind returns the credential kind this provider handles (e.g. "code").
	Kind() string
	// Generate produces the current value for the credential.
	Generate(ctx context.Context, cred *storage.Credential) (Result, error)
}

// registry maps credential kinds to their provider implementation.
var registry = map[string]PasswordProvider{}

// Register adds a provider to the registry; called from package init functions.
func Register(p PasswordProvider) {
	registry[p.Kind()] = p
}

// Get returns the provider for the given kind.
func Get(kind string) (PasswordProvider, error) {
	p, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("provider: unsupported type %q (supported: %v)", kind, Kinds())
	}
	return p, nil
}

// Kinds returns the sorted list of registered kinds.
func Kinds() []string {
	ks := make([]string, 0, len(registry))
	for k := range registry {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// Supported reports whether a kind has a registered provider.
func Supported(kind string) bool {
	_, ok := registry[kind]
	return ok
}

// Generate looks up the provider for kind and generates the value.
func Generate(ctx context.Context, kind string, cred *storage.Credential) (Result, error) {
	p, err := Get(kind)
	if err != nil {
		return Result{}, err
	}
	return p.Generate(ctx, cred)
}
