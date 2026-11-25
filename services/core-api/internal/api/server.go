// Package api is the Zevra core platform HTTP API. Every tenant handler runs inside rls.Tx and every
// gated route is wrapped in entitlements.RequireCapability — the server never trusts the UI.
package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadysbasement/zevra/services/core-api/internal/cbam"
	"github.com/shadysbasement/zevra/services/core-api/internal/entitlements"
	"github.com/shadysbasement/zevra/services/core-api/internal/freight"
	"github.com/shadysbasement/zevra/services/core-api/internal/hashchain"
	"github.com/shadysbasement/zevra/services/core-api/internal/rls"
	"github.com/shadysbasement/zevra/services/core-api/internal/tally"
	"github.com/shadysbasement/zevra/services/core-api/internal/twin"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

type Config struct {
	AppOrigin      string
	WorkerURL      string
	AuthAPIURL     string
	InternalSecret string
	InboundSecret  string // shared secret for the Postmark/SES inbound webhook
	TypstBin       string
	TemplatesDir   string
}

type Server struct {
	Cfg     Config
	DB      *pgxpool.Pool
	Broker  *twin.Broker
	Limiter platform.Limiter
}

type Principal struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Role   string
	Org    string
}

type principalKey struct{}

func principal(r *http.Request) Principal {
	p, _ := r.Context().Value(principalKey{}).(Principal)
	return p
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Recoverer, platform.SecurityHeaders, s.cors)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.DB.Ping(r.Context()); err != nil {
			platform.Error(w, 503, "db_unavailable")
			return
		}
		platform.JSON(w, 200, map[string]string{"status": "ok"})
	})

	// Public, sanitised QR passport resolver.
	r.Get("/api/v1/public/verify/{type}/{token}", s.publicVerify)
	// Inbound mail webhook (inbox+<slug>@mail.zevra.io), authenticated by shared secret.
	r.Post("/api/v1/webhooks/mail-inbound", s.mailInbound)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(s.authenticate)
		gate := entitlements.RequireCapability

		r.Get("/me/manifest", s.manifest)
		r.Post("/hints/dismiss", s.dismissHint)

		r.Get("/plants/{id}/digital-twin", s.digitalTwin)
		r.Patch("/zones/{id}/live-state", s.patchZone)
		r.Get("/stream/plant/{id}", s.stream)

		r.Get("/dashboard/financials", s.financials)
		r.Post("/tally/import-csv", s.tallyImport)
		r.Post("/tally/connector-sync", s.tallyConnectorSync)

		r.Post("/mail/connections", s.mailConnection)
		r.Get("/notifications", s.notifications)
		r.Post("/notifications/{id}/convert-rfq", s.convertRFQ)

		r.Post("/carbon/activity", s.carbonActivity)
		r.With(gate("ocr_ingestion", "starter.ocr")).Post("/carbon/ocr-upload", s.ocrUpload)
		r.Get("/carbon/cbam-summary", s.cbamSummary)
		r.Post("/carbon/export/xml", s.exportXML)
		r.With(gate("capex_roi_simulator", "starter.capex")).Post("/carbon/simulator", s.simulator)
		r.With(gate("tamper_audit_locker", "starter.revised_upload")).Get("/carbon/audit/verify", s.verifyAuditChain)

		r.Post("/qr/batch", s.qrBatch)
		r.Post("/qr/container", s.qrContainer)
		r.Post("/qr/container/{id}/link-batches", s.qrLink)
		r.Post("/qr/{id}/anchor-hash", s.qrAnchor)
		r.Post("/qr/{id}/revoke", s.qrRevoke)
		r.Get("/qr/{id}/label.pdf", s.qrLabel)

		r.Group(func(r chi.Router) {
			r.Use(gate("freightlink_core", "starter.freight_search"))
			r.Post("/freight/listings", s.freightListing)
			r.Get("/freight/matches", s.freightMatches)
			r.Post("/freight/quotes", s.freightQuote)
			r.Post("/freight/bookings/{id}/scan-qr", s.freightScan)
		})
		r.With(gate("freightlink_enterprise_fleet", "growth.fleet")).Get("/freight/enterprise/fleet", s.fleet)

		r.Get("/connect/suppliers", s.connectSuppliers)
		r.Post("/connect/rfqs", s.createRFQ)
		r.With(gate("importer_multi_supplier", "growth.bulk_invite")).Post("/importer/suppliers/bulk-invite", s.bulkInvite)
		r.With(gate("supplier_benchmarking", "growth.benchmark")).Get("/importer/benchmarks", s.benchmarks)
	})

	// PGP-verified super-admin console: 404 unless the admin session validates.
	r.Route("/api/v1/ops-vault", func(r chi.Router) {
		r.Use(s.opsOnly)
		r.Get("/margins", s.opsMargins)
		r.Get("/kpis", s.opsKPIs)
		r.Post("/cbam-rules", s.opsPublishRule)
		r.Post("/approvals/{id}", s.opsApprove)
	})
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "Not Found", 404) })
	return r
}

/* ------------------------------------------------------------------ */
/* middleware                                                          */
/* ------------------------------------------------------------------ */

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" && o == s.Cfg.AppOrigin {
			w.Header().Set("Access-Control-Allow-Origin", o)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
				w.WriteHeader(204)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// authenticate resolves the shared zv_session cookie (issued by the auth service on .zevra.io)
// to a principal and loads the tenant's entitlement manifest.
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("zv_session")
		if err != nil {
			platform.Error(w, 401, "unauthenticated")
			return
		}
		var p Principal
		var tier string
		err = s.DB.QueryRow(r.Context(), `SELECT s.user_id, m.org_id, m.role, o.name, o.subscription_tier
			FROM user_sessions s JOIN org_memberships m ON m.user_id = s.user_id JOIN organizations o ON o.id = m.org_id
			WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > NOW()
			ORDER BY (m.org_id = s.org_id) DESC NULLS LAST LIMIT 1`, platform.SHA256(c.Value)).Scan(&p.UserID, &p.OrgID, &p.Role, &p.Org, &tier)
		if err != nil {
			platform.Error(w, 401, "unauthenticated")
			return
		}
		ctx := context.WithValue(r.Context(), principalKey{}, p)
		ctx = entitlements.WithManifest(ctx, entitlements.For(entitlements.Tier(tier)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) opsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tok == "" || s.Cfg.AuthAPIURL == "" {
			http.Error(w, "Not Found", 404)
			return
		}
		body, _ := json.Marshal(map[string]string{"token": tok, "ip": r.Header.Get("X-Client-IP"), "ua": r.Header.Get("X-Client-UA")})
		req, _ := http.NewRequestWithContext(r.Context(), "POST", s.Cfg.AuthAPIURL+"/internal/vault-auth/session", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Secret", s.Cfg.InternalSecret)
		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != 200 {
			http.Error(w, "Not Found", 404)
			return
		}
		res.Body.Close()
		next.ServeHTTP(w, r)
	})
}

