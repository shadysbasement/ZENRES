// Package pgp implements Zevra's zero-trust admin vault: an OpenPGP (RFC 9580) challenge-response
// protocol where the admin's private key never touches the browser or the server.
package pgp

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/google/uuid"

	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

var (
	ErrChallengeInvalid = errors.New("challenge_invalid")
	ErrChallengeExpired = errors.New("challenge_expired")
	ErrTooManyAttempts  = errors.New("too_many_attempts")
	ErrSignature        = errors.New("signature_invalid")
	ErrPayloadMismatch  = errors.New("payload_mismatch")
	ErrReplay           = errors.New("replay_blocked")
	ErrSubnet           = errors.New("subnet_mismatch")
	ErrKeyInactive      = errors.New("key_inactive")
)

const (
	ChallengeTTL      = 120 * time.Second
	MaxVerifyAttempts = 3
	SessionTTL        = 4 * time.Hour
	SessionIdle       = 30 * time.Minute
)

type Vault struct {
	Store  store.Vault
	Domain string
	Now    func() time.Time
	pgp    *crypto.PGPHandle
	decoy  *crypto.Key
}

func New(s store.Vault, domain string) (*Vault, error) {
	pgp := crypto.PGPWithProfile(profile.RFC9580())
	// Decoy key: unknown identifiers receive an indistinguishable challenge encrypted to this key,
	// so neither the response shape nor timing reveals whether an admin identifier exists.
	decoy, err := pgp.KeyGeneration().AddUserId("decoy", "decoy@invalid").New().GenerateKey()
	if err != nil {
		return nil, err
	}
	return &Vault{Store: s, Domain: domain, Now: time.Now, pgp: pgp, decoy: decoy}, nil
}

type ChallengeOut struct {
	ChallengeID      string `json:"challenge_id"`
	ArmoredChallenge string `json:"armored_challenge"`
	ExpiresIn        int    `json:"expires_in"`
}

// Canonical builds the exact plaintext the admin must sign.
func Canonical(id uuid.UUID, domain, fingerprint, nonce string, issued, expires time.Time) string {
	return strings.Join([]string{
		"ZEVRA-ADMIN-AUTH-V1",
		"challenge_id: " + id.String(),
		"domain: " + domain,
		"fingerprint: " + fingerprint,
		"nonce: " + nonce,
		"issued_at: " + issued.UTC().Format(time.RFC3339),
		"expires_at: " + expires.UTC().Format(time.RFC3339),
	}, "\n")
}

// normalise mirrors OpenPGP cleartext canonicalisation: CRLF→LF, trailing whitespace stripped per line.
func normalise(b []byte) []byte {
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return []byte(strings.TrimRight(strings.Join(lines, "\n"), "\n"))
}

func digest(b []byte) []byte {
	h := sha256.Sum256(normalise(b))
	return h[:]
}

// IssueChallenge implements protocol steps 1–2.
func (v *Vault) IssueChallenge(ctx context.Context, identifier, ip string) (*ChallengeOut, error) {
	identifier = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(identifier), " ", ""))
	now := v.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	nonce := platform.RandomHex(32)
	key, kerr := v.Store.AdminKeyByIdentifier(ctx, identifier)
	recipient := v.decoy
	fpr := identifier
	if kerr == nil && key.Status == "active" && key.RevokedAt == nil && (key.KeyExpiresAt == nil || key.KeyExpiresAt.After(now)) {
		pub, err := crypto.NewKeyFromArmored(key.ArmoredPublicKey)
		if err != nil {
			return nil, fmt.Errorf("stored key unreadable: %w", err)
		}
		recipient = pub
		fpr = key.Fingerprint
	}
	expires := now.Add(ChallengeTTL)
	plain := Canonical(id, v.Domain, fpr, nonce, now, expires)

	enc, err := v.pgp.Encryption().Recipient(recipient).New()
	if err != nil {
		return nil, err
	}
	msg, err := enc.Encrypt([]byte(plain))
	if err != nil {
		return nil, err
	}
	armored, err := msg.Armor()
	if err != nil {
		return nil, err
	}
	if recipient != v.decoy {
		if err := v.Store.InsertChallenge(ctx, &store.Challenge{ID: id, AdminKeyID: key.ID, CanonicalSHA256: digest([]byte(plain)), ClientIP: ip, ExpiresAt: expires}); err != nil {
			return nil, err
		}
		_ = v.Store.Audit(ctx, store.AuditEvent{AdminKeyID: &key.ID, EventType: "challenge_issued", IP: ip, Metadata: map[string]any{"challenge_id": id.String()}})
	} else {
		_ = v.Store.Audit(ctx, store.AuditEvent{EventType: "challenge_decoy", IP: ip, Metadata: map[string]any{"identifier_sha256": hex.EncodeToString(platform.SHA256(identifier))}})
	}
	return &ChallengeOut{ChallengeID: id.String(), ArmoredChallenge: armored, ExpiresIn: int(ChallengeTTL.Seconds())}, nil
}

type VerifyIn struct {
	ChallengeID string `json:"challenge_id"`
	Signed      string `json:"signed"`
	Signature   string `json:"signature,omitempty"` // optional armored detached signature over `signed`
}

type VerifyOut struct {
	Key   *store.AdminKey
	Token string // opaque admin session token (stored hashed)
}

