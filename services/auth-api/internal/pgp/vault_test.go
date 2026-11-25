package pgp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"

	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
)

// Generates real ephemeral OpenPGP keypairs and runs the full protocol:
// encrypt → offline decrypt → clearsign → strict server verification.

type fixture struct {
	vault *Vault
	mem   *store.Memory
	priv  *crypto.Key
	fpr   string
	pgp   *crypto.PGPHandle
}

func setup(t *testing.T) *fixture {
	t.Helper()
	pgp := crypto.PGPWithProfile(profile.RFC9580())
	priv, err := pgp.KeyGeneration().AddUserId("Founder", "founder@zevra.io").New().GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubArmored, err := priv.GetArmoredPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	fpr, _, err := FingerprintOf(pubArmored)
	if err != nil {
		t.Fatal(err)
	}
	mem := store.NewMemory()
	if err := mem.InsertAdminKey(context.Background(), &store.AdminKey{Handle: "zv-founder-01", Fingerprint: fpr, ArmoredPublicKey: pubArmored, Role: "super_admin", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	v, err := New(mem, "zevra.io")
	if err != nil {
		t.Fatal(err)
	}
	return &fixture{vault: v, mem: mem, priv: priv, fpr: fpr, pgp: pgp}
}

// offlineSign is what the admin does locally: gpg --decrypt | gpg --clearsign.
func (f *fixture) offlineSign(t *testing.T, armored string, signer *crypto.Key, mutate func(string) string) string {
	t.Helper()
	dec, err := f.pgp.Decryption().DecryptionKey(f.priv).New()
	if err != nil {
		t.Fatal(err)
	}
	res, err := dec.Decrypt([]byte(armored), crypto.Armor)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	plain := res.String()
	if mutate != nil {
		plain = mutate(plain)
	}
	s, err := f.pgp.Sign().SigningKey(signer).New()
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.SignCleartext([]byte(plain))
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func countEvents(m *store.Memory, kind string) int {
	n := 0
	for _, e := range m.AuditLog {
		if e.EventType == kind {
			n++
		}
	}
	return n
}

func TestValidChallengeResponse(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, err := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ch.ArmoredChallenge, "BEGIN PGP MESSAGE") || ch.ExpiresIn != 120 {
		t.Fatalf("unexpected challenge %+v", ch)
	}
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	out, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.40", "UA")
	if err != nil {
		t.Fatalf("valid flow failed: %v", err)
	}
	if out.Token == "" || out.Key.Handle != "zv-founder-01" {
		t.Fatal("expected admin session")
	}
	if k, err := f.vault.ValidateSession(ctx, out.Token, "10.4.2.99", "UA"); err != nil || k.Fingerprint != f.fpr {
		t.Fatalf("session should validate within the same /24 + UA: %v", err)
	}
	if _, err := f.vault.ValidateSession(ctx, out.Token, "10.4.2.99", "Other UA"); err == nil {
		t.Fatal("session must be bound to user agent")
	}
	if countEvents(f.mem, "verify_success") != 1 {
		t.Fatal("expected verify_success audit row")
	}
}

func TestSessionStopsWhenAdminKeyIsRevoked(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, err := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)}, "10.4.2.17", "UA")
	if err != nil {
		t.Fatal(err)
	}
	f.mem.RevokeAdminKey(out.Key.ID)
	if _, err := f.vault.ValidateSession(ctx, out.Token, "10.4.2.17", "UA"); err == nil {
		t.Fatal("revoked admin key must invalidate its active session")
	}
}

func TestReplayIsBlocked(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); err != nil {
		t.Fatal(err)
	}
	_, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA")
	if !errors.Is(err, ErrReplay) || countEvents(f.mem, "replay_blocked") != 1 {
		t.Fatalf("expected replay_blocked, got %v", err)
	}
}

func TestTamperedPayloadFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, func(s string) string {
		return strings.Replace(s, "domain: zevra.io", "domain: evil.example", 1)
	})
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); !errors.Is(err, ErrPayloadMismatch) {
		t.Fatalf("expected payload mismatch, got %v", err)
	}
	if countEvents(f.mem, "verify_failed") != 1 {
		t.Fatal("expected verify_failed audit row")
	}
}

func TestEditedSignedBlockFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	forged := strings.Replace(signed, "ZEVRA-ADMIN-AUTH-V1", "ZEVRA-ADMIN-AUTH-V2", 1)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: forged}, "10.4.2.17", "UA"); !errors.Is(err, ErrSignature) {
		t.Fatalf("expected signature failure, got %v", err)
	}
}

func TestWrongKeyFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	attacker, _ := f.pgp.KeyGeneration().AddUserId("Mallory", "m@evil.example").New().GenerateKey()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, attacker, nil)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); !errors.Is(err, ErrSignature) {
		t.Fatalf("expected signature failure for wrong key, got %v", err)
	}
}

func TestExpiredChallengeFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	f.vault.Now = func() time.Time { return time.Now().Add(121 * time.Second) }
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); !errors.Is(err, ErrChallengeExpired) {
		t.Fatalf("expected expiry, got %v", err)
	}
}

func TestSubnetMismatchFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "203.0.113.9", "UA"); !errors.Is(err, ErrSubnet) {
		t.Fatalf("expected subnet mismatch, got %v", err)
	}
}

func TestThreeAttemptsThenInvalidated(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	for i := 0; i < 3; i++ {
		_, _ = f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: "-----BEGIN PGP SIGNED MESSAGE-----\ngarbage"}, "10.4.2.17", "UA")
	}
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected attempt cap, got %v", err)
	}
}

func TestUnknownIdentifierGetsDecoy(t *testing.T) {
	f := setup(t)
	ch, err := f.vault.IssueChallenge(context.Background(), strings.Repeat("A", 40), "10.4.2.17")
	if err != nil || !strings.Contains(ch.ArmoredChallenge, "BEGIN PGP MESSAGE") {
		t.Fatalf("decoy challenge should look identical: %v", err)
	}
	if _, err := f.vault.Verify(context.Background(), VerifyIn{ChallengeID: ch.ChallengeID, Signed: "x"}, "10.4.2.17", "UA"); !errors.Is(err, ErrChallengeInvalid) {
		t.Fatalf("decoy challenge must never verify: %v", err)
	}
}

func TestRevokedKeyFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ch, _ := f.vault.IssueChallenge(ctx, f.fpr, "10.4.2.17")
	k, _ := f.mem.AdminKeyByIdentifier(ctx, f.fpr)
	f.mem.RevokeAdminKey(k.ID)
	signed := f.offlineSign(t, ch.ArmoredChallenge, f.priv, nil)
	if _, err := f.vault.Verify(ctx, VerifyIn{ChallengeID: ch.ChallengeID, Signed: signed}, "10.4.2.17", "UA"); !errors.Is(err, ErrKeyInactive) {
		t.Fatalf("expected revoked key failure, got %v", err)
	}
}

func TestRefusesPrivateKeyEnrollment(t *testing.T) {
	f := setup(t)
	armored, _ := f.priv.Armor()
	if _, _, err := FingerprintOf(armored); err == nil {
		t.Fatal("private key must be refused")
	}
}
