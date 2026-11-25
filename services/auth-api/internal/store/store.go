// Package store defines persistence for the auth service. Two implementations exist:
// Postgres (production, see postgres.go) and an in-memory store used by tests and local demos.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not_found")
	ErrConflict = errors.New("conflict")
	ErrExpired  = errors.New("token_expired")
	ErrConsumed = errors.New("token_consumed")
)

type User struct {
	ID            uuid.UUID
	Email         string
	EmailVerified bool
	FullName      string
	PasswordHash  string // Argon2id PHC string; empty for federated/passkey-only users
	AuthOrigin    string // manual | google | apple | proton
	ProtonWKDPGP  string
	Status        string
	CreatedAt     time.Time
}

type OAuthIdentity struct {
	UserID          uuid.UUID
	Provider        string // google | apple | proton_oidc
	ProviderSubject string
	ProviderEmail   string
}

type AuthToken struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	Email      string
	TokenHash  []byte
	Purpose    string // email_verify | password_reset | proton_magic
	ExpiresAt  time.Time
	ConsumedAt *time.Time
}

type Passkey struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CredentialID []byte
	PublicKey    []byte // JSON-encoded webauthn.Credential
	SignCount    uint32
	Label        string
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OrgID     *uuid.UUID
	TokenHash []byte
	IP        string
	UserAgent string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type Org struct {
	ID          uuid.UUID
	Name        string
	OrgType     string
	CountryCode string
	TaxID       string
	Tier        string
	Sector      string
	InboundSlug string
}

type AdminKey struct {
	ID               uuid.UUID
	Handle           string
	Fingerprint      string // 40 (v4) or 64 (v6) uppercase hex
	ArmoredPublicKey string
	Role             string
	Status           string
	KeyExpiresAt     *time.Time
	RevokedAt        *time.Time
}

type Challenge struct {
	ID              uuid.UUID
	AdminKeyID      uuid.UUID
	CanonicalSHA256 []byte
	ClientIP        string
	AttemptCount    int
	ExpiresAt       time.Time
	ConsumedAt      *time.Time
}

type AdminSession struct {
	ID         uuid.UUID
	AdminKeyID uuid.UUID
	TokenHash  []byte
	IPUAHash   []byte
	ExpiresAt  time.Time
	LastSeenAt time.Time
	RevokedAt  *time.Time
}

type AuditEvent struct {
	AdminKeyID *uuid.UUID
	EventType  string
	IP         string
	Metadata   map[string]any
	CreatedAt  time.Time
}

// Users is the standard-user persistence surface.
type Users interface {
	CreateUser(ctx context.Context, u *User) error
	UserByEmail(ctx context.Context, email string) (*User, error)
	UserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, u *User) error

	CreateToken(ctx context.Context, t *AuthToken) error
	// ConsumeToken atomically marks a single-use token consumed and returns it.
	ConsumeToken(ctx context.Context, hash []byte, purpose string, now time.Time) (*AuthToken, error)

	LinkIdentity(ctx context.Context, id OAuthIdentity) error
	IdentityBySubject(ctx context.Context, provider, subject string) (*OAuthIdentity, error)

	AddPasskey(ctx context.Context, p *Passkey) error
	PasskeysForUser(ctx context.Context, userID uuid.UUID) ([]Passkey, error)
	PasskeyByCredentialID(ctx context.Context, credID []byte) (*Passkey, error)
	UpdatePasskey(ctx context.Context, p *Passkey) error

	CreateSession(ctx context.Context, s *Session) error
	SessionByHash(ctx context.Context, hash []byte, now time.Time) (*Session, error)
	RevokeSession(ctx context.Context, hash []byte) error

	CreateOrg(ctx context.Context, o *Org, userID uuid.UUID, role string) error
	OrgForUser(ctx context.Context, userID uuid.UUID) (*Org, error)
}

// Vault is the hidden PGP admin persistence surface (strictly separate tables).
type Vault interface {
	AdminKeyByIdentifier(ctx context.Context, identifier string) (*AdminKey, error)
	AdminKeyByID(ctx context.Context, id uuid.UUID) (*AdminKey, error)
	InsertAdminKey(ctx context.Context, k *AdminKey) error

	InsertChallenge(ctx context.Context, c *Challenge) error
	ChallengeByID(ctx context.Context, id uuid.UUID) (*Challenge, error)
	IncrementAttempt(ctx context.Context, id uuid.UUID) (int, error)
	// ConsumeChallenge sets consumed_at only if it is still NULL. false => replay.
	ConsumeChallenge(ctx context.Context, id uuid.UUID, now time.Time) (bool, error)

	CreateAdminSession(ctx context.Context, s *AdminSession) error
	AdminSessionByHash(ctx context.Context, hash []byte, now time.Time) (*AdminSession, error)
	TouchAdminSession(ctx context.Context, id uuid.UUID, now time.Time) error
	RevokeAdminSession(ctx context.Context, hash []byte) error
	// Handoff codes are single-use and shared across auth API instances.
	CreateHandoff(ctx context.Context, codeHash []byte, token string, expiresAt time.Time) error
	ConsumeHandoff(ctx context.Context, codeHash []byte, now time.Time) (string, error)

	Audit(ctx context.Context, e AuditEvent) error
}

// KV is the ephemeral TTL store (Redis in production) for PKCE state, WebAuthn sessions
// and one-time handoff codes.
type KV interface {
	Put(key string, value []byte, ttl time.Duration)
	// Take returns and deletes the value (single use).
	Take(key string) ([]byte, bool)
}
