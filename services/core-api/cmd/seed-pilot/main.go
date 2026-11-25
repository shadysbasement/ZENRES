// Command seed-pilot populates a complete Indian aluminium MSME pilot — "Deccan Alloy Works" — as three
// demo organisations (Starter, Growth, EU Importer Hub) so every module opens with zero empty states.
//
//	DATABASE_URL=postgres://zevra:zevra@localhost:5432/zevra?sslmode=disable go run ./core-api/cmd/seed-pilot
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"

	"github.com/shadysbasement/zevra/services/core-api/internal/rls"
)

const demoPassword = "Zevra-Pilot-2026!"

func hash(pw string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	k := argon2.IDKey([]byte(pw), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=3,p=4$%s$%s", argon2.Version, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(k))
}

type zone struct {
	code, name, typ, supervisor string
	pos, size                   [3]float64
	active, target              int
	kw, fuel                    float64
}

var zones = []zone{
	{"PROD-A", "Line A · Smelting & Casting", "production", "Rakesh Patil", [3]float64{-17, 0, -8}, [3]float64{24, 9, 16}, 11, 12, 612, 38},
	{"PROD-B", "Line B · Machining & Finishing", "production", "Sunita Kale", [3]float64{9, 0, -8}, [3]float64{20, 7, 16}, 7, 8, 284, 0},
	{"PACK-1", "Packaging & Dispatch Bay", "packaging", "Imran Shaikh", [3]float64{9, 0, 11}, [3]float64{20, 6, 14}, 9, 10, 46, 0},
	{"OFF-ADMIN", "Offices & Admin Wing", "office", "Meera Joshi", [3]float64{-23, 0, 11}, [3]float64{12, 8, 14}, 6, 6, 18, 0},
	{"WH-MAIN", "Raw Material & FG Warehouse", "warehouse", "Vikram Rao", [3]float64{-8, 0, 11}, [3]float64{12, 8, 14}, 3, 4, 9, 2},
}

type product struct {
	sku, cn, name                    string
	cost, direct, indirect, pre, def float64
	quality                          string
}

var products = []product{
	{"ALB-6063", "76042910", "Aluminium bars 6063-T6", 2860, 0.62, 1.71, 3.17, 7.9, "measured"},
	{"ALP-6061", "76042100", "Extruded hollow profiles 6061", 3120, 0.66, 1.82, 3.41, 7.9, "supplier_reported"},
	{"ALW-1350", "76051100", "Aluminium wire EC grade 1350", 2780, 0.41, 1.55, 4.28, 8.4, "estimated"},
	{"ALS-5052", "76061220", "Alloy sheets 5052-H32", 3340, 0.71, 2.04, 3.51, 8.1, "measured"},
}

type batch struct {
	code, sku, zone, stage, quality, token string
	target, produced, packaged             float64
}

var batches = []batch{
	{"A-101", "ALB-6063", "PACK-1", "ready_dispatch", "measured", "01JD8ZA101BTCH3N6R1VQ8MZTE", 6, 6, 6},
	{"A-102", "ALB-6063", "PACK-1", "ready_dispatch", "measured", "01JD8ZA102BTCH9H2K7WD4FXRC", 5.5, 5.5, 5.5},
	{"A-103", "ALP-6061", "PACK-1", "ready_dispatch", "supplier_reported", "01JD8ZA103BTCH5T8P3LJ6NYGS", 6.5, 6.5, 6.5},
	{"A-104", "ALB-6063", "PROD-A", "in_production", "measured", "01JD8ZA104BTCH7Y2M5QK8RWDN", 8, 4.2, 0},
}

var emails = []struct{ cat, name, from, badge, subject, snippet, tag, sev string }{
	{"buyer_agency", "Lena Hoffmann", "l.hoffmann@rhein-alu-import.de", "Rhein Aluminium Import · EU Importer", "RFQ — 40 t 6063-T6 bars, Q1 2027 + CBAM XML", "Please quote 40 t of 6063-T6 bars for Q1 delivery to Duisburg, with your CBAM communication template.", "CN 76042910", "elevated"},
	{"buyer_agency", "Arjun Mehta", "arjun@euro-sourcing-agents.com", "Euro Sourcing Agents · Buying house", "Sample order — extruded profiles for Lyon client", "Our French client would like a 500 kg sample of 6061 hollow profiles with container specs and your QR passport link.", "CN 76042100", "normal"},
	{"buyer_agency", "Mikkel Sørensen", "ms@nordhavn-metals.dk", "Nordhavn Metals · Distributor", "Price quote request: EC grade wire 1350", "Could you share FOB Mundra pricing for 9 t of 1350 wire and confirm whether emissions data is measured?", "CN 76051100", "normal"},
	{"deadline", "EU CBAM calendar", "calendar@zevra.io", "", "Q4 embedded-emissions data due to Rhein Aluminium Import", "Your importer's declaration window requires supplier emissions data for all Q4 shipments.", "", "elevated"},
	{"deadline", "Evidence vault", "vault@zevra.io", "", "October electricity bill (MSEDCL) missing", "Scope 2 for October is currently estimated. Upload the HT bill to switch to measured.", "", "critical"},
	{"operational_alert", "FreightLink", "freightlink@zevra.io", "", "Return load matched — Mumbai → Delhi, 14 t", "MH-04-KT-7781 matched with Bhiwandi machine parts. Quote ₹38,500 awaiting review.", "", "normal"},
}

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	must(err)
	defer pool.Close()

	type demo struct{ tier, orgType, name, email, slug string }
	demos := []demo{
		{"starter", "msme", "Deccan Alloy Works (Starter demo)", "starter@deccanalloy.demo", "deccan-starter"},
		{"growth", "msme", "Deccan Alloy Works Pvt. Ltd.", "growth@deccanalloy.demo", "deccan-alloy"},
		{"importer_hub", "eu_importer", "Rhein Aluminium Import GmbH", "importer@rhein-alu.demo", "rhein-alu"},
	}
	for _, d := range demos {
		var orgID, userID uuid.UUID
		must(pool.QueryRow(ctx, `INSERT INTO organizations (name, org_type, country_code, gstin_or_eori, primary_sector, subscription_tier, inbound_mail_slug)
			VALUES ($1,$2,$3,$4,'Aluminium',$5,$6) ON CONFLICT (inbound_mail_slug) DO UPDATE SET name=EXCLUDED.name RETURNING id`,
			d.name, d.orgType, map[bool]string{true: "DE", false: "IN"}[d.tier == "importer_hub"], map[bool]string{true: "DE472918350000", false: "27AAACD4821K1Z7"}[d.tier == "importer_hub"], d.tier, d.slug).Scan(&orgID))
		must(pool.QueryRow(ctx, `INSERT INTO users (email, email_verified, full_name, password_hash, auth_origin) VALUES ($1,true,'Ananya Deshpande',$2,'manual')
			ON CONFLICT (email) DO UPDATE SET password_hash=EXCLUDED.password_hash RETURNING id`, d.email, hash(demoPassword)).Scan(&userID))
		_, err := pool.Exec(ctx, `INSERT INTO org_memberships (org_id, user_id, role) VALUES ($1,$2,'owner') ON CONFLICT DO NOTHING`, orgID, userID)
		must(err)
		must(rls.Tx(ctx, pool, orgID, func(tx pgx.Tx) error { return seedOrg(ctx, tx, d.tier) }))
		fmt.Printf("✓ %-14s %-40s login: %s / %s\n", d.tier, d.name, d.email, demoPassword)
	}
	fmt.Println("\nPilot seeded. Sign in at http://localhost:5173/login")
}

