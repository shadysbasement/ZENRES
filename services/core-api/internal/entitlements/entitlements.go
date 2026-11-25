// Package entitlements implements Silent Tier Gating on the server. The UI hides unentitled modules;
// this package guarantees zero-trust enforcement: crafted requests get a structured 403 and no data.
package entitlements

import (
	"context"
	"net/http"
	"sort"

	"github.com/shadysbasement/zevra/services/internal/platform"
)

type Tier string

const (
	Starter     Tier = "starter"
	Growth      Tier = "growth"
	ImporterHub Tier = "importer_hub"
)

type Capability string

var starterCaps = []Capability{"core_erp_tally_csv", "basic_cbam", "evidence_vault", "qr_traceability", "zevra_connect"}
var growthCaps = append(append([]Capability{}, starterCaps...), "ocr_ingestion", "whatsapp_floorbot", "precursor_tracking", "cbam_article9_ccts", "tamper_audit_locker", "capex_roi_simulator", "priority_ca_workspace", "freightlink_core")
var importerCaps = append(append([]Capability{}, growthCaps...), "importer_multi_supplier", "consolidated_cbam_xml", "supplier_benchmarking", "customs_audit_pack", "freightlink_enterprise_fleet", "ai_backhaul_bulk")

// ModuleGates maps every web-app module to the capability it needs ("" = all tiers).
var ModuleGates = map[string]Capability{
	"dashboard": "", "plants": "", "products": "", "inventory": "", "finance": "", "inbox": "", "integrations/tally": "",
	"carbon/inputs": "", "carbon/cbam": "", "carbon/exports": "", "carbon/audit": "", "carbon/simulator": "capex_roi_simulator",
	"trace/qr": "", "freight/marketplace": "freightlink_core", "freight/shipments": "freightlink_core", "freight/transporters": "freightlink_core",
	"freight/enterprise": "freightlink_enterprise_fleet", "connect": "", "importer-hub": "importer_multi_supplier",
	"settings/billing": "", "settings/team": "",
}

type Limits struct {
	MaxPlants         *int `json:"max_plants"`
	MaxCNSKUs         *int `json:"max_cn_skus"`
	IncludedSuppliers int  `json:"included_suppliers"`
	ExtraSupplierUSD  int  `json:"extra_supplier_usd"`
}

type Manifest struct {
	Tier           Tier         `json:"tier"`
	Limits         Limits       `json:"limits"`
	Capabilities   []Capability `json:"capabilities"`
	VisibleModules []string     `json:"visible_modules"`
	set            map[Capability]bool
}

func ptr(i int) *int { return &i }

func For(t Tier) Manifest {
	var caps []Capability
	m := Manifest{Tier: t, Limits: Limits{ExtraSupplierUSD: 25}}
	switch t {
	case Growth:
		caps, m.Limits.MaxPlants = growthCaps, ptr(3)
	case ImporterHub:
		caps, m.Limits.IncludedSuppliers = importerCaps, 15
	default:
		m.Tier, caps, m.Limits.MaxPlants, m.Limits.MaxCNSKUs = Starter, starterCaps, ptr(1), ptr(5)
	}
	m.Capabilities = caps
	m.set = map[Capability]bool{}
	for _, c := range caps {
		m.set[c] = true
	}
	for mod, c := range ModuleGates {
		if c == "" || m.set[c] {
			m.VisibleModules = append(m.VisibleModules, mod)
		}
	}
	sort.Strings(m.VisibleModules)
	return m
}

func (m Manifest) Has(c Capability) bool { return m.set[c] }

// SuggestedTier returns the lowest tier that grants c.
func SuggestedTier(c Capability) Tier {
	if For(Growth).Has(c) {
		return Growth
	}
	return ImporterHub
}

type ctxKey struct{}

func WithManifest(ctx context.Context, m Manifest) context.Context {
	return context.WithValue(ctx, ctxKey{}, m)
}
func FromContext(ctx context.Context) (Manifest, bool) {
	m, ok := ctx.Value(ctxKey{}).(Manifest)
	return m, ok
}

// RequireCapability is applied to every gated handler.
func RequireCapability(c Capability, hint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m, ok := FromContext(r.Context())
			if !ok || !m.Has(c) {
				platform.Error(w, http.StatusForbidden, "capability_unavailable", map[string]any{"hint_code": hint, "suggested_tier": SuggestedTier(c)})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
