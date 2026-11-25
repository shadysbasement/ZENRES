package store

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Memory implements Users, Vault and KV in-process. Used by tests and `STORE=memory` demos.
type Memory struct {
	mu         sync.Mutex
	users      map[uuid.UUID]*User
	tokens     []*AuthToken
	identities []OAuthIdentity
	passkeys   []*Passkey
	sessions   []*Session
	orgs       map[uuid.UUID]*Org
	members    map[uuid.UUID]uuid.UUID // user -> org
	adminKeys  []*AdminKey
	challenges map[uuid.UUID]*Challenge
	adminSess  []*AdminSession
	AuditLog   []AuditEvent
	kv         map[string]kvEntry
}

type kvEntry struct {
	v   []byte
	exp time.Time
}

func NewMemory() *Memory {
	return &Memory{
		users:      map[uuid.UUID]*User{},
		orgs:       map[uuid.UUID]*Org{},
		members:    map[uuid.UUID]uuid.UUID{},
		challenges: map[uuid.UUID]*Challenge{},
		kv:         map[string]kvEntry{},
	}
}

func (m *Memory) CreateUser(_ context.Context, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.users {
		if strings.EqualFold(x.Email, u.Email) {
			return ErrConflict
		}
	}
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *Memory) UserByEmail(_ context.Context, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) UserByID(_ context.Context, id uuid.UUID) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, ErrNotFound
}

func (m *Memory) UpdateUser(_ context.Context, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[u.ID]; !ok {
		return ErrNotFound
	}
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *Memory) CreateToken(_ context.Context, t *AuthToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t.ID = uuid.New()
	cp := *t
	m.tokens = append(m.tokens, &cp)
	return nil
}

