package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/shadysbasement/zevra/services/core-api/internal/twin"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

func TestUnlaunchedFeatureCannotReachHandler(t *testing.T) {
	called := false
	h := releaseGate(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/tally/connector-sync", nil))
	if rec.Code != http.StatusForbidden || called {
		t.Fatalf("disabled feature: status %d, handler called %t", rec.Code, called)
	}
}

func TestPublicTokensAreNonGuessable(t *testing.T) {
	re := regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	seen := map[string]bool{}
	for i := 0; i < 2000; i++ {
		tok := NewPublicToken()
		if !re.MatchString(tok) || seen[tok] {
			t.Fatalf("bad or duplicate token %q", tok)
		}
		seen[tok] = true
	}
}

func TestOpsVaultIs404WithoutAdminSession(t *testing.T) {
	s := &Server{Broker: twin.NewBroker(), Limiter: platform.NewMemLimiter()}
	h := s.Routes()
	for _, p := range []string{"/api/v1/ops-vault/margins", "/api/v1/ops-vault/kpis"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", p, nil))
		if rec.Code != 404 {
			t.Fatalf("%s: expected 404, got %d", p, rec.Code)
		}
	}
}

func TestPublicResolverRejectsMalformedTokens(t *testing.T) {
	s := &Server{Broker: twin.NewBroker(), Limiter: platform.NewMemLimiter()}
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/public/verify/c/not-a-token", nil))
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestInboundWebhookRequiresSecret(t *testing.T) {
	s := &Server{Cfg: Config{InboundSecret: "s3cret"}, Broker: twin.NewBroker(), Limiter: platform.NewMemLimiter()}
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/webhooks/mail-inbound", nil))
	if rec.Code != 404 {
		t.Fatalf("expected 404 without secret, got %d", rec.Code)
	}
}