/* ------------------------------------------------------------------ */
/* helpers                                                             */
/* ------------------------------------------------------------------ */

func (s *Server) tx(r *http.Request, fn func(pgx.Tx) error, opts ...rls.Opt) error {
	return rls.Tx(r.Context(), s.DB, principal(r).OrgID, fn, opts...)
}

func rowsJSON(ctx context.Context, tx pgx.Tx, sql string, args ...any) ([]map[string]any, error) {
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, pgx.RowToMap)
	if out == nil {
		out = []map[string]any{}
	}
	return out, err
}

func (s *Server) list(w http.ResponseWriter, r *http.Request, sql string, args ...any) {
	var out []map[string]any
	err := s.tx(r, func(tx pgx.Tx) error {
		var err error
		out, err = rowsJSON(r.Context(), tx, sql, args...)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"items": out})
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, 404, "not_found")
		return
	}
	slog.Error("core-api", "err", err)
	platform.Error(w, 500, "internal_error")
}

func idParam(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	return id, err == nil
}

/* ------------------------------------------------------------------ */
/* entitlements                                                        */
/* ------------------------------------------------------------------ */

func (s *Server) manifest(w http.ResponseWriter, r *http.Request) {
	m, _ := entitlements.FromContext(r.Context())
	p := principal(r)
	dismissed := map[string]int64{}
	var plantID *uuid.UUID
	_ = s.tx(r, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT hint_code, dismissed_at FROM user_hint_dismissals WHERE user_id=$1 AND dismissed_at > NOW() - interval '14 days'`, p.UserID)
		if err == nil {
			for rows.Next() {
				var c string
				var t time.Time
				if rows.Scan(&c, &t) == nil {
					dismissed[c] = t.UnixMilli()
				}
			}
			rows.Close()
		}
		_ = tx.QueryRow(r.Context(), `SELECT id FROM plants ORDER BY created_at LIMIT 1`).Scan(&plantID)
		return nil
	})
	platform.JSON(w, 200, map[string]any{"tier": m.Tier, "limits": m.Limits, "capabilities": m.Capabilities, "visible_modules": m.VisibleModules, "dismissed": dismissed, "org_name": p.Org, "plant_id": plantID})
}

func (s *Server) dismissHint(w http.ResponseWriter, r *http.Request) {
	var in struct {
		HintCode string `json:"hint_code"`
	}
	if platform.DecodeLoose(r, &in) != nil || len(in.HintCode) > 64 || in.HintCode == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	err := s.tx(r, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO user_hint_dismissals (user_id, hint_code) VALUES ($1,$2)
			ON CONFLICT (user_id, hint_code) DO UPDATE SET dismissed_at = NOW()`, principal(r).UserID, in.HintCode)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(204)
}

/* ------------------------------------------------------------------ */
/* digital twin & telemetry                                            */
/* ------------------------------------------------------------------ */

func (s *Server) digitalTwin(w http.ResponseWriter, r *http.Request) {
	var zones []map[string]any
	err := s.tx(r, func(tx pgx.Tx) error {
		var err error
		q := `SELECT z.id::text, z.zone_code AS code, z.name, z.zone_type AS type, z.geometry_3d->'pos' AS pos, z.geometry_3d->'size' AS size,
			COALESCE(z.supervisor_name,'') AS supervisor, z.active_workers, z.shift_target_workers, z.live_power_kw::float8 AS live_power_kw, z.fuel_lph::float8 AS fuel_lph
			FROM factory_zones z JOIN plants p ON p.id = z.plant_id`
		if id, ok := idParam(r); ok {
			zones, err = rowsJSON(r.Context(), tx, q+` WHERE p.id=$1 ORDER BY z.zone_code`, id)
		} else {
			zones, err = rowsJSON(r.Context(), tx, q+` ORDER BY z.zone_code`)
		}
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"zones": zones})
}

func (s *Server) patchZone(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		platform.Error(w, 400, "invalid_id")
		return
	}
	var in struct {
		ActiveWorkers *int     `json:"active_workers"`
		LivePowerKW   *float64 `json:"live_power_kw"`
	}
	if platform.DecodeLoose(r, &in) != nil || (in.ActiveWorkers != nil && (*in.ActiveWorkers < 0 || *in.ActiveWorkers > 999)) {
		platform.Error(w, 400, "invalid_input")
		return
	}
	err := s.tx(r, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `UPDATE factory_zones SET active_workers = COALESCE($2, active_workers), live_power_kw = COALESCE($3, live_power_kw), updated_at = NOW() WHERE id = $1`, id, in.ActiveWorkers, in.LivePowerKW)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	patch := map[string]any{}
	if in.ActiveWorkers != nil {
		patch["active_workers"] = *in.ActiveWorkers
	}
	if in.LivePowerKW != nil {
		patch["live_power_kw"] = *in.LivePowerKW
	}
	s.Broker.Publish(principal(r).OrgID, map[string]any{"type": "zone", "id": id.String(), "patch": patch})
	platform.JSON(w, 200, map[string]any{"status": "updated"})
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	s.Broker.Serve(w, r, principal(r).OrgID)
}

/* ------------------------------------------------------------------ */
/* finance & tally                                                     */
/* ------------------------------------------------------------------ */

