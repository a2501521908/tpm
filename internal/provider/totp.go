package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/zhangshuaike/tpm/internal/storage"
)

// defaultPeriod is the standard TOTP step in seconds, compatible with Google and
// Microsoft Authenticator.
const defaultPeriod = 30

// TotpProvider implements RFC 6238 TOTP (6 digits, SHA1, 30s), used by Google
// and Microsoft Authenticator by default. It backs the "code" kind.
type TotpProvider struct {
	// now is injectable for deterministic testing; defaults to time.Now.
	now func() time.Time
}

// NewTotpProvider returns a TotpProvider using the real clock.
func NewTotpProvider() *TotpProvider {
	return &TotpProvider{now: time.Now}
}

func init() {
	Register(NewTotpProvider())
}

// Kind returns the credential kind handled by this provider.
func (p *TotpProvider) Kind() string { return storage.KindCode }

// Generate computes the current TOTP code and the seconds left in the window.
func (p *TotpProvider) Generate(_ context.Context, cred *storage.Credential) (Result, error) {
	secret := strings.TrimSpace(cred.Secret)
	if secret == "" {
		return Result{}, fmt.Errorf("provider/code: empty secret")
	}

	period := uint(defaultPeriod)
	if cred.TTL > 0 {
		period = uint(cred.TTL)
	}

	now := p.now()
	code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    period,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return Result{}, fmt.Errorf("provider/code: generate code: %w", err)
	}

	elapsed := now.Unix() % int64(period)
	expiresIn := int(int64(period) - elapsed)

	return Result{Code: code, ExpiresIn: expiresIn}, nil
}
