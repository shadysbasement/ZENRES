package entitlements

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestTierManifests(t *testing.T) {
	s, g, h := For(Starter), For(Growth), For(ImporterHub)
	if slices.Contains(s.VisibleModules, "freight/marketplace") || slices.Contains(s.VisibleModules, "importer-hub") {
		t.Fatal("Starter must not see higher-tier modules")
	}
	if *s.Limits.MaxPlants != 1 || *s.Limits.MaxCNSKUs != 5 || *g.Limits.MaxPlants != 3 || g.Limits.MaxCNSKUs != nil {
		t.Fatal("limits do not match blueprint")
	}
	if !h.Has("consolidated_cbam_xml") || h.Limits.IncludedSuppliers != 15 {
		t.Fatal("Importer Hub capabilities missing")
	}
}

func TestRequireCapabilityReturnsStructured403(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("SECRET DATA")) })
	h := RequireCapability("freightlink_core", "starter.freight_search")(ok)
	req := httptest.NewRequest("GET", "/api/v1/freight/matches", nil).WithContext(WithManifest(httptest.NewRequest("GET", "/", nil).Context(), For(Starter)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 || strings.Contains(rec.Body.String(), "SECRET") {
		t.Fatalf("expected 403 with no data, got %d %s", rec.Code, rec.Body)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "capability_unavailable" || body["suggested_tier"] != "growth" || body["hint_code"] != "starter.freight_search" {
		t.Fatalf("unexpected body %v", body)
	}
	req = req.WithContext(WithManifest(req.Context(), For(Growth)))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal("Growth should pass")
	}
}