func (s *Server) financials(w http.ResponseWriter, r *http.Request) {
	s.list(w, r, `SELECT to_char(date_trunc('month', voucher_date), 'YYYY-MM') AS month,
		SUM(amount_usd) FILTER (WHERE voucher_type='sales')::float8 AS revenue,
		SUM(cost_usd) FILTER (WHERE voucher_type='sales')::float8 AS cogs,
		SUM(amount_usd) FILTER (WHERE voucher_type IN ('expense','utility','freight'))::float8 AS opex,
		SUM(amount_usd) FILTER (WHERE voucher_type='sales' AND is_export)::float8 AS export_revenue
		FROM financial_vouchers WHERE voucher_date > NOW() - interval '12 months' GROUP BY 1 ORDER BY 1`)
}

func (s *Server) upsertVouchers(r *http.Request, vs []tally.Voucher, source string) (inserted, dup int, err error) {
	err = s.tx(r, func(tx pgx.Tx) error {
		for _, v := range vs {
			tag, err := tx.Exec(r.Context(), `INSERT INTO financial_vouchers (org_id, source, external_ref, voucher_type, counterparty, voucher_date, amount_usd, cost_usd, is_export)
				VALUES (current_org(), $1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (org_id, source, external_ref) DO NOTHING`,
				source, v.ExternalRef, v.Type, v.Counterparty, v.Date, v.AmountINR/88, v.AmountINR/88*0.71, v.IsExport)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 1 {
				inserted++
			} else {
				dup++
			}
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO tally_sync_log (org_id, source, inserted, duplicates) VALUES (current_org(), $1, $2, $3)`, source, inserted, dup)
		return err
	})
	return
}

func (s *Server) tallyImport(w http.ResponseWriter, r *http.Request) {
	body := io.LimitReader(r.Body, 10<<20)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		f, _, err := r.FormFile("file")
		if err != nil {
			platform.Error(w, 400, "file_required")
			return
		}
		defer f.Close()
		body = f
	}
	vs, errs := tally.ParseCSV(body, "tally_csv")
	ins, dup, err := s.upsertVouchers(r, vs, "tally_csv")
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"inserted": ins, "duplicates": dup, "errors": len(errs)})
}

func (s *Server) tallyConnectorSync(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Mode     string `json:"mode"` // xml | odbc
		Vouchers []struct {
			GUID, Type, Party, Date string
			AmountINR               float64 `json:"amount_inr"`
			Export                  bool
		} `json:"vouchers"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	src := "tally_xml"
	if in.Mode == "odbc" {
		src = "tally_odbc"
	}
	var vs []tally.Voucher
	for _, v := range in.Vouchers {
		d, _ := time.Parse("2006-01-02", v.Date)
		vs = append(vs, tally.Voucher{ExternalRef: v.GUID, Type: v.Type, Counterparty: v.Party, Date: d, AmountINR: v.AmountINR, IsExport: v.Export})
	}
	ins, dup, err := s.upsertVouchers(r, vs, src)
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"inserted": ins, "duplicates": dup})
}

/* ------------------------------------------------------------------ */
/* mail & bell                                                         */
/* ------------------------------------------------------------------ */

func (s *Server) mailConnection(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Provider, AccountEmail string
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	// OAuth tokens / IMAP credentials are exchanged and envelope-encrypted by the connector flow;
	// this endpoint only registers the connection shell.
	err := s.tx(r, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO mail_connections (org_id, provider, account_email) VALUES (current_org(), $1, $2)`, in.Provider, in.AccountEmail)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]string{"status": "pending_authorisation"})
}

func (s *Server) notifications(w http.ResponseWriter, r *http.Request) {
	s.list(w, r, `SELECT id::text, category, sender_name, sender_email::text, org_badge, subject, snippet, detected_tag AS tag,
		detected_deadline AS deadline, severity, linked_entity AS link, is_read AS read, received_at
		FROM mail_messages_and_notifications ORDER BY received_at DESC LIMIT 100`)
}

type classified struct {
	Category string     `json:"category"`
	Tag      string     `json:"tag"`
	Deadline *time.Time `json:"deadline"`
	Severity string     `json:"severity"`
	OrgBadge string     `json:"org_badge"`
}

// classify asks the intel-worker and falls back to a deterministic keyword classifier.
func (s *Server) classify(ctx context.Context, from, subject, body string) classified {
	if s.Cfg.WorkerURL != "" {
		b, _ := json.Marshal(map[string]string{"sender": from, "subject": subject, "body": body})
		req, _ := http.NewRequestWithContext(ctx, "POST", s.Cfg.WorkerURL+"/worker/mail/classify", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if res, err := (&http.Client{Timeout: 4 * time.Second}).Do(req); err == nil {
			defer res.Body.Close()
			var c classified
			if res.StatusCode == 200 && json.NewDecoder(res.Body).Decode(&c) == nil && c.Category != "" {
				return c
			}
		}
	}
	text := strings.ToLower(subject + " " + body)
	c := classified{Category: "operational_alert", Severity: "normal"}
	for _, k := range []string{"rfq", "cbam xml", "sample order", "container specs", "price quote", "quotation"} {
		if strings.Contains(text, k) {
			c.Category, c.Severity = "buyer_agency", "elevated"
		}
	}
	return c
}

func (s *Server) mailInbound(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.InboundSecret == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Inbound-Secret")), []byte(s.Cfg.InboundSecret)) != 1 {
		http.Error(w, "Not Found", 404)
		return
	}
	var in struct {
		To, From, FromName, Subject, TextBody string
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	// inbox+<slug>@mail.zevra.io → org
	slug := ""
	if i, j := strings.Index(in.To, "+"), strings.Index(in.To, "@"); i >= 0 && j > i {
		slug = in.To[i+1 : j]
	}
	var org uuid.UUID
	if err := s.DB.QueryRow(r.Context(), `SELECT id FROM organizations WHERE inbound_mail_slug=$1`, slug).Scan(&org); err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	c := s.classify(r.Context(), in.From, in.Subject, in.TextBody)
	snippet := in.TextBody
	if len(snippet) > 240 {
		snippet = snippet[:240] + "…"
	}
	var id uuid.UUID
	err := rls.Tx(r.Context(), s.DB, org, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `INSERT INTO mail_messages_and_notifications (org_id, category, sender_name, sender_email, org_badge, subject, snippet, body_text, detected_tag, detected_deadline, severity)
			VALUES (current_org(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, c.Category, in.FromName, in.From, c.OrgBadge, in.Subject, snippet, in.TextBody, c.Tag, c.Deadline, c.Severity).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.Broker.Publish(org, map[string]any{"type": "notification", "notification": map[string]any{"id": id, "category": c.Category, "sender_name": in.FromName, "sender_email": in.From, "org_badge": c.OrgBadge, "subject": in.Subject, "snippet": snippet, "tag": c.Tag, "deadline": c.Deadline, "severity": c.Severity, "read": false, "received_at": time.Now()}})
	platform.JSON(w, 202, map[string]any{"id": id, "category": c.Category})
}

func (s *Server) convertRFQ(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		platform.Error(w, 400, "invalid_id")
		return
	}
	var rfq uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		var sender, tag string
		if err := tx.QueryRow(r.Context(), `SELECT sender_email::text, COALESCE(detected_tag,'') FROM mail_messages_and_notifications WHERE id=$1 AND category='buyer_agency'`, id).Scan(&sender, &tag); err != nil {
			return err
		}
		var product *uuid.UUID
		_ = tx.QueryRow(r.Context(), `SELECT id FROM products WHERE $1 LIKE '%' || cn_code || '%' LIMIT 1`, tag).Scan(&product)
		if err := tx.QueryRow(r.Context(), `INSERT INTO connect_rfqs (external_buyer, supplier_org_id, product_id, requested_qty) VALUES ($1, current_org(), $2, 1) RETURNING id`, sender, product).Scan(&rfq); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(), `UPDATE mail_messages_and_notifications SET is_read=true, linked_entity=jsonb_build_object('type','rfq','id',$2::text) WHERE id=$1`, id, rfq.String())
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"rfq_id": rfq})
}

