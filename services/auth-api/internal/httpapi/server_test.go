package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/shadysbasement/zevra/services/auth-api/internal/auth"
	"github.com/shadysbasement/zevra/services/auth-api/internal/pgp"
	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

type capture struct {
	mu   sync.Mutex
	last string
}

func (c *capture) Send(_ context.Context, _, _, body string, _ *crypto.Key) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last = body
	return nil
}

func (c *capture) token() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := regexp.MustCompile(`token=([A-Za-z0-9_%-]+)`).FindStringSubmatch(c.last)
	if m == nil {
		return ""
	}
	t, _ := url.QueryUnescape(m[1])
	return t
}

type fakeIdP struct{ claims auth.Claims }

func (f fakeIdP) Name() string { return "google" }
func (f fakeIdP) AuthURL(state, nonce, ch string) (string, error) {
	return "https://idp.example/auth?state=" + state + "&code_challenge=" + ch, nil
}
func (f fakeIdP) Exchange(_ context.Context, code, verifier, _ string) (*auth.Claims, error) {
	c := f.claims
	return &c, nil
}

func newServer(t *testing.T) (*Server, *store.Memory, *capture) {
	t.Helper()
	mem := store.NewMemory()
	v, err := pgp.New(mem, "zevra.io")
	if err != nil {
		t.Fatal(err)
	}
	wa, _ := webauthn.New(&webauthn.Config{RPID: "localhost", RPDisplayName: "Zevra", RPOrigins: []string{"http://localhost:5173"}})
	mail := &capture{}
	s := &Server{
		Cfg:   Config{PublicURL: "https://zevra.io", AppURL: "https://app.zevra.io", CookieDomain: ".zevra.io", CookieSecure: true, GateSecret: "gate-secret", InternalSecret: "internal"},
		Users: mem, KV: mem, Vault: v, Limiter: platform.NewMemLimiter(), Mail: mail, WebAuthn: wa, Now: time.Now,
		Providers: map[string]auth.IdentityProvider{"google": fakeIdP{claims: auth.Claims{Subject: "g-123", Email: "ananya@deccanalloy.in", EmailVerified: true, Name: "Ananya"}}},
	}
	auth.MXLookup = func(string) ([]*net.MX, error) { return nil, nil }
	return s, mem, mail
}