func (m *Memory) ConsumeToken(_ context.Context, hash []byte, purpose string, now time.Time) (*AuthToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if bytes.Equal(t.TokenHash, hash) && t.Purpose == purpose {
			if t.ConsumedAt != nil {
				return nil, ErrConsumed
			}
			if now.After(t.ExpiresAt) {
				return nil, ErrExpired
			}
			n := now
			t.ConsumedAt = &n
			cp := *t
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) LinkIdentity(_ context.Context, id OAuthIdentity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.identities {
		if x.Provider == id.Provider && x.ProviderSubject == id.ProviderSubject {
			return ErrConflict
		}
	}
	m.identities = append(m.identities, id)
	return nil
}

func (m *Memory) IdentityBySubject(_ context.Context, provider, subject string) (*OAuthIdentity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.identities {
		if x.Provider == provider && x.ProviderSubject == subject {
			cp := x
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) AddPasskey(_ context.Context, p *Passkey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = uuid.New()
	cp := *p
	m.passkeys = append(m.passkeys, &cp)
	return nil
}

func (m *Memory) PasskeysForUser(_ context.Context, userID uuid.UUID) ([]Passkey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Passkey
	for _, p := range m.passkeys {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *Memory) PasskeyByCredentialID(_ context.Context, credID []byte) (*Passkey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.passkeys {
		if bytes.Equal(p.CredentialID, credID) {
			cp := *p
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) UpdatePasskey(_ context.Context, p *Passkey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.passkeys {
		if x.ID == p.ID {
			cp := *p
			m.passkeys[i] = &cp
			return nil
		}
	}
	return ErrNotFound
}

func (m *Memory) CreateSession(_ context.Context, s *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.ID = uuid.New()
	cp := *s
	m.sessions = append(m.sessions, &cp)
	return nil
}

func (m *Memory) SessionByHash(_ context.Context, hash []byte, now time.Time) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if bytes.Equal(s.TokenHash, hash) && s.RevokedAt == nil && now.Before(s.ExpiresAt) {
			cp := *s
			if org, ok := m.members[s.UserID]; ok {
				cp.OrgID = &org
			}
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) RevokeSession(_ context.Context, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, s := range m.sessions {
		if bytes.Equal(s.TokenHash, hash) {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *Memory) CreateOrg(_ context.Context, o *Org, userID uuid.UUID, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o.ID = uuid.New()
	cp := *o
	m.orgs[o.ID] = &cp
	m.members[userID] = o.ID
	return nil
}

func (m *Memory) OrgForUser(_ context.Context, userID uuid.UUID) (*Org, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, ok := m.members[userID]; ok {
		cp := *m.orgs[id]
		return &cp, nil
	}
	return nil, ErrNotFound
}

/* ---------------- vault ---------------- */

func (m *Memory) AdminKeyByIdentifier(_ context.Context, identifier string) (*AdminKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range m.adminKeys {
		if k.Fingerprint == strings.ToUpper(identifier) || strings.EqualFold(k.Handle, identifier) {
			cp := *k
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) AdminKeyByID(_ context.Context, id uuid.UUID) (*AdminKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range m.adminKeys {
		if k.ID == id {
			cp := *k
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) InsertAdminKey(_ context.Context, k *AdminKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.adminKeys {
		if x.Fingerprint == k.Fingerprint || strings.EqualFold(x.Handle, k.Handle) {
			return ErrConflict
		}
	}
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	cp := *k
	m.adminKeys = append(m.adminKeys, &cp)
	return nil
}

// RevokeAdminKey is a test helper.
func (m *Memory) RevokeAdminKey(id uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, k := range m.adminKeys {
		if k.ID == id {
			k.Status = "revoked"
			k.RevokedAt = &now
		}
	}
}

func (m *Memory) InsertChallenge(_ context.Context, c *Challenge) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.challenges[c.ID] = &cp
	return nil
}

func (m *Memory) ChallengeByID(_ context.Context, id uuid.UUID) (*Challenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.challenges[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, ErrNotFound
}

func (m *Memory) IncrementAttempt(_ context.Context, id uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.challenges[id]
	if !ok {
		return 0, ErrNotFound
	}
	c.AttemptCount++
	return c.AttemptCount, nil
}

func (m *Memory) ConsumeChallenge(_ context.Context, id uuid.UUID, now time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.challenges[id]
	if !ok || c.ConsumedAt != nil {
		return false, nil
	}
	c.ConsumedAt = &now
	return true, nil
}

func (m *Memory) CreateAdminSession(_ context.Context, s *AdminSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.ID = uuid.New()
	cp := *s
	m.adminSess = append(m.adminSess, &cp)
	return nil
}

func (m *Memory) AdminSessionByHash(_ context.Context, hash []byte, now time.Time) (*AdminSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.adminSess {
		if bytes.Equal(s.TokenHash, hash) && s.RevokedAt == nil && now.Before(s.ExpiresAt) {
			cp := *s
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *Memory) TouchAdminSession(_ context.Context, id uuid.UUID, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.adminSess {
		if s.ID == id {
			s.LastSeenAt = now
		}
	}
	return nil
}

func (m *Memory) CreateHandoff(_ context.Context, codeHash []byte, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv["handoff:"+string(codeHash)] = kvEntry{v: []byte(token), exp: expiresAt}
	return nil
}

func (m *Memory) ConsumeHandoff(_ context.Context, codeHash []byte, now time.Time) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := "handoff:" + string(codeHash)
	e, ok := m.kv[key]
	delete(m.kv, key)
	if !ok || !now.Before(e.exp) {
		return "", ErrNotFound
	}
	return string(e.v), nil
}

func (m *Memory) Audit(_ context.Context, e AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.CreatedAt = time.Now()
	m.AuditLog = append(m.AuditLog, e)
	return nil
}

/* ---------------- kv ---------------- */

func (m *Memory) Put(key string, value []byte, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] = kvEntry{v: value, exp: time.Now().Add(ttl)}
}

func (m *Memory) Take(key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.kv[key]
	delete(m.kv, key)
	if !ok || time.Now().After(e.exp) {
		return nil, false
	}
	return e.v, true
}