/* ------------------------------------------------------------------ */
/* carbon & CBAM                                                       */
/* ------------------------------------------------------------------ */

var emissionFactors = map[string]struct {
	EF   float64
	Rule string
}{
	"electricity": {0.000727, "CEA-v20 / IR (EU) 2025/2621"},
	"fuel":        {0.00268, "IPCC 2006 / IR (EU) 2025/2621"},
	"precursor":   {3.9, "Supplier EPD / IR (EU) 2025/2621"},
	"process":     {1.62, "IR (EU) 2025/2621 Annex III"},
}

func (s *Server) carbonActivity(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PlantID     uuid.UUID `json:"plant_id"`
		PeriodStart string    `json:"period_start"`
		PeriodEnd   string    `json:"period_end"`
		RecordType  string    `json:"record_type"`
		Mode        string    `json:"ingestion_mode"`
		Quantity    float64   `json:"quantity"`
		Unit        string    `json:"unit"`
		EvidenceSHA string    `json:"evidence_sha256"`
		Quality     string    `json:"data_quality"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Quantity <= 0 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	ef, ok := emissionFactors[in.RecordType]
	if !ok {
		platform.Error(w, 400, "invalid_record_type")
		return
	}
	m, _ := entitlements.FromContext(r.Context())
	if (in.Mode == "ocr" && !m.Has("ocr_ingestion")) || ((in.Mode == "whatsapp" || in.Mode == "floorbot") && !m.Has("whatsapp_floorbot")) {
		platform.Error(w, 403, "capability_unavailable", map[string]any{"hint_code": "starter.ocr", "suggested_tier": "growth"})
		return
	}
	t := cbam.ActivityEmissions(in.Quantity, ef.EF)
	var id uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		var prev *string
		if m.Has("tamper_audit_locker") {
			_ = tx.QueryRow(r.Context(), `SELECT audit_hash FROM activity_records WHERE audit_hash IS NOT NULL ORDER BY created_at DESC LIMIT 1`).Scan(&prev)
		}
		rec := map[string]any{"period_start": in.PeriodStart, "record_type": in.RecordType, "quantity": in.Quantity, "unit": in.Unit, "ef": ef.EF, "tco2e": t, "evidence": in.EvidenceSHA}
		var hash *string
		if m.Has("tamper_audit_locker") {
			p := hashchain.Genesis
			if prev != nil {
				p = *prev
			}
			h := hashchain.Link(p, rec)
			hash, prev = &h, &p
		}
		return tx.QueryRow(r.Context(), `INSERT INTO activity_records (org_id, plant_id, period_start, period_end, record_type, ingestion_mode, quantity, unit, emission_factor, rule_version, computed_tco2e, data_quality, evidence_sha256, prev_audit_hash, audit_hash)
			VALUES (current_org(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE(NULLIF($11,''),'estimated'), NULLIF($12,''), $13, $14) RETURNING id`,
			in.PlantID, in.PeriodStart, in.PeriodEnd, in.RecordType, in.Mode, in.Quantity, in.Unit, ef.EF, ef.Rule, t, in.Quality, in.EvidenceSHA, prev, hash).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id, "computed_tco2e": t, "rule_version": ef.Rule})
}

func (s *Server) ocrUpload(w http.ResponseWriter, r *http.Request) {
	if s.Cfg.WorkerURL == "" {
		platform.Error(w, 503, "worker_unavailable")
		return
	}
	req, _ := http.NewRequestWithContext(r.Context(), "POST", s.Cfg.WorkerURL+"/worker/ocr/extract", io.LimitReader(r.Body, 20<<20))
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	res, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		platform.Error(w, 502, "worker_error")
		return
	}
	defer res.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}

func loadProducts(ctx context.Context, tx pgx.Tx) ([]cbam.Product, map[string]float64, error) {
	rows, err := tx.Query(ctx, `SELECT p.cn_code, p.name, COALESCE(p.direct_tco2e_t,0)::float8, COALESCE(p.indirect_tco2e_t,0)::float8, COALESCE(p.precursor_tco2e_t,0)::float8,
		COALESCE(p.default_tco2e_t,0)::float8, p.data_quality, COALESCE(SUM(b.target_qty),0)::float8
		FROM products p LEFT JOIN production_batches b ON b.product_id = p.id GROUP BY p.id ORDER BY p.sku_code`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var ps []cbam.Product
	tonnes := map[string]float64{}
	for rows.Next() {
		var p cbam.Product
		var t float64
		if err := rows.Scan(&p.CN, &p.Name, &p.DirectT, &p.IndirectT, &p.PrecursorT, &p.DefaultT, &p.DataQuality, &t); err != nil {
			return nil, nil, err
		}
		ps = append(ps, p)
		tonnes[p.CN] += t
	}
	return ps, tonnes, rows.Err()
}

func (s *Server) cbamSummary(w http.ResponseWriter, r *http.Request) {
	m, _ := entitlements.FromContext(r.Context())
	var out []map[string]any
	var s1, s2 float64
	err := s.tx(r, func(tx pgx.Tx) error {
		ps, tonnes, err := loadProducts(r.Context(), tx)
		if err != nil {
			return err
		}
		for _, p := range ps {
			if !m.Has("precursor_tracking") {
				p.PrecursorT = 0
			}
			c := cbam.CompareDefaultVsActual(tonnes[p.CN], p, cbam.CertPriceEUR)
			out = append(out, map[string]any{"cn": p.CN, "name": p.Name, "tonnes": tonnes[p.CN], "specific_tco2e_t": c.Actual, "default_tco2e_t": c.Default, "default_eur": c.DefaultEUR, "actual_eur": c.ActualEUR, "savings_eur": c.SavingsEUR, "data_quality": p.DataQuality, "label": "estimate_until_verified"})
		}
		return tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(computed_tco2e) FILTER (WHERE record_type IN ('fuel','process')),0)::float8, COALESCE(SUM(computed_tco2e) FILTER (WHERE record_type='electricity'),0)::float8 FROM activity_records`).Scan(&s1, &s2)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"rule_set": cbam.RuleSetDefinitive, "scope1_tco2e": s1, "scope2_tco2e": s2, "products": out, "logistics_reported_separately": true})
}

