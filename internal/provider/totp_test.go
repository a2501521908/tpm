package provider

import (
	"context"
	"testing"
	"time"

	"github.com/zhangshuaike/tpm/internal/storage"
)

// RFC 6238 reference secret: base32 of "12345678901234567890" (SHA1). At Unix
// time 59 the 6-digit TOTP is 287082, a widely cited test vector.
const rfcSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestTotpKnownVector(t *testing.T) {
	p := &TotpProvider{now: func() time.Time { return time.Unix(59, 0) }}
	cred := &storage.Credential{Secret: rfcSecret, TTL: 30}

	res, err := p.Generate(context.Background(), cred)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if res.Code != "287082" {
		t.Fatalf("got code %q want 287082", res.Code)
	}
	if res.ExpiresIn != 1 {
		t.Fatalf("got ExpiresIn %d want 1", res.ExpiresIn)
	}
}

func TestTotpEmptySecret(t *testing.T) {
	p := NewTotpProvider()
	if _, err := p.Generate(context.Background(), &storage.Credential{}); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestStaticProvider(t *testing.T) {
	p := NewStaticProvider()
	res, err := p.Generate(context.Background(), &storage.Credential{Secret: "hunter2"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if res.Code != "hunter2" || res.ExpiresIn != 0 {
		t.Fatalf("unexpected static result: %+v", res)
	}
}

func TestRegistryLookup(t *testing.T) {
	if _, err := Get(storage.KindCode); err != nil {
		t.Fatalf("code should be registered: %v", err)
	}
	if _, err := Get(storage.KindPassword); err != nil {
		t.Fatalf("password should be registered: %v", err)
	}
	if _, err := Get("does-not-exist"); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}
