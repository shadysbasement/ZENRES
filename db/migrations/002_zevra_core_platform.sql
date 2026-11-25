-- ZEVRA // 002 — core platform with Row-Level Security (PostgreSQL 17)
-- Every tenant-scoped table enforces: org_id = current_setting('app.current_org_id')::uuid.
-- The Go core API runs `SELECT set_config('app.current_org_id', $1, true)` (SET LOCAL) inside each transaction.

-- 1. ORGANIZATIONS, SUBSCRIPTIONS & SILENT ENTITLEMENT GOVERNANCE
CREATE TABLE organizations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name               TEXT NOT NULL,
    org_type           TEXT NOT NULL CHECK (org_type IN ('msme','eu_importer','transporter','buyer','intermediary')),
    country_code       CHAR(2) NOT NULL DEFAULT 'IN',
    gstin_or_eori      TEXT,
    primary_sector     TEXT,
    subscription_tier  TEXT NOT NULL DEFAULT 'starter' CHECK (subscription_tier IN ('starter','growth','importer_hub')),
    billing_cycle      TEXT NOT NULL DEFAULT 'monthly' CHECK (billing_cycle IN ('monthly','annual')),
    extra_suppliers    INT NOT NULL DEFAULT 0,
    inbound_mail_slug  TEXT NOT NULL UNIQUE, -- inbox+<slug>@mail.zevra.io
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE org_memberships (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role    TEXT NOT NULL CHECK (role IN ('owner','finance','production','compliance','logistics_partner','buyer','verifier')),
    UNIQUE (org_id, user_id)
);

CREATE TABLE user_hint_dismissals (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hint_code    TEXT NOT NULL,
    dismissed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, hint_code)
);

-- 2. PLANTS, 3D FACTORY ZONES & LIVE WORKFORCE / SKU TELEMETRY
CREATE TABLE plants (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    city           TEXT NOT NULL,
    state_region   TEXT,
    un_locode      TEXT,
    grid_region    TEXT NOT NULL DEFAULT 'IN-NORTH',
    floor_dims_m   JSONB NOT NULL DEFAULT '{"width":60,"depth":40,"height":12}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE factory_zones (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plant_id            UUID NOT NULL REFERENCES plants(id) ON DELETE CASCADE,
    org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    zone_code           TEXT NOT NULL, -- 'PROD-A', 'PACK-1', 'OFF-ADMIN', 'WH-MAIN'
    name                TEXT NOT NULL,
    zone_type           TEXT NOT NULL CHECK (zone_type IN ('production','packaging','office','warehouse')),
    geometry_3d         JSONB NOT NULL, -- {pos:[x,y,z], size:[w,h,d], color_token, equipment_preset}
    supervisor_name     TEXT,
    active_workers      INT NOT NULL DEFAULT 0,
    shift_target_workers INT NOT NULL DEFAULT 0,
    live_power_kw       NUMERIC(10,2) NOT NULL DEFAULT 0,
    fuel_lph            NUMERIC(10,2) NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plant_id, zone_code)
);

-- 3. PRODUCTS (CN-CODE SKUS), BOM, BATCHES & INVENTORY
CREATE TABLE products (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    sku_code         TEXT NOT NULL,
    cn_code          TEXT NOT NULL CHECK (cn_code ~ '^[0-9]{8}$'),
    name             TEXT NOT NULL,
    sector           TEXT NOT NULL,
    unit             TEXT NOT NULL DEFAULT 'tonne',
    bom_json         JSONB NOT NULL DEFAULT '[]'::jsonb,
    unit_cost_usd    NUMERIC(12,2),
    direct_tco2e_t   NUMERIC(10,4),
    indirect_tco2e_t NUMERIC(10,4),
    precursor_tco2e_t NUMERIC(10,4),
    default_tco2e_t  NUMERIC(10,4),
    data_quality     TEXT NOT NULL DEFAULT 'estimated' CHECK (data_quality IN ('measured','supplier_reported','estimated','verified')),
    published_connect BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (org_id, sku_code)
);

CREATE TABLE production_batches (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plant_id         UUID NOT NULL REFERENCES plants(id),
    zone_id          UUID REFERENCES factory_zones(id),
    product_id       UUID NOT NULL REFERENCES products(id),
    batch_code       TEXT NOT NULL,
    stage            TEXT NOT NULL CHECK (stage IN ('scheduled','in_production','packaging','ready_dispatch','shipped')),
    target_qty       NUMERIC(12,3) NOT NULL,
    produced_qty     NUMERIC(12,3) NOT NULL DEFAULT 0,
    packaged_qty     NUMERIC(12,3) NOT NULL DEFAULT 0,
    direct_tco2e     NUMERIC(12,4),
    indirect_tco2e   NUMERIC(12,4),
    data_quality     TEXT NOT NULL DEFAULT 'estimated' CHECK (data_quality IN ('measured','supplier_reported','estimated','verified')),
    produced_on      DATE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, batch_code)
);