func (s *Server) exportXML(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Period       string `json:"period"`
		Consolidated bool   `json:"consolidated"`
		Declarant    string `json:"declarant"`
		EORI         string `json:"eori"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	m, _ := entitlements.FromContext(r.Context())
	if in.Consolidated && !m.Has("consolidated_cbam_xml") {
		platform.Error(w, 403, "capability_unavailable", map[string]any{"hint_code": "growth.consolidated_xml", "suggested_tier": "importer_hub"})
		return
	}
	var doc []byte
	err := s.tx(r, func(tx pgx.Tx) error {
		ps, tonnes, err := loadProducts(r.Context(), tx)
		if err != nil {
			return err
		}
		var plant, unloc string
		_ = tx.QueryRow(r.Context(), `SELECT name, COALESCE(un_locode,'') FROM plants ORDER BY created_at LIMIT 1`).Scan(&plant, &unloc)
		var lines []cbam.Line
		for _, p := range ps {
			if tonnes[p.CN] > 0 {
				lines = append(lines, cbam.Line{Product: p, Tonnes: tonnes[p.CN]})
			}
		}
		doc, err = cbam.RegistryXML(cbam.ReportMeta{Period: in.Period, Declarant: in.Declarant, EORI: in.EORI, Installation: plant, Operator: principal(r).Org, UNLOCODE: unloc, Country: "IN", Consolidated: in.Consolidated}, lines)
		return err
	})
	if err != nil {
		platform.Error(w, 422, "content_rule_failed", map[string]any{"detail": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", `attachment; filename="CBAM_`+in.Period+`.xml"`)
	_, _ = w.Write(doc)
}

func (s *Server) simulator(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CapexUSD, AnnualSavingsUSD, AnnualAvoidedT float64
		Years                                      int
	}
	if platform.DecodeLoose(r, &in) != nil || in.Years <= 0 || in.Years > 30 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	cbamSave := in.AnnualAvoidedT * 0.7 * cbam.CertPriceEUR / 0.86
	cum, payback := -in.CapexUSD, -1.0
	series := []float64{cum}
	for y := 1; y <= in.Years; y++ {
		prev := cum
		cum += in.AnnualSavingsUSD + cbamSave
		series = append(series, cum)
		if payback < 0 && cum >= 0 {
			payback = float64(y-1) + -prev/(in.AnnualSavingsUSD+cbamSave)
		}
	}
	platform.JSON(w, 200, map[string]any{"payback_years": payback, "cumulative_usd": series, "cbam_savings_usd_per_year": cbamSave})
}

func (s *Server) verifyAuditChain(w http.ResponseWriter, r *http.Request) {
	broken := -1
	n := 0
	err := s.tx(r, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT period_start::text, record_type, quantity::float8, unit, emission_factor::float8, computed_tco2e::float8, COALESCE(evidence_sha256,''), prev_audit_hash, audit_hash
			FROM activity_records WHERE audit_hash IS NOT NULL ORDER BY created_at`)
		if err != nil {
			return err
		}
		defer rows.Close()
		var chain []hashchain.Entry
		for rows.Next() {
			var ps, rt, unit, ev, prev, hash string
			var q, ef, t float64
			if err := rows.Scan(&ps, &rt, &q, &unit, &ef, &t, &ev, &prev, &hash); err != nil {
				return err
			}
			chain = append(chain, hashchain.Entry{Record: map[string]any{"period_start": ps, "record_type": rt, "quantity": q, "unit": unit, "ef": ef, "tco2e": t, "evidence": ev}, PrevHash: prev, Hash: hash})
		}
		n = len(chain)
		broken = hashchain.Verify(chain)
		return rows.Err()
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"entries": n, "intact": broken == -1, "broken_at": broken})
}

/* ------------------------------------------------------------------ */
/* QR traceability                                                     */
/* ------------------------------------------------------------------ */

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// NewPublicToken returns a non-guessable 26-char Crockford base32 token (130 bits from crypto/rand).
func NewPublicToken() string {
	b := make([]byte, 26)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = crockford[b[i]&31]
	}
	return string(b)
}