func do(t *testing.T, h http.Handler, method, path string, body any, cookies []*http.Cookie, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func cookieNamed(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestRegisterVerifyLoginAndCookieFlags(t *testing.T) {
	s, _, mail := newServer(t)
	h := s.Routes()
	rec := do(t, h, "POST", "/api/v1/auth/register", map[string]string{"full_name": "Ananya Deshpande", "email": "Ananya@DeccanAlloy.in", "password": "Deccan-Alloy-2026!"}, nil, nil)
	if rec.Code != 201 {
		t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, h, "POST", "/api/v1/auth/login", map[string]string{"email": "ananya@deccanalloy.in", "password": "Deccan-Alloy-2026!"}, nil, nil); rec.Code != 403 {
		t.Fatalf("unverified login should be 403, got %d", rec.Code)
	}
	rec = do(t, h, "POST", "/api/v1/auth/verify-email", map[string]string{"token": mail.token()}, nil, nil)
	if rec.Code != 200 {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "POST", "/api/v1/auth/login", map[string]string{"email": "ananya@deccanalloy.in", "password": "Deccan-Alloy-2026!"}, nil, nil)
	if rec.Code != 200 {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	c := cookieNamed(rec, SessionCookie)
	if c == nil || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Domain != "zevra.io" && c.Domain != ".zevra.io" {
		t.Fatalf("session cookie flags wrong: %+v", c)
	}
	if len(c.Value) != 43 {
		t.Fatal("session token should be 256-bit opaque")
	}
	if rec := do(t, h, "POST", "/api/v1/auth/login", map[string]string{"email": "ananya@deccanalloy.in", "password": "wrong-password-123"}, nil, nil); rec.Code != 401 {
		t.Fatalf("wrong password: %d", rec.Code)
	}
}

func TestProtonMagicLinkIsSingleUse(t *testing.T) {
	s, _, mail := newServer(t)
	h := s.Routes()
	rec := do(t, h, "POST", "/api/v1/auth/proton/magic-link", map[string]any{"email": "founder@proton.me"}, nil, nil)
	if rec.Code != 200 {
		t.Fatalf("magic link: %d %s", rec.Code, rec.Body)
	}
	tok := mail.token()
	first := do(t, h, "POST", "/api/v1/auth/proton/verify", map[string]string{"token": tok}, nil, nil)
	if first.Code != 200 || !strings.Contains(first.Body.String(), `"offer_passkey":true`) {
		t.Fatalf("first use: %d %s", first.Code, first.Body)
	}
	second := do(t, h, "POST", "/api/v1/auth/proton/verify", map[string]string{"token": tok}, nil, nil)
	if second.Code != 410 || !strings.Contains(second.Body.String(), "token_consumed") {
		t.Fatalf("second use must fail: %d %s", second.Code, second.Body)
	}
}

func TestProtonRejectsNonProtonWithoutOverride(t *testing.T) {
	s, _, _ := newServer(t)
	rec := do(t, s.Routes(), "POST", "/api/v1/auth/proton/magic-link", map[string]any{"email": "someone@gmail.com"}, nil, nil)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "not_proton") {
		t.Fatalf("expected not_proton, got %d %s", rec.Code, rec.Body)
	}
}

func TestOAuthPKCEAndAccountLinkingGuard(t *testing.T) {
	s, mem, mail := newServer(t)
	h := s.Routes()
	// Pre-existing manual account with the same email.
	do(t, h, "POST", "/api/v1/auth/register", map[string]string{"full_name": "Ananya", "email": "ananya@deccanalloy.in", "password": "Deccan-Alloy-2026!"}, nil, nil)
	do(t, h, "POST", "/api/v1/auth/verify-email", map[string]string{"token": mail.token()}, nil, nil)

	start := do(t, h, "GET", "/api/v1/auth/oauth/google/start", nil, nil, nil)
	loc, _ := url.Parse(start.Header().Get("Location"))
	state, challenge := loc.Query().Get("state"), loc.Query().Get("code_challenge")
	raw, ok := mem.Take("oauth:" + state)
	if !ok {
		t.Fatal("state not stored")
	}
	mem.Put("oauth:"+state, raw, time.Minute)
	var st oauthState
	_ = json.Unmarshal(raw, &st)
	if !auth.ValidVerifier(st.Verifier) || auth.PKCEChallenge(st.Verifier) != challenge {
		t.Fatal("PKCE S256 challenge must derive from the stored verifier")
	}

	cb := do(t, h, "GET", "/api/v1/auth/oauth/google/callback?state="+state+"&code=abc", nil, nil, nil)
	if !strings.Contains(cb.Header().Get("Location"), "/login?link=") {
		t.Fatalf("must not silently merge — expected link confirmation redirect, got %s", cb.Header().Get("Location"))
	}
	if _, err := mem.IdentityBySubject(context.Background(), "google", "g-123"); err == nil {
		t.Fatal("identity must not be linked before confirmation")
	}
	linkTok := strings.Split(strings.Split(cb.Header().Get("Location"), "link=")[1], "&")[0]
	if rec := do(t, h, "POST", "/api/v1/auth/link/confirm", map[string]string{"link_token": linkTok, "password": "Deccan-Alloy-2026!"}, nil, nil); rec.Code != 200 {
		t.Fatalf("link confirm: %d %s", rec.Code, rec.Body)
	}
	if _, err := mem.IdentityBySubject(context.Background(), "google", "g-123"); err != nil {
		t.Fatal("identity should be linked after password confirmation")
	}
	// Replayed state must fail.
	replay := do(t, h, "GET", "/api/v1/auth/oauth/google/callback?state="+state+"&code=abc", nil, nil, nil)
	if !strings.Contains(replay.Header().Get("Location"), "error=oauth_state") {
		t.Fatal("state must be single-use")
	}
}

func TestVaultEndpointsAreStealth(t *testing.T) {
	s, _, _ := newServer(t)
	h := s.Routes()
	if rec := do(t, h, "POST", "/api/v1/vault-auth/challenge", map[string]string{"identifier": "AAAA"}, nil, nil); rec.Code != 404 {
		t.Fatalf("missing gate header must 404, got %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/api/v1/vault-auth/challenge", map[string]string{"identifier": strings.Repeat("A", 40)}, nil, map[string]string{"X-Zevra-Gate": "gate-secret"}); rec.Code != 200 {
		t.Fatalf("gated challenge should succeed (decoy), got %d", rec.Code)
	}
	if rec := do(t, h, "POST", "/internal/vault-auth/session", map[string]string{"token": "x"}, nil, nil); rec.Code != 404 {
		t.Fatalf("internal endpoints must 404 without secret, got %d", rec.Code)
	}
	for _, p := range []string{"/admin", "/administrator", "/ops", "/pgp-login"} {
		if rec := do(t, h, "GET", p, nil, nil, nil); rec.Code != 404 {
			t.Fatalf("%s should 404", p)
		}
	}
}

func TestVaultClientIPUsesAuthenticatedProxyHeader(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/v1/vault-auth/challenge", nil)
	r.RemoteAddr = "192.0.2.12:443"
	r.Header.Set("X-Zenres-Client-IP", "198.51.100.42")
	if got := vaultClientIP(r); got != "198.51.100.42" {
		t.Fatalf("forwarded client IP = %q", got)
	}
	r.Header.Set("X-Zenres-Client-IP", "not-an-ip")
	if got := vaultClientIP(r); got != "192.0.2.12" {
		t.Fatalf("invalid forwarded client IP was trusted: %q", got)
	}
}

func TestCSRFRequiredWithSession(t *testing.T) {
	s, _, mail := newServer(t)
	h := s.Routes()
	do(t, h, "POST", "/api/v1/auth/proton/magic-link", map[string]any{"email": "a@proton.me"}, nil, nil)
	rec := do(t, h, "POST", "/api/v1/auth/proton/verify", map[string]string{"token": mail.token()}, nil, nil)
	sess, csrf := cookieNamed(rec, SessionCookie), cookieNamed(rec, CSRFCookie)
	body := map[string]any{"role": "msme", "org_name": "Deccan Alloy Works", "country_code": "IN", "plan": "growth"}
	if r := do(t, h, "POST", "/api/v1/onboarding/complete", body, []*http.Cookie{sess, csrf}, nil); r.Code != 403 {
		t.Fatalf("missing CSRF header should be rejected, got %d", r.Code)
	}
	r := do(t, h, "POST", "/api/v1/onboarding/complete", body, []*http.Cookie{sess, csrf}, map[string]string{"X-CSRF-Token": csrf.Value})
	if r.Code != 201 || !strings.Contains(r.Body.String(), "app.zevra.io/dashboard") {
		t.Fatalf("onboarding: %d %s", r.Code, r.Body)
	}
}