CREATE TABLE inventory_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plant_id       UUID NOT NULL REFERENCES plants(id),
    item_code      TEXT NOT NULL,
    name           TEXT NOT NULL,
    category       TEXT NOT NULL CHECK (category IN ('raw_material','wip','finished_good','consumable')),
    qty_on_hand    NUMERIC(14,3) NOT NULL DEFAULT 0,
    unit           TEXT NOT NULL DEFAULT 't',
    reorder_level  NUMERIC(14,3) NOT NULL DEFAULT 0,
    unit_value_usd NUMERIC(12,2) NOT NULL DEFAULT 0,
    tally_stock_id TEXT,
    lot_ref        TEXT,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, item_code)
);

CREATE TABLE stock_movements (
    id          BIGSERIAL PRIMARY KEY,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    item_id     UUID NOT NULL REFERENCES inventory_items(id),
    kind        TEXT NOT NULL CHECK (kind IN ('goods_received','issued_to_line','packed_into_batch','dispatched','adjustment')),
    qty         NUMERIC(14,3) NOT NULL,
    reference   TEXT,
    actor       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. FINANCIAL LEDGER & TALLY.ERP 9 SYNC
CREATE TABLE financial_vouchers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source          TEXT NOT NULL CHECK (source IN ('manual','tally_csv','tally_xml','tally_odbc','freightlink')),
    external_ref    TEXT, -- Stable Tally voucher GUID for idempotency
    voucher_type    TEXT NOT NULL CHECK (voucher_type IN ('sales','purchase','expense','utility','freight')),
    counterparty    TEXT NOT NULL,
    voucher_date    DATE NOT NULL,
    amount_usd      NUMERIC(14,2) NOT NULL,
    cost_usd        NUMERIC(14,2) DEFAULT 0,
    is_export       BOOLEAN NOT NULL DEFAULT FALSE,
    raw_payload     JSONB,
    UNIQUE (org_id, source, external_ref)
);

CREATE TABLE tally_sync_log (
    id          BIGSERIAL PRIMARY KEY,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source      TEXT NOT NULL,
    inserted    INT NOT NULL DEFAULT 0,
    duplicates  INT NOT NULL DEFAULT 0,
    errors      INT NOT NULL DEFAULT 0,
    detail      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. CONNECTED MAILING SERVICE & NOTIFICATION BELL
CREATE TABLE mail_connections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL CHECK (provider IN ('gmail_oauth','ms_graph','imap_smtp','zevra_inbound')),
    account_email   CITEXT NOT NULL,
    credentials_enc BYTEA, -- KMS-encrypted OAuth tokens or IMAP creds
    status          TEXT NOT NULL DEFAULT 'active',
    last_synced_at  TIMESTAMPTZ
);