func (s *Server) qrBatch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		BatchID uuid.UUID `json:"batch_id"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	tok := NewPublicToken()
	var id uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `INSERT INTO qr_identities (org_id, public_token, qr_type, batch_id) VALUES (current_org(), $1, 'product_batch', $2) RETURNING id`, tok, in.BatchID).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id, "public_token": tok})
}

func (s *Server) qrContainer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code, Route string
	}
	if platform.DecodeLoose(r, &in) != nil || in.Code == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	tok := NewPublicToken()
	var id uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `INSERT INTO qr_identities (org_id, public_token, qr_type, container_code, route, dispatch_status) VALUES (current_org(), $1, 'container_shipment', $2, $3, 'Packing') RETURNING id`, tok, in.Code, in.Route).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id, "public_token": tok})
}

func (s *Server) qrLink(w http.ResponseWriter, r *http.Request) {
	cid, ok := idParam(r)
	var in struct {
		Batches []struct {
			QRID uuid.UUID `json:"qr_id"`
			Qty  float64   `json:"qty"`
		} `json:"batches"`
	}
	if !ok || platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	err := s.tx(r, func(tx pgx.Tx) error {
		for _, b := range in.Batches {
			if _, err := tx.Exec(r.Context(), `INSERT INTO container_batch_links (container_qr_id, batch_qr_id, org_id, qty_packed) VALUES ($1,$2,current_org(),$3)`, cid, b.QRID, b.Qty); err != nil {
				return err
			}
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO qr_events (org_id, qr_id, version, event) SELECT current_org(), $1, COALESCE(MAX(version),0)+1, $2 FROM qr_events WHERE qr_id=$1`, cid, "batches linked")
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"status": "linked"})
}

func (s *Server) qrAnchor(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		platform.Error(w, 400, "invalid_id")
		return
	}
	var root string
	err := s.tx(r, func(tx pgx.Tx) error {
		recs, err := rowsJSON(r.Context(), tx, `SELECT q.public_token, q.status, l.qty_packed::float8 AS qty, b.batch_code
			FROM qr_identities q LEFT JOIN container_batch_links l ON l.container_qr_id = q.id
			LEFT JOIN qr_identities bq ON bq.id = l.batch_qr_id LEFT JOIN production_batches b ON b.id = bq.batch_id WHERE q.id=$1`, id)
		if err != nil {
			return err
		}
		chain := hashchain.Build(recs)
		if len(chain) == 0 {
			return pgx.ErrNoRows
		}
		root = chain[len(chain)-1].Hash // Only the hash leaves Zevra — never business data.
		_, err = tx.Exec(r.Context(), `UPDATE qr_identities SET anchor_tx_hash=$2, anchored_at=NOW() WHERE id=$1`, id, root)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"merkle_root": root, "note": "Anchoring proves post-anchor immutability, not source-input truthfulness."})
}

func (s *Server) qrRevoke(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		platform.Error(w, 400, "invalid_id")
		return
	}
	err := s.tx(r, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `UPDATE qr_identities SET status='revoked' WHERE id=$1`, id)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"status": "revoked"})
}

