-- ZEVRA // 001 — standard auth + strictly separated hidden PGP admin vault (PostgreSQL 17)
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT NOT NULL UNIQUE,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    full_name       TEXT NOT NULL,
    password_hash   TEXT, -- NULL for federated/passkey-only users (Argon2id PHC string)
    auth_origin     TEXT NOT NULL CHECK (auth_origin IN ('manual','google','apple','proton')),
    proton_wkd_pgp  TEXT, -- Optional cached armored public key for Proton WKD users
    totp_secret_enc BYTEA,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','pending')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE oauth_identities (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT NOT NULL CHECK (provider IN ('google','apple','proton_oidc')),
    provider_subject TEXT NOT NULL,
    provider_email   CITEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_subject)
);

CREATE TABLE auth_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    email       CITEXT NOT NULL,
    token_hash  BYTEA NOT NULL UNIQUE, -- SHA-256 of opaque token
    purpose     TEXT NOT NULL CHECK (purpose IN ('email_verify','password_reset','proton_magic')),
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_passkeys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id   BYTEA NOT NULL UNIQUE,
    public_key      BYTEA NOT NULL,
    sign_count      BIGINT NOT NULL DEFAULT 0,
    label           TEXT NOT NULL DEFAULT 'Proton Pass / Device Key',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id      UUID, -- Populated after onboarding
    token_hash  BYTEA NOT NULL UNIQUE,
    ip_address  INET,
    user_agent  TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- STRICTLY SEPARATED HIDDEN PGP ADMIN TABLES
CREATE TABLE admin_pgp_keys (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    handle             TEXT NOT NULL UNIQUE,
    fingerprint        VARCHAR(64) NOT NULL UNIQUE, -- Uppercase full v4 (40) / v6 (64) fingerprint
    armored_public_key TEXT NOT NULL,
    role               TEXT NOT NULL DEFAULT 'super_admin',
    status             TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','revoked')),
    key_expires_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at         TIMESTAMPTZ
);

CREATE TABLE admin_auth_challenges (
    id                   UUID PRIMARY KEY, -- UUIDv7
    admin_key_id         UUID NOT NULL REFERENCES admin_pgp_keys(id) ON DELETE CASCADE,
    canonical_sha256     BYTEA NOT NULL,
    client_ip            INET NOT NULL,
    attempt_count        SMALLINT NOT NULL DEFAULT 0,
    expires_at           TIMESTAMPTZ NOT NULL,
    consumed_at          TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE admin_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_key_id UUID NOT NULL REFERENCES admin_pgp_keys(id) ON DELETE CASCADE,
    token_hash   BYTEA NOT NULL UNIQUE,
    ip_ua_hash   BYTEA NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE admin_audit_logs (
    id           BIGSERIAL PRIMARY KEY,
    admin_key_id UUID REFERENCES admin_pgp_keys(id),
    event_type   TEXT NOT NULL, -- 'challenge_issued','challenge_decoy','verify_success','verify_failed','replay_blocked'
    ip_address   INET,
    metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit rows are immutable.
CREATE FUNCTION forbid_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'admin_audit_logs is append-only'; END $$;
CREATE TRIGGER admin_audit_logs_immutable BEFORE UPDATE OR DELETE ON admin_audit_logs
    FOR EACH ROW EXECUTE FUNCTION forbid_audit_mutation();

CREATE INDEX ON auth_tokens (expires_at) WHERE consumed_at IS NULL;
CREATE INDEX ON user_sessions (user_id) WHERE revoked_at IS NULL;
CREATE INDEX ON admin_auth_challenges (expires_at) WHERE consumed_at IS NULL;