// Verify implements protocol step 4 with strict ordering and an atomic consume.
func (v *Vault) Verify(ctx context.Context, in VerifyIn, ip, ua string) (*VerifyOut, error) {
	now := v.Now()
	id, err := uuid.Parse(in.ChallengeID)
	if err != nil {
		return nil, ErrChallengeInvalid
	}
	ch, err := v.Store.ChallengeByID(ctx, id)
	if err != nil {
		_ = v.Store.Audit(ctx, store.AuditEvent{EventType: "verify_failed", IP: ip, Metadata: map[string]any{"reason": "unknown_challenge"}})
		return nil, ErrChallengeInvalid
	}
	fail := func(reason error) (*VerifyOut, error) {
		_ = v.Store.Audit(ctx, store.AuditEvent{AdminKeyID: &ch.AdminKeyID, EventType: "verify_failed", IP: ip, Metadata: map[string]any{"reason": reason.Error(), "challenge_id": id.String()}})
		return nil, reason
	}
	if ch.ConsumedAt != nil {
		_ = v.Store.Audit(ctx, store.AuditEvent{AdminKeyID: &ch.AdminKeyID, EventType: "replay_blocked", IP: ip, Metadata: map[string]any{"challenge_id": id.String()}})
		return nil, ErrReplay
	}
	if now.After(ch.ExpiresAt) {
		return fail(ErrChallengeExpired)
	}
	if !platform.SameSubnet(ch.ClientIP, ip) {
		return fail(ErrSubnet)
	}
	if n, err := v.Store.IncrementAttempt(ctx, id); err != nil || n > MaxVerifyAttempts {
		return fail(ErrTooManyAttempts)
	}
	key, err := v.Store.AdminKeyByID(ctx, ch.AdminKeyID)
	if err != nil || key.Status != "active" || key.RevokedAt != nil || (key.KeyExpiresAt != nil && key.KeyExpiresAt.Before(now)) {
		return fail(ErrKeyInactive)
	}
	pub, err := crypto.NewKeyFromArmored(key.ArmoredPublicKey)
	if err != nil || !strings.EqualFold(pub.GetFingerprint(), key.Fingerprint) {
		return fail(ErrKeyInactive)
	}
	verifier, err := v.pgp.Verify().VerificationKey(pub).New()
	if err != nil {
		return nil, err
	}

	// Only bytes returned by the verifier are trusted — never wrapper headers or surrounding text.
	var verified []byte
	var res *crypto.VerifyResult
	if strings.TrimSpace(in.Signature) != "" {
		r, err := verifier.VerifyDetached([]byte(in.Signed), []byte(in.Signature), crypto.Armor)
		if err != nil {
			return fail(ErrSignature)
		}
		res, verified = r, []byte(in.Signed)
	} else {
		r, err := verifier.VerifyCleartext([]byte(in.Signed))
		if err != nil {
			return fail(ErrSignature)
		}
		res, verified = &r.VerifyResult, r.Cleartext()
	}
	if res.SignatureError() != nil {
		return fail(ErrSignature)
	}
	// Signer must be the pinned primary key (or one of its subkeys).
	if signer := res.SignedByKey(); signer == nil || !strings.EqualFold(signer.GetFingerprint(), key.Fingerprint) {
		return fail(ErrSignature)
	}
	if subtle.ConstantTimeCompare(digest(verified), ch.CanonicalSHA256) != 1 {
		return fail(ErrPayloadMismatch)
	}
	ok, err := v.Store.ConsumeChallenge(ctx, id, now)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = v.Store.Audit(ctx, store.AuditEvent{AdminKeyID: &key.ID, EventType: "replay_blocked", IP: ip, Metadata: map[string]any{"challenge_id": id.String()}})
		return nil, ErrReplay
	}
	token := platform.RandomToken(32)
	if err := v.Store.CreateAdminSession(ctx, &store.AdminSession{AdminKeyID: key.ID, TokenHash: platform.SHA256(token), IPUAHash: IPUAHash(ip, ua), ExpiresAt: now.Add(SessionTTL), LastSeenAt: now}); err != nil {
		return nil, err
	}
	_ = v.Store.Audit(ctx, store.AuditEvent{AdminKeyID: &key.ID, EventType: "verify_success", IP: ip, Metadata: map[string]any{"challenge_id": id.String()}})
	return &VerifyOut{Key: key, Token: token}, nil
}

// IPUAHash binds an admin session to the client's /24 subnet and user agent.
func IPUAHash(ip, ua string) []byte {
	return platform.SHA256(subnetKey(ip) + "|" + ua)
}

func subnetKey(ip string) string {
	if i := strings.LastIndex(ip, "."); i > 0 && strings.Count(ip, ".") == 3 {
		return ip[:i]
	}
	return ip
}

// ValidateSession enforces absolute (4h) and idle (30m) expiry plus the UA+IP binding.
func (v *Vault) ValidateSession(ctx context.Context, token, ip, ua string) (*store.AdminKey, error) {
	now := v.Now()
	s, err := v.Store.AdminSessionByHash(ctx, platform.SHA256(token), now)
	if err != nil {
		return nil, err
	}
	if now.Sub(s.LastSeenAt) > SessionIdle || subtle.ConstantTimeCompare(s.IPUAHash, IPUAHash(ip, ua)) != 1 {
		return nil, store.ErrNotFound
	}
	k, err := v.Store.AdminKeyByID(ctx, s.AdminKeyID)
	if err != nil || k.Status != "active" || k.RevokedAt != nil || (k.KeyExpiresAt != nil && !now.Before(*k.KeyExpiresAt)) {
		return nil, store.ErrNotFound
	}
	_ = v.Store.TouchAdminSession(ctx, s.ID, now)
	return k, nil
}

// FingerprintOf parses an armored public key and returns its uppercase fingerprint.
func FingerprintOf(armored string) (string, *crypto.Key, error) {
	k, err := crypto.NewKeyFromArmored(armored)
	if err != nil {
		return "", nil, err
	}
	if k.IsPrivate() {
		return "", nil, errors.New("refusing to enroll a PRIVATE key — export the public key only")
	}
	return strings.ToUpper(k.GetFingerprint()), k, nil
}