// qrLabel renders printable label sheets with Typst (templates/typst/qr_labels.typ).
func (s *Server) qrLabel(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		platform.Error(w, 400, "invalid_id")
		return
	}
	var token, code string
	err := s.tx(r, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT q.public_token, COALESCE(q.container_code, b.batch_code, '') FROM qr_identities q LEFT JOIN production_batches b ON b.id=q.batch_id WHERE q.id=$1`, id).Scan(&token, &code)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	bin := s.Cfg.TypstBin
	if bin == "" {
		bin = "typst"
	}
	data, _ := json.Marshal(map[string]any{"labels": []map[string]string{{"code": code, "token": token, "url": "https://zevra.io/verify/c/" + token}}})
	cmd := exec.CommandContext(r.Context(), bin, "compile", "--input", "data="+string(data), s.Cfg.TemplatesDir+"/qr_labels.typ", "-")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		platform.Error(w, 503, "typst_unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	_, _ = w.Write(out.Bytes())
}

/* ------------------------------------------------------------------ */
/* public resolver                                                     */
/* ------------------------------------------------------------------ */

func (s *Server) publicVerify(w http.ResponseWriter, r *http.Request) {
	typ, token := chi.URLParam(r, "type"), strings.ToUpper(chi.URLParam(r, "token"))
	if (typ != "p" && typ != "c") || len(token) != 26 {
		http.Error(w, "Not Found", 404)
		return
	}
	if !s.Limiter.Allow("verify:"+platform.ClientIP(r), 120, time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	qrType := map[string]string{"p": "product_batch", "c": "container_shipment"}[typ]
	var out map[string]any
	err := rls.Tx(r.Context(), s.DB, uuid.Nil, func(tx pgx.Tx) error {
		var orgID uuid.UUID
		var status, tier string
		var anchor *string
		var anchoredAt *time.Time
		var batchID *uuid.UUID
		var container, route, dispatch *string
		if err := tx.QueryRow(r.Context(), `SELECT org_id, status, verification_tier, anchor_tx_hash, anchored_at, batch_id, container_code, route, dispatch_status FROM qr_identities WHERE public_token=$1 AND qr_type=$2`, token, qrType).
			Scan(&orgID, &status, &tier, &anchor, &anchoredAt, &batchID, &container, &route, &dispatch); err != nil {
			return err
		}
		var mName, mCity, mCountry string
		_ = tx.QueryRow(r.Context(), `SELECT o.name, COALESCE(p.city,''), o.country_code FROM organizations o LEFT JOIN plants p ON p.org_id=o.id WHERE o.id=$1 LIMIT 1`, orgID).Scan(&mName, &mCity, &mCountry)
		out = map[string]any{"type": typ, "token": token, "status": status, "verification_tier": tier,
			"manufacturer": map[string]any{"name": mName, "city": mCity, "country": mCountry, "verified": tier != "Self-Declared"},
			"anchor":       map[string]any{"anchored": anchor != nil, "anchored_at": anchoredAt, "merkle_root": anchor}}
		// Only public-approved fields are returned; partner/private data never leaves this branch.
		if typ == "c" {
			var n int
			_ = tx.QueryRow(r.Context(), `SELECT count(*) FROM container_batch_links l JOIN qr_identities q ON q.id=l.container_qr_id WHERE q.public_token=$1`, token).Scan(&n)
			out["container"] = map[string]any{"code": container, "dispatch_status": dispatch, "route": strings.Split(deref(route), " → "), "milestones": []any{}}
			out["batch_count"], out["partner_gated"] = n, true
		}
		return nil
	}, rls.PublicResolver)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	platform.JSON(w, 200, out)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

/* ------------------------------------------------------------------ */
/* FreightLink                                                         */
/* ------------------------------------------------------------------ */

func (s *Server) freightListing(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind, Origin, Dest, Corridor, Cargo string
		WindowStart, WindowEnd              time.Time
		PayloadT                            float64 `json:"payload_tonnes"`
		QuoteINR                            float64 `json:"quote_inr"`
		Port                                string  `json:"onward_port"`
	}
	if platform.DecodeLoose(r, &in) != nil || (in.Kind != "return_truck" && in.Kind != "load_demand") {
		platform.Error(w, 400, "invalid_input")
		return
	}
	m, _ := entitlements.FromContext(r.Context())
	if in.Port != "" && strings.Contains("Rotterdam Antwerp Hamburg Genoa Le Havre Valencia", in.Port) && !m.Has("freightlink_enterprise_fleet") {
		platform.Error(w, 403, "capability_unavailable", map[string]any{"hint_code": "growth.eu_port", "suggested_tier": "importer_hub"})
		return
	}
	var id uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `INSERT INTO freight_listings (org_id, listing_kind, origin_city, dest_city, corridor_code, window_start, window_end, payload_tonnes, cargo_type, quote_inr, expires_at)
			VALUES (current_org(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$6) RETURNING id`, in.Kind, in.Origin, in.Dest, in.Corridor, in.WindowStart, in.WindowEnd, in.PayloadT, in.Cargo, in.QuoteINR).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id})
}

func (s *Server) freightMatches(w http.ResponseWriter, r *http.Request) {
	loadID, err := uuid.Parse(r.URL.Query().Get("load_id"))
	if err != nil {
		platform.Error(w, 400, "load_id_required")
		return
	}
	var ranked []freight.Match
	err = s.tx(r, func(tx pgx.Tx) error {
		var l freight.Load
		if err := tx.QueryRow(r.Context(), `SELECT corridor_code, payload_tonnes::float8, window_start, window_end, COALESCE(quote_inr,0)::float8 FROM freight_listings WHERE id=$1 AND listing_kind='load_demand'`, loadID).
			Scan(&l.Corridor, &l.WeightT, &l.WindowStart, &l.WindowEnd, &l.BudgetINR); err != nil {
			return err
		}
		rows, err := tx.Query(r.Context(), `SELECT f.id::text, COALESCE(v.registration_no,''), COALESCE(v.vehicle_type,''), f.corridor_code, f.payload_tonnes::float8, f.window_start, f.window_end,
			COALESCE(f.detour_km,0)::float8, COALESCE(f.quote_inr,0)::float8, COALESCE(v.reliability_score,4)::float8, COALESCE(v.verified_status='verified' AND v.insurance_valid > NOW(), false)
			FROM freight_listings f LEFT JOIN transporter_vehicles v ON v.id = f.vehicle_id
			WHERE f.listing_kind='return_truck' AND f.status='open' AND f.expires_at > NOW()`)
		if err != nil {
			return err
		}
		defer rows.Close()
		var trucks []freight.Truck
		for rows.Next() {
			var t freight.Truck
			if err := rows.Scan(&t.ID, &t.Reg, &t.Vehicle, &t.Corridor, &t.PayloadT, &t.WindowStart, &t.WindowEnd, &t.DetourKm, &t.QuoteINR, &t.Reliability, &t.DocsValid); err != nil {
				return err
			}
			trucks = append(trucks, t)
		}
		ranked = freight.Rank(trucks, l)
		return rows.Err()
	}, rls.Marketplace)
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"matches": ranked})
}

func (s *Server) freightQuote(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TruckListingID uuid.UUID `json:"truck_listing_id"`
		LoadListingID  uuid.UUID `json:"load_listing_id"`
		Accepted       bool      `json:"accepted"`
		AuthorisedBy   string    `json:"authorised_by"`
		FreightINR     float64   `json:"freight_inr"`
		FeePct         float64   `json:"disclosed_fee_pct"`
		Baseline       float64   `json:"baseline_fuel_l"`
		Matched        float64   `json:"matched_fuel_l"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	b, err := freight.Book(freight.Match{Truck: freight.Truck{QuoteINR: in.FreightINR}}, in.Accepted, in.AuthorisedBy, in.FeePct)
	if err != nil {
		platform.Error(w, 400, err.Error())
		return
	}
	var id uuid.UUID
	err = s.tx(r, func(tx pgx.Tx) error {
		var carrier uuid.UUID
		if err := tx.QueryRow(r.Context(), `SELECT org_id FROM freight_listings WHERE id=$1`, in.TruckListingID).Scan(&carrier); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `INSERT INTO freight_bookings (shipper_org_id, transporter_org_id, truck_listing_id, load_listing_id, agreed_freight_usd, disclosed_fee_pct, accepted_by, pickup_qr_token, delivery_qr_token, baseline_fuel_l, matched_fuel_l)
			VALUES (current_org(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`, carrier, in.TruckListingID, in.LoadListingID, in.FreightINR/88, in.FeePct, b.AuthorisedBy, NewPublicToken(), NewPublicToken(), in.Baseline, in.Matched).Scan(&id); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(), `UPDATE freight_listings SET status='booked' WHERE id IN ($1,$2)`, in.TruckListingID, in.LoadListingID)
		return err
	}, rls.Marketplace)
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"booking_id": id, "booking": b, "avoided": cbam.Avoided(in.Baseline, in.Matched)})
}

func (s *Server) freightScan(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	var in struct {
		Token string `json:"token"`
		Stage string `json:"stage"` // pickup | delivery
	}
	if !ok || platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	col, next := "pickup_qr_token", "in_transit"
	if in.Stage == "delivery" {
		col, next = "delivery_qr_token", "delivered"
	}
	var updated int64
	err := s.tx(r, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(), `UPDATE freight_bookings SET status=$3 WHERE id=$1 AND `+col+`=$2`, id, in.Token, next)
		updated = tag.RowsAffected()
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if updated == 0 {
		platform.Error(w, 400, "qr_mismatch")
		return
	}
	platform.JSON(w, 200, map[string]string{"status": next})
}

func (s *Server) fleet(w http.ResponseWriter, r *http.Request) {
	s.list(w, r, `SELECT b.id::text, b.status, b.gps_consent_granted, b.baseline_fuel_l::float8, b.matched_fuel_l::float8, f.corridor_code
		FROM freight_bookings b LEFT JOIN freight_listings f ON f.id = b.truck_listing_id WHERE b.status <> 'delivered' ORDER BY b.created_at DESC`)
}