func seedOrg(ctx context.Context, tx pgx.Tx, tier string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM plants)`).Scan(&exists); err != nil || exists {
		return err
	}
	var plant uuid.UUID
	if err := tx.QueryRow(ctx, `INSERT INTO plants (org_id, name, city, state_region, un_locode, grid_region) VALUES (current_org(),'Chakan Extrusion Plant','Chakan MIDC, Pune','Maharashtra','INPNQ','IN-WEST') RETURNING id`).Scan(&plant); err != nil {
		return err
	}
	zoneIDs := map[string]uuid.UUID{}
	for _, z := range zones {
		var id uuid.UUID
		geo := fmt.Sprintf(`{"pos":[%g,%g,%g],"size":[%g,%g,%g]}`, z.pos[0], z.pos[1], z.pos[2], z.size[0], z.size[1], z.size[2])
		if err := tx.QueryRow(ctx, `INSERT INTO factory_zones (plant_id, org_id, zone_code, name, zone_type, geometry_3d, supervisor_name, active_workers, shift_target_workers, live_power_kw, fuel_lph)
			VALUES ($1, current_org(), $2,$3,$4,$5::jsonb,$6,$7,$8,$9,$10) RETURNING id`, plant, z.code, z.name, z.typ, geo, z.supervisor, z.active, z.target, z.kw, z.fuel).Scan(&id); err != nil {
			return err
		}
		zoneIDs[z.code] = id
	}
	prodIDs := map[string]uuid.UUID{}
	for _, p := range products {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `INSERT INTO products (org_id, sku_code, cn_code, name, sector, unit_cost_usd, direct_tco2e_t, indirect_tco2e_t, precursor_tco2e_t, default_tco2e_t, data_quality, published_connect)
			VALUES (current_org(), $1,$2,$3,'Aluminium',$4,$5,$6,$7,$8,$9,true) RETURNING id`, p.sku, p.cn, p.name, p.cost, p.direct, p.indirect, p.pre, p.def, p.quality).Scan(&id); err != nil {
			return err
		}
		prodIDs[p.sku] = id
	}
	var container uuid.UUID
	ctok := "01JD8ZV001CNTR7Q4M2X9KPWBA"
	if tier != "growth" { // tokens are globally unique; only the flagship demo carries the canonical ZV-001 passports
		ctok = ""
	}
	if ctok != "" {
		if err := tx.QueryRow(ctx, `INSERT INTO qr_identities (org_id, public_token, qr_type, container_code, route, dispatch_status, verification_tier, anchor_tx_hash, anchored_at)
			VALUES (current_org(), $1, 'container_shipment', 'ZV-001', 'Pune → JNPT Nhava Sheva → Rotterdam', 'Ready to dispatch', 'Document-Reviewed', '9f2c4be1d07a3e55c1f8a2b64d9e0c7731aa58e2f6b4d1c09e3a7f25b8c6e41a', '2026-11-18T09:42:00Z') RETURNING id`, ctok).Scan(&container); err != nil {
			return err
		}
	}
	for _, b := range batches {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `INSERT INTO production_batches (org_id, plant_id, zone_id, product_id, batch_code, stage, target_qty, produced_qty, packaged_qty, direct_tco2e, indirect_tco2e, data_quality, produced_on)
			VALUES (current_org(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, NOW()::date) RETURNING id`, plant, zoneIDs[b.zone], prodIDs[b.sku], b.code, b.stage, b.target, b.produced, b.packaged, b.produced*0.62, b.produced*1.71, b.quality).Scan(&id); err != nil {
			return err
		}
		if ctok != "" {
			var qr uuid.UUID
			if err := tx.QueryRow(ctx, `INSERT INTO qr_identities (org_id, public_token, qr_type, batch_id, verification_tier) VALUES (current_org(), $1, 'product_batch', $2, 'Document-Reviewed') RETURNING id`, b.token, id).Scan(&qr); err != nil {
				return err
			}
			if b.stage == "ready_dispatch" {
				if _, err := tx.Exec(ctx, `INSERT INTO container_batch_links (container_qr_id, batch_qr_id, org_id, qty_packed) VALUES ($1,$2,current_org(),$3)`, container, qr, b.packaged); err != nil {
					return err
				}
			}
		}
	}
	for _, it := range []struct {
		code, name, cat, unit string
		qty, reorder, value   float64
	}{
		{"RM-ALI-P1020", "Primary aluminium ingot P1020", "raw_material", "t", 86.4, 60, 2410},
		{"RM-SCRAP-PC", "Post-consumer aluminium scrap", "raw_material", "t", 21.7, 25, 1540},
		{"WIP-BIL-6063", "Cast billets 6063 (WIP)", "wip", "t", 14.2, 0, 2550},
		{"FG-ALB-6063", "Aluminium bars 6063-T6", "finished_good", "t", 18.6, 8, 2860},
		{"CON-PAL-EU", "Heat-treated EUR pallets", "consumable", "pcs", 140, 120, 14},
	} {
		if _, err := tx.Exec(ctx, `INSERT INTO inventory_items (org_id, plant_id, item_code, name, category, qty_on_hand, unit, reorder_level, unit_value_usd) VALUES (current_org(), $1,$2,$3,$4,$5,$6,$7,$8)`, plant, it.code, it.name, it.cat, it.qty, it.unit, it.reorder, it.value); err != nil {
			return err
		}
	}
	// 12 months of Tally vouchers (deterministic).
	buyers := []string{"Rhein Aluminium Import GmbH", "Nordhavn Metals ApS", "Lyon Profilés SAS", "Bharat Fabricators", "Kolhapur Auto Components", "Valencia Extrusiones SL"}
	start := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	for m := 0; m < 12; m++ {
		month := start.AddDate(0, m, 0)
		for k := 0; k < 9; k++ {
			b := buyers[(m*7+k*3)%len(buyers)]
			amt := math.Round((14000 + float64((m*977+k*1531)%22000)) * (1 + float64(m)*0.035))
			if _, err := tx.Exec(ctx, `INSERT INTO financial_vouchers (org_id, source, external_ref, voucher_type, counterparty, voucher_date, amount_usd, cost_usd, is_export) VALUES (current_org(),'tally_xml',$1,'sales',$2,$3,$4,$5,$6)`,
				fmt.Sprintf("SAL/%s/%04d", month.Format("200601"), 400+k), b, month.AddDate(0, 0, 2+k*3), amt, amt*0.7, b != "Bharat Fabricators" && b != "Kolhapur Auto Components"); err != nil {
				return err
			}
		}
		for _, v := range []struct {
			t, party string
			amt      float64
		}{{"purchase", "Hindalco Renukoot", 31000}, {"utility", "MSEDCL", 17500}, {"freight", "Shree Ganesh Roadlines", 6100}, {"expense", "Payroll & overheads", 10400}} {
			if _, err := tx.Exec(ctx, `INSERT INTO financial_vouchers (org_id, source, external_ref, voucher_type, counterparty, voucher_date, amount_usd, cost_usd) VALUES (current_org(),'tally_xml',$1,$2,$3,$4,$5,$5)`,
				fmt.Sprintf("%s/%s", v.t, month.Format("200601")), v.t, v.party, month.AddDate(0, 0, 15), v.amt*(1+float64(m)*0.02)); err != nil {
				return err
			}
		}
	}
	for i, e := range emails {
		var deadline *time.Time
		if e.cat == "deadline" {
			d := time.Now().AddDate(0, 0, 4-i*2)
			deadline = &d
		}
		if _, err := tx.Exec(ctx, `INSERT INTO mail_messages_and_notifications (org_id, category, sender_name, sender_email, org_badge, subject, snippet, detected_tag, detected_deadline, severity, received_at)
			VALUES (current_org(), $1,$2,$3,NULLIF($4,''),$5,$6,NULLIF($7,''),$8,$9, NOW() - ($10 || ' minutes')::interval)`, e.cat, e.name, e.from, e.badge, e.subject, e.snippet, e.tag, deadline, e.sev, fmt.Sprint(20+i*90)); err != nil {
			return err
		}
	}
	for _, a := range []struct {
		typ, mode, unit, rule, quality string
		qty, ef                        float64
	}{
		{"electricity", "tally_csv", "kWh", "CEA-v20 / IR (EU) 2025/2621", "measured", 398100, 0.000727},
		{"fuel", "manual", "L diesel", "IPCC 2006 / IR (EU) 2025/2621", "measured", 28400, 0.00268},
		{"precursor", "manual", "t P1020 ingot", "Supplier EPD / IR (EU) 2025/2621", "supplier_reported", 118, 3.9},
	} {
		if _, err := tx.Exec(ctx, `INSERT INTO activity_records (org_id, plant_id, period_start, period_end, record_type, ingestion_mode, quantity, unit, emission_factor, rule_version, computed_tco2e, data_quality)
			VALUES (current_org(), $1, '2026-10-01', '2026-10-31', $2,$3,$4,$5,$6,$7,$8,$9)`, plant, a.typ, a.mode, a.qty, a.unit, a.ef, a.rule, math.Round(a.qty*a.ef*10000)/10000, a.quality); err != nil {
			return err
		}
	}
	// FreightLink corridors: Delhi–Mumbai and Ludhiana–Delhi return trucks + load demand.
	for _, l := range []struct {
		kind, o, d, corr, cargo string
		payload, detour, quote  float64
		startH, endH            int
	}{
		{"return_truck", "Mumbai", "Delhi", "BOM-DEL", "", 21, 18, 38500, 24, 60},
		{"return_truck", "Mumbai", "Delhi", "BOM-DEL", "", 16, 42, 35200, 30, 72},
		{"return_truck", "Delhi", "Ludhiana", "DEL-LDH", "", 10, 6, 14800, 8, 26},
		{"load_demand", "Mumbai", "Delhi", "BOM-DEL", "Packaged machine parts", 14, 0, 40000, 26, 58},
	} {
		if _, err := tx.Exec(ctx, `INSERT INTO freight_listings (org_id, listing_kind, origin_city, dest_city, corridor_code, window_start, window_end, payload_tonnes, cargo_type, detour_km, quote_inr, expires_at)
			VALUES (current_org(), $1,$2,$3,$4, NOW() + ($5 || ' hours')::interval, NOW() + ($6 || ' hours')::interval, $7, NULLIF($8,''), $9, $10, NOW() + ($6 || ' hours')::interval)`,
			l.kind, l.o, l.d, l.corr, fmt.Sprint(l.startH), fmt.Sprint(l.endH), l.payload, l.cargo, l.detour, l.quote); err != nil {
			return err
		}
	}
	if tier == "importer_hub" {
		for i, s := range []struct {
			name, cn, status string
			intensity, t     float64
		}{
			{"Deccan Alloy Works", "76042910", "verified", 5.5, 18}, {"Vizag Aluminium Foils", "76071190", "submitted", 6.8, 9.2},
			{"Gujarat Steel Tubes", "73063019", "verified", 2.1, 64}, {"Coimbatore Castings", "73251000", "pending_evidence", 1.9, 22.5},
			{"Jindal Wire Products", "72172000", "submitted", 2.4, 31}, {"Bhilai Structurals", "72163100", "verified", 2.6, 55},
		} {
			tok := make([]byte, 24)
			_, _ = rand.Read(tok)
			if _, err := tx.Exec(ctx, `INSERT INTO importer_supplier_links (importer_org_id, supplier_name, invited_email, invite_token_hash, cn_code, intensity_tco2e_t, imported_tonnes, submission_status)
				VALUES (current_org(), $1, $2, $3, $4, $5, $6, $7)`, s.name, fmt.Sprintf("compliance%d@supplier.demo", i+1), tok, s.cn, s.intensity, s.t, s.status); err != nil {
				return err
			}
		}
	}
	return nil
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed-pilot:", err)
		os.Exit(1)
	}
}
