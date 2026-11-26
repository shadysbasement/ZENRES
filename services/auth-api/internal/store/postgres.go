package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements Users and Vault against the schema in db/migrations/001_auth_and_pgp_vault.sql.
// Queries are hand-written with pgx; db/queries/*.sql mirrors them for sqlc generation.
type Postgres struct{ DB *pgxpool.Pool }

func NewPostgres(ctx context.Context, url string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Postgres{DB: pool}, nil
}

func isUnique(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (p *Postgres) CreateUser(ctx context.Context, u *User) error {
	err := p.DB.QueryRow(ctx, `INSERT INTO users (email, email_verified, full_name, password_hash, auth_origin, proton_wkd_pgp, status)
		VALUES ($1,$2,$3,$4,$5,$6,COALESCE(NULLIF($7,''),'active')) RETURNING id, created_at`,
		u.Email, u.EmailVerified, u.FullName, nullStr(u.PasswordHash), u.AuthOrigin, nullStr(u.ProtonWKDPGP), u.Status).Scan(&u.ID, &u.CreatedAt)
	if isUnique(err) {
		return ErrConflict
	}
	return err
}

func (p *Postgres) CreateUserWithLegal(ctx context.Context, u *User, termsVersion, privacyVersion string, marketing bool) error {
	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `INSERT INTO users (email, email_verified, full_name, password_hash, auth_origin, proton_wkd_pgp, status)
		VALUES ($1,$2,$3,$4,$5,$6,COALESCE(NULLIF($7,''),'active')) RETURNING id, created_at`,
		u.Email, u.EmailVerified, u.FullName, nullStr(u.PasswordHash), u.AuthOrigin, nullStr(u.ProtonWKDPGP), u.Status).Scan(&u.ID, &u.CreatedAt)
	if isUnique(err) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	for _, choice := range []struct {
		id, version string
		accepted    bool
	}{
		{"terms", termsVersion, true}, {"privacy_ack", privacyVersion, true}, {"marketing", "v1", marketing},
	} {
		if _, err = tx.Exec(ctx, `INSERT INTO user_policy_events (user_id, policy_id, policy_version, accepted, source)
			VALUES ($1,$2,$3,$4,'registration')`, u.ID, choice.id, choice.version, choice.accepted); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *Postgres) RecordPolicyEvent(ctx context.Context, e PolicyEvent) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO user_policy_events (user_id, org_id, policy_id, policy_version, accepted, source, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, e.UserID, e.OrgID, e.PolicyID, e.PolicyVersion, e.Accepted, e.Source, e.OccurredAt)
	return err
}

func (p *Postgres) PolicyEventsForUser(ctx context.Context, userID uuid.UUID) ([]PolicyEvent, error) {
	rows, err := p.DB.Query(ctx, `SELECT user_id, org_id, policy_id, policy_version, accepted, source, occurred_at
		FROM user_policy_events WHERE user_id=$1 ORDER BY occurred_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PolicyEvent
	for rows.Next() {
		var e PolicyEvent
		if err := rows.Scan(&e.UserID, &e.OrgID, &e.PolicyID, &e.PolicyVersion, &e.Accepted, &e.Source, &e.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const userCols = `id, email, email_verified, full_name, COALESCE(password_hash,''), auth_origin, COALESCE(proton_wkd_pgp,''), status, created_at`

func scanUser(row pgx.Row) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.EmailVerified, &u.FullName, &u.PasswordHash, &u.AuthOrigin, &u.ProtonWKDPGP, &u.Status, &u.CreatedAt); err != nil {
		return nil, notFound(err)
	}
	return &u, nil
}

func (p *Postgres) UserByEmail(ctx context.Context, email string) (*User, error) {
	return scanUser(p.DB.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE email = $1`, email))
}