/* ------------------------------------------------------------------ */
/* Connect & Importer Hub                                              */
/* ------------------------------------------------------------------ */

func (s *Server) connectSuppliers(w http.ResponseWriter, r *http.Request) {
	q := "%" + r.URL.Query().Get("q") + "%"
	var out []map[string]any
	err := rls.Tx(r.Context(), s.DB, uuid.Nil, func(tx pgx.Tx) error {
		var err error
		out, err = rowsJSON(r.Context(), tx, `SELECT o.id::text, o.name, o.country_code, o.primary_sector FROM organizations o WHERE o.org_type='msme' AND o.name ILIKE $1 ORDER BY o.name LIMIT 50`, q)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"items": out})
}

func (s *Server) createRFQ(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SupplierOrgID uuid.UUID  `json:"supplier_org_id"`
		ProductID     *uuid.UUID `json:"product_id"`
		Qty           float64    `json:"requested_qty"`
		Incoterm      string     `json:"target_incoterm"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Qty <= 0 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	var id uuid.UUID
	err := s.tx(r, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `INSERT INTO connect_rfqs (buyer_org_id, supplier_org_id, product_id, requested_qty, target_incoterm) VALUES (current_org(), $1, $2, $3, $4) RETURNING id`, in.SupplierOrgID, in.ProductID, in.Qty, in.Incoterm).Scan(&id)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"id": id})
}

func (s *Server) bulkInvite(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Emails       []string `json:"emails"`
		ConfirmExtra bool     `json:"confirm_extra_seats"`
	}
	if platform.DecodeLoose(r, &in) != nil || len(in.Emails) == 0 || len(in.Emails) > 500 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	m, _ := entitlements.FromContext(r.Context())
	var links []map[string]string
	err := s.tx(r, func(tx pgx.Tx) error {
		var active int
		if err := tx.QueryRow(r.Context(), `SELECT count(*) FROM importer_supplier_links`).Scan(&active); err != nil {
			return err
		}
		extra := active + len(in.Emails) - m.Limits.IncludedSuppliers
		if extra > 0 && !in.ConfirmExtra {
			return &seatError{extra: extra}
		}
		for _, e := range in.Emails {
			tok := platform.RandomToken(24)
			if _, err := tx.Exec(r.Context(), `INSERT INTO importer_supplier_links (importer_org_id, invited_email, invite_token_hash) VALUES (current_org(), $1, $2)`, strings.ToLower(e), platform.SHA256(tok)); err != nil {
				return err
			}
			links = append(links, map[string]string{"email": e, "invite_url": "https://zevra.io/signup?invite=" + tok})
		}
		if extra > 0 {
			_, err := tx.Exec(r.Context(), `UPDATE organizations SET extra_suppliers = GREATEST(extra_suppliers, $1) WHERE id = current_org()`, extra)
			return err
		}
		return nil
	})
	var se *seatError
	if errors.As(err, &se) {
		platform.Error(w, 402, "extra_seats_confirmation_required", map[string]any{"extra_seats": se.extra, "usd_per_seat_month": m.Limits.ExtraSupplierUSD})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"invites": links})
}

type seatError struct{ extra int }

func (e *seatError) Error() string { return "extra seats require confirmation" }

func (s *Server) benchmarks(w http.ResponseWriter, r *http.Request) {
	s.list(w, r, `SELECT COALESCE(supplier_name, invited_email::text) AS supplier, cn_code, intensity_tco2e_t::float8 AS intensity, imported_tonnes::float8 AS tonnes,
		(intensity_tco2e_t * imported_tonnes * 76)::float8 AS exposure_eur, submission_status
		FROM importer_supplier_links WHERE intensity_tco2e_t IS NOT NULL ORDER BY intensity_tco2e_t`)
}

/* ------------------------------------------------------------------ */
/* Ops vault (PGP super-admin)                                         */
/* ------------------------------------------------------------------ */

func (s *Server) opsMargins(w http.ResponseWriter, r *http.Request) {
	platform.JSON(w, 200, map[string]any{"tiers": []map[string]any{
		{"tier": "starter", "price_usd": 99, "direct_cost_ceiling_usd": 19.80, "target_gm_pct": 80, "annual_ceiling_usd": 16.50},
		{"tier": "growth", "price_usd": 249, "direct_cost_ceiling_usd": 49.80, "target_gm_pct": 80, "annual_ceiling_usd": 41.50},
		{"tier": "importer_hub", "price_usd": 499, "direct_cost_ceiling_usd": 74.85, "target_gm_pct": 85, "annual_ceiling_usd": 62.38},
	}})
}

func (s *Server) opsKPIs(w http.ResponseWriter, r *http.Request) {
	var out map[string]any
	err := rls.Tx(r.Context(), s.DB, uuid.Nil, func(tx pgx.Tx) error {
		var bookings, delivered int
		var base, matched float64
		_ = tx.QueryRow(r.Context(), `SELECT count(*), count(*) FILTER (WHERE status='delivered'), COALESCE(SUM(baseline_fuel_l),0)::float8, COALESCE(SUM(matched_fuel_l),0)::float8 FROM freight_bookings`).Scan(&bookings, &delivered, &base, &matched)
		out = map[string]any{"bookings": bookings, "delivered": delivered, "avoided": cbam.Avoided(base, matched)}
		return nil
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, out)
}

func (s *Server) opsPublishRule(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Version, EffectiveOn, Note string
		DefaultValues              map[string]float64 `json:"default_values"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Version == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	dv, _ := json.Marshal(in.DefaultValues)
	if _, err := s.DB.Exec(r.Context(), `INSERT INTO cbam_rule_sets (version, effective_on, default_values, note) VALUES ($1,$2,$3,$4)`, in.Version, in.EffectiveOn, dv, in.Note); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]string{"status": "scheduled"})
}

func (s *Server) opsApprove(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	var in struct {
		Decision string `json:"decision"`
	}
	if !ok || platform.DecodeLoose(r, &in) != nil || (in.Decision != "verified" && in.Decision != "suspended") {
		platform.Error(w, 400, "invalid_input")
		return
	}
	if _, err := s.DB.Exec(r.Context(), `UPDATE transporter_vehicles SET verified_status=$2 WHERE id=$1`, id, in.Decision); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]string{"status": in.Decision})
}
