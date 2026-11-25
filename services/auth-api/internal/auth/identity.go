package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"strings"
)

// IdentityProvider is the common OIDC adapter surface. Google and Apple are live implementations;
// ProtonOIDCStub sits behind ENABLE_PROTON_OIDC=false so a future public Proton IdP is a
// zero-refactor cutover. We never fabricate a non-existent Proton OAuth endpoint.
type IdentityProvider interface {
	Name() string
	AuthURL(state, nonce, codeChallenge string) (string, error)
	Exchange(ctx context.Context, code, codeVerifier, nonce string) (*Claims, error)
}

type Claims struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

var ErrProviderDisabled = errors.New("provider_disabled")

// ProtonOIDCStub is intentionally inert until Proton publishes a third-party OIDC issuer.
type ProtonOIDCStub struct{ Enabled bool }

func (ProtonOIDCStub) Name() string { return "proton_oidc" }
func (p ProtonOIDCStub) AuthURL(string, string, string) (string, error) {
	return "", ErrProviderDisabled
}
func (p ProtonOIDCStub) Exchange(context.Context, string, string, string) (*Claims, error) {
	return nil, ErrProviderDisabled
}

/* ---------------- PKCE (RFC 7636, S256) ---------------- */

// PKCEChallenge derives the S256 code_challenge for a verifier.
func PKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// ValidVerifier enforces RFC 7636 §4.1: 43–128 chars of [A-Za-z0-9-._~].
func ValidVerifier(v string) bool {
	if len(v) < 43 || len(v) > 128 {
		return false
	}
	for _, c := range v {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~", c)) {
			return false
		}
	}
	return true
}

/* ---------------- Proton domain recognition ---------------- */

var protonDomains = map[string]bool{"proton.me": true, "protonmail.com": true, "pm.me": true, "protonmail.ch": true}

// MXLookup is swappable in tests.
var MXLookup = net.LookupMX

// IsProtonEmail reports whether the address is hosted by Proton (known domains or MX → protonmail.ch).
func IsProtonEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	if protonDomains[domain] {
		return true
	}
	mx, err := MXLookup(domain)
	if err != nil {
		return false
	}
	for _, r := range mx {
		h := strings.TrimSuffix(strings.ToLower(r.Host), ".")
		if h == "mail.protonmail.ch" || h == "mailsec.protonmail.ch" {
			return true
		}
	}
	return false
}

// IsAppleRelay recognises Hide-My-Email private relay addresses (fully supported).
func IsAppleRelay(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@privaterelay.appleid.com")
}