func (p *Postgres) UserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return scanUser(p.DB.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

func (p *Postgres) UpdateUser(ctx context.Context, u *User) error {
	_, err := p.DB.Exec(ctx, `UPDATE users SET email_verified=$2, full_name=$3, password_hash=$4, proton_wkd_pgp=$5, status=$6, updated_at=NOW() WHERE id=$1`,
		u.ID, u.EmailVerified, u.FullName, nullStr(u.PasswordHash), nullStr(u.ProtonWKDPGP), u.Status)
	return err
}

func (p *Postgres) CreateToken(ctx context.Context, t *AuthToken) error {
	return p.DB.QueryRow(ctx, `INSERT INTO auth_tokens (user_id, email, token_hash, purpose, expires_at) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		t.UserID, t.Email, t.TokenHash, t.Purpose, t.ExpiresAt).Scan(&t.ID)
}

func (p *Postgres) ConsumeToken(ctx context.Context, hash []byte, purpose string, now time.Time) (*AuthToken, error) {
	var t AuthToken
	err := p.DB.QueryRow(ctx, `UPDATE auth_tokens SET consumed_at = $3
		WHERE token_hash = $1 AND purpose = $2 AND consumed_at IS NULL AND expires_at > $3
		RETURNING id, user_id, email, token_hash, purpose, expires_at, consumed_at`, hash, purpose, now).
		Scan(&t.ID, &t.UserID, &t.Email, &t.TokenHash, &t.Purpose, &t.ExpiresAt, &t.ConsumedAt)
	if err == nil {
		return &t, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// Distinguish consumed / expired / unknown for user-facing messaging.
	var consumed *time.Time
	var exp time.Time
	if err := p.DB.QueryRow(ctx, `SELECT consumed_at, expires_at FROM auth_tokens WHERE token_hash=$1 AND purpose=$2`, hash, purpose).Scan(&consumed, &exp); err != nil {
		return nil, ErrNotFound
	}
	if consumed != nil {
		return nil, ErrConsumed
	}
	return nil, ErrExpired
}

func (p *Postgres) LinkIdentity(ctx context.Context, id OAuthIdentity) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO oauth_identities (user_id, provider, provider_subject, provider_email) VALUES ($1,$2,$3,$4)`,
		id.UserID, id.Provider, id.ProviderSubject, nullStr(id.ProviderEmail))
	if isUnique(err) {
		return ErrConflict
	}
	return err
}

func (p *Postgres) IdentityBySubject(ctx context.Context, provider, subject string) (*OAuthIdentity, error) {
	var o OAuthIdentity
	err := p.DB.QueryRow(ctx, `SELECT user_id, provider, provider_subject, COALESCE(provider_email,'') FROM oauth_identities WHERE provider=$1 AND provider_subject=$2`, provider, subject).
		Scan(&o.UserID, &o.Provider, &o.ProviderSubject, &o.ProviderEmail)
	if err != nil {
		return nil, notFound(err)
	}
	return &o, nil
}

func (p *Postgres) AddPasskey(ctx context.Context, k *Passkey) error {
	return p.DB.QueryRow(ctx, `INSERT INTO user_passkeys (user_id, credential_id, public_key, sign_count, label) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		k.UserID, k.CredentialID, k.PublicKey, int64(k.SignCount), k.Label).Scan(&k.ID)
}

func (p *Postgres) PasskeysForUser(ctx context.Context, userID uuid.UUID) ([]Passkey, error) {
	rows, err := p.DB.Query(ctx, `SELECT id, user_id, credential_id, public_key, sign_count, label FROM user_passkeys WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Passkey
	for rows.Next() {
		var k Passkey
		var sc int64
		if err := rows.Scan(&k.ID, &k.UserID, &k.CredentialID, &k.PublicKey, &sc, &k.Label); err != nil {
			return nil, err
		}
		k.SignCount = uint32(sc)
		out = append(out, k)
	}
	return out, rows.Err()
}

func (p *Postgres) PasskeyByCredentialID(ctx context.Context, credID []byte) (*Passkey, error) {
	var k Passkey
	var sc int64
	err := p.DB.QueryRow(ctx, `SELECT id, user_id, credential_id, public_key, sign_count, label FROM user_passkeys WHERE credential_id=$1`, credID).
		Scan(&k.ID, &k.UserID, &k.CredentialID, &k.PublicKey, &sc, &k.Label)
	if err != nil {
		return nil, notFound(err)
	}
	k.SignCount = uint32(sc)
	return &k, nil
}

func (p *Postgres) UpdatePasskey(ctx context.Context, k *Passkey) error {
	_, err := p.DB.Exec(ctx, `UPDATE user_passkeys SET public_key=$2, sign_count=$3 WHERE id=$1`, k.ID, k.PublicKey, int64(k.SignCount))
	return err
}

func (p *Postgres) CreateSession(ctx context.Context, s *Session) error {
	return p.DB.QueryRow(ctx, `INSERT INTO user_sessions (user_id, org_id, token_hash, ip_address, user_agent, expires_at) VALUES ($1,$2,$3,NULLIF($4,'')::inet,$5,$6) RETURNING id`,
		s.UserID, s.OrgID, s.TokenHash, s.IP, s.UserAgent, s.ExpiresAt).Scan(&s.ID)
}

func (p *Postgres) SessionByHash(ctx context.Context, hash []byte, now time.Time) (*Session, error) {
	var s Session
	err := p.DB.QueryRow(ctx, `SELECT s.id, s.user_id, COALESCE(s.org_id, m.org_id), s.token_hash, s.expires_at
		FROM user_sessions s LEFT JOIN org_memberships m ON m.user_id = s.user_id
		WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at > $2 LIMIT 1`, hash, now).
		Scan(&s.ID, &s.UserID, &s.OrgID, &s.TokenHash, &s.ExpiresAt)
	if err != nil {
		return nil, notFound(err)
	}
	return &s, nil
}

func (p *Postgres) RevokeSession(ctx context.Context, hash []byte) error {
	_, err := p.DB.Exec(ctx, `UPDATE user_sessions SET revoked_at=NOW() WHERE token_hash=$1`, hash)
	return err
}

func (p *Postgres) CreateOrg(ctx context.Context, o *Org, userID uuid.UUID, role string) error {
	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := tx.QueryRow(ctx, `INSERT INTO organizations (name, org_type, country_code, gstin_or_eori, subscription_tier, inbound_mail_slug, primary_sector)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, o.Name, o.OrgType, o.CountryCode, nullStr(o.TaxID), o.Tier, o.InboundSlug, o.Sector).Scan(&o.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO org_memberships (org_id, user_id, role) VALUES ($1,$2,$3)`, o.ID, userID, role); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) OrgForUser(ctx context.Context, userID uuid.UUID) (*Org, error) {
	var o Org
	err := p.DB.QueryRow(ctx, `SELECT o.id, o.name, o.org_type, o.country_code, COALESCE(o.gstin_or_eori,''), o.subscription_tier, COALESCE(o.primary_sector,''), o.inbound_mail_slug
		FROM organizations o JOIN org_memberships m ON m.org_id=o.id WHERE m.user_id=$1 LIMIT 1`, userID).
		Scan(&o.ID, &o.Name, &o.OrgType, &o.CountryCode, &o.TaxID, &o.Tier, &o.Sector, &o.InboundSlug)
	if err != nil {
		return nil, notFound(err)
	}
	return &o, nil
}

/* ---------------- vault ---------------- */

const adminCols = `id, handle, fingerprint, armored_public_key, role, status, key_expires_at, revoked_at`

func scanAdmin(row pgx.Row) (*AdminKey, error) {
	var k AdminKey
	if err := row.Scan(&k.ID, &k.Handle, &k.Fingerprint, &k.ArmoredPublicKey, &k.Role, &k.Status, &k.KeyExpiresAt, &k.RevokedAt); err != nil {
		return nil, notFound(err)
	}
	k.Fingerprint = strings.TrimSpace(k.Fingerprint)
	return &k, nil
}

func (p *Postgres) AdminKeyByIdentifier(ctx context.Context, identifier string) (*AdminKey, error) {
	return scanAdmin(p.DB.QueryRow(ctx, `SELECT `+adminCols+` FROM admin_pgp_keys WHERE status='active' AND revoked_at IS NULL AND (fingerprint=upper($1) OR lower(handle)=lower($1))`, identifier))
}

func (p *Postgres) AdminKeyByID(ctx context.Context, id uuid.UUID) (*AdminKey, error) {
	return scanAdmin(p.DB.QueryRow(ctx, `SELECT `+adminCols+` FROM admin_pgp_keys WHERE id=$1`, id))
}

func (p *Postgres) InsertAdminKey(ctx context.Context, k *AdminKey) error {
	err := p.DB.QueryRow(ctx, `INSERT INTO admin_pgp_keys (handle, fingerprint, armored_public_key, role, key_expires_at) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		k.Handle, k.Fingerprint, k.ArmoredPublicKey, k.Role, k.KeyExpiresAt).Scan(&k.ID)
	if isUnique(err) {
		return ErrConflict
	}
	return err
}

func (p *Postgres) InsertChallenge(ctx context.Context, c *Challenge) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO admin_auth_challenges (id, admin_key_id, canonical_sha256, client_ip, expires_at) VALUES ($1,$2,$3,$4::inet,$5)`,
		c.ID, c.AdminKeyID, c.CanonicalSHA256, c.ClientIP, c.ExpiresAt)
	return err
}

func (p *Postgres) ChallengeByID(ctx context.Context, id uuid.UUID) (*Challenge, error) {
	var c Challenge
	err := p.DB.QueryRow(ctx, `SELECT id, admin_key_id, canonical_sha256, host(client_ip), attempt_count, expires_at, consumed_at FROM admin_auth_challenges WHERE id=$1`, id).
		Scan(&c.ID, &c.AdminKeyID, &c.CanonicalSHA256, &c.ClientIP, &c.AttemptCount, &c.ExpiresAt, &c.ConsumedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return &c, nil
}

func (p *Postgres) IncrementAttempt(ctx context.Context, id uuid.UUID) (int, error) {
	var n int
	err := p.DB.QueryRow(ctx, `UPDATE admin_auth_challenges SET attempt_count = attempt_count + 1 WHERE id=$1 RETURNING attempt_count`, id).Scan(&n)
	return n, notFound(err)
}

func (p *Postgres) ConsumeChallenge(ctx context.Context, id uuid.UUID, now time.Time) (bool, error) {
	tag, err := p.DB.Exec(ctx, `UPDATE admin_auth_challenges SET consumed_at=$2 WHERE id=$1 AND consumed_at IS NULL`, id, now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (p *Postgres) CreateAdminSession(ctx context.Context, s *AdminSession) error {
	return p.DB.QueryRow(ctx, `INSERT INTO admin_sessions (admin_key_id, token_hash, ip_ua_hash, expires_at) VALUES ($1,$2,$3,$4) RETURNING id`,
		s.AdminKeyID, s.TokenHash, s.IPUAHash, s.ExpiresAt).Scan(&s.ID)
}

func (p *Postgres) AdminSessionByHash(ctx context.Context, hash []byte, now time.Time) (*AdminSession, error) {
	var s AdminSession
	err := p.DB.QueryRow(ctx, `SELECT id, admin_key_id, token_hash, ip_ua_hash, expires_at, last_seen_at FROM admin_sessions WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > $2`, hash, now).
		Scan(&s.ID, &s.AdminKeyID, &s.TokenHash, &s.IPUAHash, &s.ExpiresAt, &s.LastSeenAt)
	if err != nil {
		return nil, notFound(err)
	}
	return &s, nil
}

func (p *Postgres) TouchAdminSession(ctx context.Context, id uuid.UUID, now time.Time) error {
	_, err := p.DB.Exec(ctx, `UPDATE admin_sessions SET last_seen_at=$2 WHERE id=$1`, id, now)
	return err
}

func (p *Postgres) RevokeAdminSession(ctx context.Context, hash []byte) error {
	_, err := p.DB.Exec(ctx, `UPDATE admin_sessions SET revoked_at=NOW() WHERE token_hash=$1 AND revoked_at IS NULL`, hash)
	return err
}

func (p *Postgres) CreateHandoff(ctx context.Context, codeHash []byte, token string, expiresAt time.Time) error {
	_, err := p.DB.Exec(ctx, `INSERT INTO admin_handoff_codes (code_hash, session_token, expires_at) VALUES ($1,$2,$3)`, codeHash, token, expiresAt)
	return err
}

func (p *Postgres) ConsumeHandoff(ctx context.Context, codeHash []byte, now time.Time) (string, error) {
	var token string
	err := p.DB.QueryRow(ctx, `DELETE FROM admin_handoff_codes WHERE code_hash=$1 AND expires_at>$2 RETURNING session_token`, codeHash, now).Scan(&token)
	return token, notFound(err)
}

func (p *Postgres) Audit(ctx context.Context, e AuditEvent) error {
	meta, _ := json.Marshal(e.Metadata)
	if e.Metadata == nil {
		meta = []byte("{}")
	}
	_, err := p.DB.Exec(ctx, `INSERT INTO admin_audit_logs (admin_key_id, event_type, ip_address, metadata) VALUES ($1,$2,NULLIF($3,'')::inet,$4)`, e.AdminKeyID, e.EventType, e.IP, meta)
	return err
}
