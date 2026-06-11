package provider

import (
	"context"
	"fmt"

	"github.com/zhangshuaike/tpm/internal/storage"
)

// StaticProvider returns a stored static secret verbatim. It backs the
// "password" kind: a fixed password or token that does not rotate.
type StaticProvider struct{}

// NewStaticProvider returns a StaticProvider.
func NewStaticProvider() *StaticProvider { return &StaticProvider{} }

func init() {
	Register(NewStaticProvider())
}

// Kind returns the credential kind handled by this provider.
func (p *StaticProvider) Kind() string { return storage.KindPassword }

// Generate returns the stored secret as-is.
func (p *StaticProvider) Generate(_ context.Context, cred *storage.Credential) (Result, error) {
	if cred.Secret == "" {
		return Result{}, fmt.Errorf("provider/password: empty secret")
	}
	return Result{Code: cred.Secret, ExpiresIn: 0}, nil
}