CREATE TABLE mail_messages_and_notifications (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    category         TEXT NOT NULL CHECK (category IN ('buyer_agency','deadline','operational_alert')),
    sender_name      TEXT,
    sender_email     CITEXT,
    org_badge        TEXT,
    subject          TEXT NOT NULL,
    snippet          TEXT NOT NULL,
    body_text        TEXT,
    detected_tag     TEXT,
    detected_deadline TIMESTAMPTZ,
    severity         TEXT NOT NULL DEFAULT 'normal' CHECK (severity IN ('normal','elevated','critical')),
    linked_entity    JSONB, -- {type:'rfq'|'shipment'|'cbam_report', id:'...'}
    is_read          BOOLEAN NOT NULL DEFAULT FALSE,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. CARBON, CBAM EVIDENCE LOCKER & HASH-CHAINED AUDIT TRAIL
CREATE TABLE activity_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plant_id        UUID NOT NULL REFERENCES plants(id),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    record_type     TEXT NOT NULL CHECK (record_type IN ('electricity','fuel','precursor','process')),
    ingestion_mode  TEXT NOT NULL CHECK (ingestion_mode IN ('manual','tally_csv','ocr','whatsapp','floorbot')),
    quantity        NUMERIC(14,4) NOT NULL,
    unit            TEXT NOT NULL,
    emission_factor NUMERIC(12,6) NOT NULL,
    rule_version    TEXT NOT NULL,
    computed_tco2e  NUMERIC(14,4) NOT NULL,
    data_quality    TEXT NOT NULL DEFAULT 'estimated' CHECK (data_quality IN ('measured','supplier_reported','estimated','verified')),
    evidence_s3_key TEXT,
    evidence_sha256 CHAR(64),
    prev_audit_hash CHAR(64), -- Tamper-evident hash chain (Growth+)
    audit_hash      CHAR(64),
    verified_status TEXT NOT NULL DEFAULT 'unverified',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbam_rule_sets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version       TEXT NOT NULL UNIQUE,
    effective_on  DATE NOT NULL,
    status        TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','live','archived')),
    default_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    note          TEXT,
    published_by  UUID REFERENCES admin_pgp_keys(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. DUAL QR TRACEABILITY & BLOCKCHAIN ANCHORING
CREATE TABLE qr_identities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    public_token    TEXT NOT NULL UNIQUE CHECK (public_token ~ '^[0-9A-Z]{26}$'), -- Non-guessable 26-char ULID/base32
    qr_type         TEXT NOT NULL CHECK (qr_type IN ('product_batch','container_shipment')),
    batch_id        UUID REFERENCES production_batches(id),
    container_code  TEXT, -- e.g. 'ZV-001'
    dispatch_status TEXT,
    route           TEXT,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','reprinted','revoked')),
    verification_tier TEXT NOT NULL DEFAULT 'Self-Declared',
    anchor_tx_hash  TEXT, -- Optional off-chain hash anchor reference
    anchored_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE container_batch_links (
    container_qr_id UUID NOT NULL REFERENCES qr_identities(id) ON DELETE CASCADE,
    batch_qr_id     UUID NOT NULL REFERENCES qr_identities(id) ON DELETE CASCADE,
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    qty_packed      NUMERIC(12,3) NOT NULL,
    PRIMARY KEY (container_qr_id, batch_qr_id)
);

CREATE TABLE qr_events (
    id          BIGSERIAL PRIMARY KEY,
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    qr_id       UUID NOT NULL REFERENCES qr_identities(id) ON DELETE CASCADE,
    version     INT NOT NULL,
    event       TEXT NOT NULL,
    supersedes  BIGINT REFERENCES qr_events(id), -- versioned corrections, never silent overwrites
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. FREIGHTLINK BACKHAUL MARKETPLACE, QUOTES, SHIPMENTS & DISPUTES
CREATE TABLE transporter_vehicles (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transporter_org   UUID NOT NULL REFERENCES organizations(id),
    registration_no   TEXT NOT NULL UNIQUE,
    vehicle_type      TEXT NOT NULL,
    payload_tonnes    NUMERIC(8,2) NOT NULL,
    insurance_valid   DATE NOT NULL,
    permit_type       TEXT,
    verified_status   TEXT NOT NULL DEFAULT 'pending' CHECK (verified_status IN ('pending','verified','suspended')),
    reliability_score NUMERIC(4,2) DEFAULT 5.00
);

CREATE TABLE freight_listings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    listing_kind    TEXT NOT NULL CHECK (listing_kind IN ('return_truck','load_demand')),
    vehicle_id      UUID REFERENCES transporter_vehicles(id),
    origin_city     TEXT NOT NULL,
    dest_city       TEXT NOT NULL,
    corridor_code   TEXT NOT NULL, -- e.g. 'BOM-DEL', 'DEL-LDH'
    window_start    TIMESTAMPTZ NOT NULL,
    window_end      TIMESTAMPTZ NOT NULL,
    payload_tonnes  NUMERIC(8,2) NOT NULL,
    cargo_type      TEXT,
    detour_km       NUMERIC(8,2) DEFAULT 0,
    quote_inr       NUMERIC(12,2),
    budget_usd      NUMERIC(10,2),
    status          TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','matched','booked','expired','cancelled')),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE freight_bookings (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipper_org_id       UUID NOT NULL REFERENCES organizations(id),
    transporter_org_id   UUID NOT NULL REFERENCES organizations(id),
    truck_listing_id     UUID REFERENCES freight_listings(id),
    load_listing_id      UUID REFERENCES freight_listings(id),
    agreed_freight_usd   NUMERIC(12,2) NOT NULL,
    disclosed_fee_pct    NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    accepted_by          TEXT NOT NULL, -- explicit, attributable acceptance (never auto-booked)
    gps_consent_granted  BOOLEAN NOT NULL DEFAULT FALSE,
    pickup_qr_token      TEXT UNIQUE,
    delivery_qr_token    TEXT UNIQUE,
    status               TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed','picked_up','in_transit','delivered','exception')),
    baseline_fuel_l      NUMERIC(10,2),
    matched_fuel_l       NUMERIC(10,2), -- Used for separate logistics avoided-emissions calc
    pod_s3_key           TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE freight_disputes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id  UUID NOT NULL REFERENCES freight_bookings(id),
    raised_by   UUID NOT NULL REFERENCES organizations(id),
    issue       TEXT NOT NULL,
    amount_inr  NUMERIC(12,2),
    status      TEXT NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. ZEVRA CONNECT RFQS & EU IMPORTER HUB SUPPLIER PORTAL
CREATE TABLE connect_rfqs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_org_id    UUID REFERENCES organizations(id),
    external_buyer  CITEXT, -- Supports RFQs converted directly from Mail Bell
    supplier_org_id UUID NOT NULL REFERENCES organizations(id),
    product_id      UUID REFERENCES products(id),
    requested_qty   NUMERIC(12,2) NOT NULL,
    target_incoterm TEXT,
    status          TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','quoted','negotiating','awarded','closed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE importer_supplier_links (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    importer_org_id    UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    supplier_org_id    UUID REFERENCES organizations(id),
    supplier_name      TEXT,
    invited_email      CITEXT NOT NULL,
    invite_token_hash  BYTEA NOT NULL UNIQUE,
    cn_code            TEXT,
    intensity_tco2e_t  NUMERIC(10,4),
    imported_tonnes    NUMERIC(12,3) DEFAULT 0,
    submission_status  TEXT NOT NULL DEFAULT 'invited' CHECK (submission_status IN ('invited','onboarding','pending_evidence','submitted','verified')),
    last_reminder_at   TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/* -------------------------------------------------------------------------- */
/* ROW-LEVEL SECURITY                                                          */
/* -------------------------------------------------------------------------- */

-- Application role: no BYPASSRLS. The API connects as zevra_app.
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'zevra_app') THEN
        CREATE ROLE zevra_app NOLOGIN;
    END IF;
END $$;

CREATE FUNCTION current_org() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT NULLIF(current_setting('app.current_org_id', true), '')::uuid $$;

DO $$
DECLARE t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['plants','factory_zones','products','production_batches','inventory_items','stock_movements',
        'financial_vouchers','tally_sync_log','mail_connections','mail_messages_and_notifications','activity_records',
        'qr_identities','container_batch_links','qr_events','freight_listings']
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (org_id = current_org()) WITH CHECK (org_id = current_org())', t);
        EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON %I TO zevra_app', t);
    END LOOP;
END $$;

-- Freight bookings are visible to both sides of the booking.
ALTER TABLE freight_bookings ENABLE ROW LEVEL SECURITY;
ALTER TABLE freight_bookings FORCE ROW LEVEL SECURITY;
CREATE POLICY booking_parties ON freight_bookings USING (current_org() IN (shipper_org_id, transporter_org_id));

-- Open marketplace listings are discoverable on the corridor; edits stay with the owner.
CREATE POLICY marketplace_read ON freight_listings FOR SELECT USING (status = 'open' AND current_setting('app.marketplace', true) = 'on');

-- RFQs: visible to buyer and supplier only.
ALTER TABLE connect_rfqs ENABLE ROW LEVEL SECURITY;
ALTER TABLE connect_rfqs FORCE ROW LEVEL SECURITY;
CREATE POLICY rfq_parties ON connect_rfqs USING (current_org() IN (buyer_org_id, supplier_org_id));

-- Importer Hub: the importer sees its supplier links; a supplier sees only its own row —
-- never another supplier's pricing, BOM or records.
ALTER TABLE importer_supplier_links ENABLE ROW LEVEL SECURITY;
ALTER TABLE importer_supplier_links FORCE ROW LEVEL SECURITY;
CREATE POLICY importer_owner ON importer_supplier_links USING (importer_org_id = current_org());
CREATE POLICY supplier_self ON importer_supplier_links FOR SELECT USING (supplier_org_id = current_org());

-- Public QR resolver: read-only access to ACTIVE/REVOKED passports when the resolver flag is set.
CREATE POLICY public_resolver ON qr_identities FOR SELECT USING (current_setting('app.public_resolver', true) = 'on');

GRANT SELECT, INSERT, UPDATE ON freight_bookings, connect_rfqs, importer_supplier_links, freight_disputes TO zevra_app;
GRANT SELECT ON organizations, cbam_rule_sets, transporter_vehicles TO zevra_app;
GRANT SELECT, INSERT, DELETE ON user_hint_dismissals TO zevra_app;

CREATE INDEX ON factory_zones (plant_id);
CREATE INDEX ON production_batches (org_id, stage);
CREATE INDEX ON financial_vouchers (org_id, voucher_date);
CREATE INDEX ON mail_messages_and_notifications (org_id, is_read, received_at DESC);
CREATE INDEX ON freight_listings (corridor_code, status, window_start);
CREATE INDEX ON importer_supplier_links (importer_org_id, submission_status);
