// Package httpapi wires the Zevra auth service: manual Argon2id accounts, Google/Apple OIDC with PKCE,
// the privacy-first Proton flow (magic link + WKD + passkeys), onboarding, and the hidden PGP vault.
package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadysbasement/zevra/services/auth-api/internal/auth"
	"github.com/shadysbasement/zevra/services/auth-api/internal/pgp"
	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

type Config struct {
	PublicURL          string // https://zevra.io
	AppURL             string // https://app.zevra.io
	CookieDomain       string // .zevra.io (empty for host-only in dev)
	CookieSecure       bool
	GateSecret         string // X-Zevra-Gate header required on vault endpoints
	InternalSecret     string // X-Internal-Secret for service-to-service calls
	SessionTTL         time.Duration
	LegalApproved      bool
	TermsVersion       string
	PrivacyVersion     string
	SupportNotify      bool
	SupportPlanEnforce bool
}

type Server struct {
	Cfg       Config
	Users     store.Users
	KV        store.KV
	Vault     *pgp.Vault
	Limiter   platform.Limiter
	Mail      auth.Mailer
	Providers map[string]auth.IdentityProvider
	WebAuthn  *webauthn.WebAuthn
	SupportDB *pgxpool.Pool
	Now       func() time.Time
}

const (
	SessionCookie = "zv_session"
	CSRFCookie    = "zv_csrf"
)

var emailRe = regexp.MustCompile(`^[^@\s]{1,64}@[^@\s]+\.[^@\s]{2,}$`)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Recoverer, platform.SecurityHeaders, middleware.Timeout(15*time.Second))
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { platform.JSON(w, 200, map[string]string{"status": "ok"}) })

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(s.csrf)
		r.Post("/register", s.register)
		r.Post("/verify-email", s.verifyEmail)
		r.Post("/login", s.login)
		r.Post("/logout", s.logout)
		r.Get("/me", s.me)
		r.Get("/consent", s.consent)
		r.Post("/consent/marketing", s.updateMarketing)
		r.Post("/password/forgot", s.forgotPassword)
		r.Post("/password/reset", s.resetPassword)
		r.Post("/proton/wkd", s.protonWKD)
		r.Post("/proton/magic-link", s.protonMagicLink)
		r.Post("/proton/verify", s.protonVerify)
		r.Post("/passkey/register/start", s.passkeyRegisterStart)
		r.Post("/passkey/register/finish", s.passkeyRegisterFinish)
		r.Post("/passkey/login/start", s.passkeyLoginStart)
		r.Post("/passkey/login/finish", s.passkeyLoginFinish)
		r.Get("/oauth/{provider}/start", s.oauthStart)
		r.Get("/oauth/{provider}/callback", s.oauthCallback)
		r.Post("/oauth/{provider}/callback", s.oauthCallback) // Apple response_mode=form_post
		r.Post("/link/confirm", s.linkConfirm)
	})
	r.With(s.csrf).Post("/api/v1/onboarding/complete", s.onboardingComplete)
	r.Route("/api/v1/support", func(r chi.Router) {
		r.Get("/announcements", s.supportAnnouncements)
		r.With(s.csrf).Get("/tickets", s.supportTickets)
		r.With(s.csrf).Post("/tickets", s.supportCreateTicket)
		r.With(s.csrf).Get("/tickets/{id}", s.supportTicket)
		r.With(s.csrf).Post("/tickets/{id}/messages", s.supportCustomerMessage)
		r.With(s.csrf).Post("/tickets/{id}/feedback", s.supportFeedback)
	})
	r.Route("/internal/support", func(r chi.Router) {
		r.Use(s.internalOnly, s.supportAdminAuth)
		r.Get("/tickets", s.supportAdminTickets)
		r.Get("/agents", s.supportAdminAgents)
		r.Get("/costs", s.supportAdminCosts)
		r.Get("/tickets/{id}", s.supportAdminTicket)
		r.Post("/tickets/{id}", s.supportAdminUpdate)
		r.Post("/announcements", s.supportPublishAnnouncement)
	})

	// Hidden vault: 404 unless the SvelteKit gate route adds the secret header.
	r.Route("/api/v1/vault-auth", func(r chi.Router) {
		r.Use(s.stealth)
		r.Post("/challenge", s.vaultChallenge)
		r.Post("/verify", s.vaultVerify)
	})
	r.Route("/internal/vault-auth", func(r chi.Router) {
		r.Use(s.internalOnly)
		r.Post("/exchange", s.vaultExchange)
		r.Post("/session", s.vaultSession)
		r.Post("/logout", s.vaultLogout)
	})
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "Not Found", http.StatusNotFound) })
	return r
}

/* ------------------------------------------------------------------ */
/* middleware                                                          */
/* ------------------------------------------------------------------ */

// csrf enforces double-submit tokens on state-changing requests that carry a session cookie.
func (s *Server) csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || strings.HasSuffix(r.URL.Path, "/callback") {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := r.Cookie(SessionCookie); err == nil {
			c, err := r.Cookie(CSRFCookie)
			h := r.Header.Get("X-CSRF-Token")
			if err != nil || h == "" || subtle.ConstantTimeCompare([]byte(c.Value), []byte(h)) != 1 {
				platform.Error(w, http.StatusForbidden, "csrf_failed")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func notFound(w http.ResponseWriter) { http.Error(w, "Not Found", http.StatusNotFound) }

func (s *Server) stealth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		w.Header().Set("Cache-Control", "no-store")
		h := r.Header.Get("X-Zevra-Gate")
		if s.Cfg.GateSecret == "" || subtle.ConstantTimeCompare([]byte(h), []byte(s.Cfg.GateSecret)) != 1 {
			notFound(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) internalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("X-Internal-Secret")
		if s.Cfg.InternalSecret == "" || subtle.ConstantTimeCompare([]byte(h), []byte(s.Cfg.InternalSecret)) != 1 {
			notFound(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

/* ------------------------------------------------------------------ */
/* sessions                                                            */
/* ------------------------------------------------------------------ */

func (s *Server) issueSession(ctx context.Context, w http.ResponseWriter, r *http.Request, u *store.User) error {
	token := platform.RandomToken(32)
	ttl := s.Cfg.SessionTTL
	if ttl == 0 {
		ttl = 14 * 24 * time.Hour
	}
	if err := s.Users.CreateSession(ctx, &store.Session{UserID: u.ID, TokenHash: platform.SHA256(token), IP: platform.ClientIP(r), UserAgent: r.UserAgent(), ExpiresAt: s.Now().Add(ttl)}); err != nil {
		return err
	}
	// Rotate: revoke any session the browser previously held (privilege elevation / fixation).
	if old, err := r.Cookie(SessionCookie); err == nil {
		_ = s.Users.RevokeSession(ctx, platform.SHA256(old.Value))
	}
	http.SetCookie(w, s.cookie(SessionCookie, token, int(ttl.Seconds()), true))
	http.SetCookie(w, s.cookie(CSRFCookie, platform.RandomToken(24), int(ttl.Seconds()), false))
	return nil
}

func (s *Server) cookie(name, value string, maxAge int, httpOnly bool) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: "/", Domain: s.Cfg.CookieDomain, MaxAge: maxAge, HttpOnly: httpOnly, Secure: s.Cfg.CookieSecure, SameSite: http.SameSiteLaxMode}
}

func (s *Server) currentUser(r *http.Request) (*store.User, *store.Session, error) {
	c, err := r.Cookie(SessionCookie)
	if err != nil {
		return nil, nil, store.ErrNotFound
	}
	sess, err := s.Users.SessionByHash(r.Context(), platform.SHA256(c.Value), s.Now())
	if err != nil {
		return nil, nil, err
	}
	u, err := s.Users.UserByID(r.Context(), sess.UserID)
	return u, sess, err
}

/* ------------------------------------------------------------------ */
/* manual email + password                                             */
/* ------------------------------------------------------------------ */

type registerIn struct {
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	Plan           string `json:"plan"`
	TermsAccepted  bool   `json:"terms_accepted"`
	TermsVersion   string `json:"terms_version"`
	PrivacyVersion string `json:"privacy_version"`
	MarketingOptIn bool   `json:"marketing_opt_in"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var in registerIn
	if platform.DecodeLoose(r, &in) != nil || !emailRe.MatchString(in.Email) || strings.TrimSpace(in.FullName) == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	if s.Cfg.LegalApproved && (!in.TermsAccepted || in.TermsVersion != s.Cfg.TermsVersion || in.PrivacyVersion != s.Cfg.PrivacyVersion || s.Cfg.TermsVersion == "" || s.Cfg.PrivacyVersion == "") {
		platform.Error(w, 400, "legal_terms_required")
		return
	}
	if !s.Limiter.Allow("register:"+platform.ClientIP(r), 10, time.Hour) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		platform.Error(w, 400, "weak_password")
		return
	}
	u := &store.User{Email: strings.ToLower(strings.TrimSpace(in.Email)), FullName: strings.TrimSpace(in.FullName), PasswordHash: hash, AuthOrigin: "manual", Status: "pending"}
	var createErr error
	if s.Cfg.LegalApproved {
		createErr = s.Users.CreateUserWithLegal(r.Context(), u, s.Cfg.TermsVersion, s.Cfg.PrivacyVersion, in.MarketingOptIn)
	} else {
		createErr = s.Users.CreateUser(r.Context(), u)
	}
	if err := createErr; err != nil {
		if errors.Is(err, store.ErrConflict) {
			platform.Error(w, 409, "email_taken")
			return
		}
		s.fail(w, err)
		return
	}
	if err := s.sendToken(r.Context(), u, "email_verify", 24*time.Hour, "Verify your ZENRES email", nil); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"status": "verification_sent"})
}

func (s *Server) sendToken(ctx context.Context, u *store.User, purpose string, ttl time.Duration, subject string, key any) error {
	tok := platform.RandomToken(32)
	uid := u.ID
	if err := s.Users.CreateToken(ctx, &store.AuthToken{UserID: &uid, Email: u.Email, TokenHash: platform.SHA256(tok), Purpose: purpose, ExpiresAt: s.Now().Add(ttl)}); err != nil {
		return err
	}
	path := "/auth/verify-email?purpose=" + purpose + "&token=" + url.QueryEscape(tok)
	body := fmt.Sprintf("Hello %s,\n\nUse this single-use link within %s:\n%s%s\n\nIf you did not request this, ignore this email.\n— ZENRES", u.FullName, ttl, s.Cfg.PublicURL, path)
	return s.Mail.Send(ctx, u.Email, subject, body, nil)
}

func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Token == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	t, err := s.Users.ConsumeToken(r.Context(), platform.SHA256(in.Token), "email_verify", s.Now())
	if err != nil {
		s.tokenError(w, err)
		return
	}
	u, err := s.Users.UserByID(r.Context(), *t.UserID)
	if err != nil {
		s.fail(w, err)
		return
	}
	u.EmailVerified, u.Status = true, "active"
	_ = s.Users.UpdateUser(r.Context(), u)
	if err := s.issueSession(r.Context(), w, r, u); err != nil {
		s.fail(w, err)
		return
	}
	_, orgErr := s.Users.OrgForUser(r.Context(), u.ID)
	platform.JSON(w, 200, map[string]any{"offer_passkey": false, "onboarded": orgErr == nil})
}

func (s *Server) tokenError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrExpired):
		platform.Error(w, 410, "token_expired")
	case errors.Is(err, store.ErrConsumed):
		platform.Error(w, 410, "token_consumed")
	default:
		platform.Error(w, 400, "token_invalid")
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	ip := platform.ClientIP(r)
	if !s.Limiter.Allow("login-ip:"+ip, 20, 15*time.Minute) || !s.Limiter.Allow("login:"+email, 5, 15*time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	u, err := s.Users.UserByEmail(r.Context(), email)
	phc := ""
	if err == nil {
		phc = u.PasswordHash
	}
	if !auth.VerifyOrDummy(in.Password, phc) || err != nil {
		platform.Error(w, 401, "invalid_credentials")
		return
	}
	if !u.EmailVerified {
		_ = s.sendToken(r.Context(), u, "email_verify", 24*time.Hour, "Verify your ZENRES email", nil)
		platform.Error(w, 403, "email_unverified")
		return
	}
	if err := s.issueSession(r.Context(), w, r, u); err != nil {
		s.fail(w, err)
		return
	}
	_, orgErr := s.Users.OrgForUser(r.Context(), u.ID)
	platform.JSON(w, 200, map[string]any{"onboarded": orgErr == nil})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(SessionCookie); err == nil {
		_ = s.Users.RevokeSession(r.Context(), platform.SHA256(c.Value))
	}
	http.SetCookie(w, s.cookie(SessionCookie, "", -1, true))
	http.SetCookie(w, s.cookie(CSRFCookie, "", -1, false))
	w.WriteHeader(204)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, _, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	out := map[string]any{"id": u.ID, "email": u.Email, "fullName": u.FullName, "authOrigin": u.AuthOrigin, "onboarded": false}
	if org, err := s.Users.OrgForUser(r.Context(), u.ID); err == nil {
		out["onboarded"] = true
		out["org"] = map[string]any{"id": org.ID, "name": org.Name, "tier": org.Tier, "type": org.OrgType}
	}
	platform.JSON(w, 200, out)
}

func (s *Server) consent(w http.ResponseWriter, r *http.Request) {
	u, _, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	events, err := s.Users.PolicyEventsForUser(r.Context(), u.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	marketing := false
	for _, e := range events {
		if e.PolicyID == "marketing" {
			marketing = e.Accepted
		}
	}
	platform.JSON(w, 200, map[string]any{"marketing_opt_in": marketing, "events": events})
}

func (s *Server) updateMarketing(w http.ResponseWriter, r *http.Request) {
	u, sess, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	var in struct {
		MarketingOptIn *bool `json:"marketing_opt_in"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.MarketingOptIn == nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	if err := s.Users.RecordPolicyEvent(r.Context(), store.PolicyEvent{UserID: u.ID, OrgID: sess.OrgID, PolicyID: "marketing", PolicyVersion: "v1", Accepted: *in.MarketingOptIn, Source: "account_preferences", OccurredAt: s.Now().UTC()}); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"marketing_opt_in": *in.MarketingOptIn})
}

func (s *Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	_ = platform.DecodeLoose(r, &in)
	if s.Limiter.Allow("forgot:"+platform.ClientIP(r), 5, time.Hour) {
		if u, err := s.Users.UserByEmail(r.Context(), strings.ToLower(in.Email)); err == nil && u.PasswordHash != "" {
			_ = s.sendToken(r.Context(), u, "password_reset", time.Hour, "Reset your ZENRES password", nil)
		}
	}
	platform.JSON(w, 200, map[string]string{"status": "if_exists_sent"}) // identical response either way
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		platform.Error(w, 400, "weak_password")
		return
	}
	t, err := s.Users.ConsumeToken(r.Context(), platform.SHA256(in.Token), "password_reset", s.Now())
	if err != nil {
		s.tokenError(w, err)
		return
	}
	u, err := s.Users.UserByID(r.Context(), *t.UserID)
	if err != nil {
		s.fail(w, err)
		return
	}
	u.PasswordHash = hash
	_ = s.Users.UpdateUser(r.Context(), u)
	platform.JSON(w, 200, map[string]string{"status": "password_updated"})
}

/* ------------------------------------------------------------------ */
/* Proton: magic link + WKD + passkeys                                 */
/* ------------------------------------------------------------------ */

func (s *Server) protonWKD(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if platform.DecodeLoose(r, &in) != nil || !emailRe.MatchString(in.Email) {
		platform.Error(w, 400, "invalid_input")
		return
	}
	if !s.Limiter.Allow("wkd:"+platform.ClientIP(r), 30, 15*time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	out := map[string]any{"proton": auth.IsProtonEmail(in.Email), "wkd_key": false}
	if k, _ := auth.LookupWKD(ctx, in.Email); k != nil && k.CanEncrypt(s.Now().Unix()) {
		out["wkd_key"], out["fingerprint"] = true, strings.ToUpper(k.GetFingerprint())
	}
	platform.JSON(w, 200, out)
}

func (s *Server) protonMagicLink(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Encrypt  bool   `json:"encrypt"`
		Override bool   `json:"override"`
		Role     string `json:"role"`
	}
	if platform.DecodeLoose(r, &in) != nil || !emailRe.MatchString(in.Email) {
		platform.Error(w, 400, "invalid_input")
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !s.Limiter.Allow("magic:"+email, 5, 15*time.Minute) || !s.Limiter.Allow("magic-ip:"+platform.ClientIP(r), 20, 15*time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	if !auth.IsProtonEmail(email) && !in.Override {
		platform.Error(w, 400, "not_proton")
		return
	}
	tok := platform.RandomToken(32) // 256-bit single-use token, 15-minute TTL
	if err := s.Users.CreateToken(r.Context(), &store.AuthToken{Email: email, TokenHash: platform.SHA256(tok), Purpose: "proton_magic", ExpiresAt: s.Now().Add(15 * time.Minute)}); err != nil {
		s.fail(w, err)
		return
	}
	link := fmt.Sprintf("%s/auth/verify-email?purpose=proton_magic&token=%s", s.Cfg.PublicURL, url.QueryEscape(tok))
	body := "Sign in to ZENRES with this single-use link (valid 15 minutes):\n" + link + "\n\nAfter signing in you can save a passkey to Proton Pass."
	var key *crypto.Key
	if in.Encrypt {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if k, _ := auth.LookupWKD(ctx, email); k != nil {
			key = k
		}
	}
	if err := s.Mail.Send(r.Context(), email, "Your ZENRES sign-in link", body, key); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"status": "sent", "encrypted": key != nil})
}

func (s *Server) protonVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Token == "" {
		platform.Error(w, 400, "invalid_input")
		return
	}
	t, err := s.Users.ConsumeToken(r.Context(), platform.SHA256(in.Token), "proton_magic", s.Now())
	if err != nil {
		s.tokenError(w, err)
		return
	}
	u, err := s.Users.UserByEmail(r.Context(), t.Email)
	if errors.Is(err, store.ErrNotFound) {
		if s.Cfg.LegalApproved {
			platform.Error(w, 409, "legal_signup_requires_email")
			return
		}
		u = &store.User{Email: t.Email, EmailVerified: true, FullName: strings.Split(t.Email, "@")[0], AuthOrigin: "proton", Status: "active"}
		if err := s.Users.CreateUser(r.Context(), u); err != nil {
			s.fail(w, err)
			return
		}
	} else if err != nil {
		s.fail(w, err)
		return
	} else if !u.EmailVerified {
		// Possessing the mailbox proves ownership; mark verified.
		u.EmailVerified, u.Status = true, "active"
		_ = s.Users.UpdateUser(r.Context(), u)
	}
	if err := s.issueSession(r.Context(), w, r, u); err != nil {
		s.fail(w, err)
		return
	}
	pk, _ := s.Users.PasskeysForUser(r.Context(), u.ID)
	_, orgErr := s.Users.OrgForUser(r.Context(), u.ID)
	platform.JSON(w, 200, map[string]any{"offer_passkey": len(pk) == 0, "onboarded": orgErr == nil})
}

type waUser struct {
	u     *store.User
	creds []webauthn.Credential
}

func (w waUser) WebAuthnID() []byte                         { return w.u.ID[:] }
func (w waUser) WebAuthnName() string                       { return w.u.Email }
func (w waUser) WebAuthnDisplayName() string                { return w.u.FullName }
func (w waUser) WebAuthnCredentials() []webauthn.Credential { return w.creds }

func (s *Server) loadWAUser(ctx context.Context, u *store.User) (waUser, error) {
	pks, err := s.Users.PasskeysForUser(ctx, u.ID)
	if err != nil {
		return waUser{}, err
	}
	out := waUser{u: u}
	for _, p := range pks {
		var c webauthn.Credential
		if json.Unmarshal(p.PublicKey, &c) == nil {
			out.creds = append(out.creds, c)
		}
	}
	return out, nil
}

func (s *Server) passkeyRegisterStart(w http.ResponseWriter, r *http.Request) {
	u, _, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	wu, err := s.loadWAUser(r.Context(), u)
	if err != nil {
		s.fail(w, err)
		return
	}
	creation, sess, err := s.WebAuthn.BeginRegistration(wu,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(wu.creds).CredentialDescriptors()))
	if err != nil {
		s.fail(w, err)
		return
	}
	b, _ := json.Marshal(sess)
	s.KV.Put("wa:reg:"+u.ID.String(), b, 5*time.Minute)
	platform.JSON(w, 200, creation)
}

func (s *Server) passkeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	u, _, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	raw, ok := s.KV.Take("wa:reg:" + u.ID.String())
	if !ok {
		platform.Error(w, 400, "ceremony_expired")
		return
	}
	var sess webauthn.SessionData
	_ = json.Unmarshal(raw, &sess)
	wu, _ := s.loadWAUser(r.Context(), u)
	cred, err := s.WebAuthn.FinishRegistration(wu, sess, r)
	if err != nil {
		platform.Error(w, 400, "attestation_invalid")
		return
	}
	cj, _ := json.Marshal(cred)
	if err := s.Users.AddPasskey(r.Context(), &store.Passkey{UserID: u.ID, CredentialID: cred.ID, PublicKey: cj, SignCount: cred.Authenticator.SignCount, Label: "Proton Pass / Device Key"}); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]string{"status": "passkey_saved"})
}

func (s *Server) passkeyLoginStart(w http.ResponseWriter, r *http.Request) {
	if !s.Limiter.Allow("pk:"+platform.ClientIP(r), 20, 15*time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	assertion, sess, err := s.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		s.fail(w, err)
		return
	}
	id := platform.RandomToken(16)
	b, _ := json.Marshal(sess)
	s.KV.Put("wa:login:"+id, b, 5*time.Minute)
	http.SetCookie(w, &http.Cookie{Name: "zv_wa", Value: id, Path: "/", MaxAge: 300, HttpOnly: true, Secure: s.Cfg.CookieSecure, SameSite: http.SameSiteStrictMode})
	platform.JSON(w, 200, assertion)
}

func (s *Server) passkeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("zv_wa")
	if err != nil {
		platform.Error(w, 400, "ceremony_expired")
		return
	}
	raw, ok := s.KV.Take("wa:login:" + c.Value)
	if !ok {
		platform.Error(w, 400, "ceremony_expired")
		return
	}
	var sess webauthn.SessionData
	_ = json.Unmarshal(raw, &sess)
	var owner *store.User
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		pk, err := s.Users.PasskeyByCredentialID(r.Context(), rawID)
		if err != nil {
			return nil, err
		}
		u, err := s.Users.UserByID(r.Context(), pk.UserID)
		if err != nil {
			return nil, err
		}
		owner = u
		return s.loadWAUser(r.Context(), u)
	}
	cred, err := s.WebAuthn.FinishDiscoverableLogin(handler, sess, r)
	if err != nil || owner == nil {
		platform.Error(w, 401, "passkey_invalid")
		return
	}
	if pk, err := s.Users.PasskeyByCredentialID(r.Context(), cred.ID); err == nil {
		pk.SignCount = cred.Authenticator.SignCount
		pk.PublicKey, _ = json.Marshal(cred)
		_ = s.Users.UpdatePasskey(r.Context(), pk)
	}
	if err := s.issueSession(r.Context(), w, r, owner); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"status": "ok"})
}

/* ------------------------------------------------------------------ */
/* Google / Apple OIDC with PKCE + safe account linking                */
/* ------------------------------------------------------------------ */

type oauthState struct {
	Provider string `json:"p"`
	Verifier string `json:"v"`
	Nonce    string `json:"n"`
	Next     string `json:"next"`
}

func (s *Server) oauthStart(w http.ResponseWriter, r *http.Request) {
	p, ok := s.Providers[chi.URLParam(r, "provider")]
	if !ok {
		notFound(w)
		return
	}
	verifier := platform.RandomToken(48) // 64 chars, RFC 7636 compliant
	st := oauthState{Provider: p.Name(), Verifier: verifier, Nonce: platform.RandomToken(16), Next: safeNext(r.URL.Query().Get("next"))}
	state := platform.RandomToken(24)
	b, _ := json.Marshal(st)
	s.KV.Put("oauth:"+state, b, 10*time.Minute)
	u, err := p.AuthURL(state, st.Nonce, auth.PKCEChallenge(verifier))
	if err != nil {
		platform.Error(w, 503, "provider_unavailable")
		return
	}
	http.Redirect(w, r, u, http.StatusFound)
}

func safeNext(n string) string {
	if strings.HasPrefix(n, "/") && !strings.HasPrefix(n, "//") {
		return n
	}
	return ""
}

func (s *Server) oauthCallback(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	state, code := r.Form.Get("state"), r.Form.Get("code")
	raw, ok := s.KV.Take("oauth:" + state)
	if !ok || code == "" {
		http.Redirect(w, r, s.Cfg.PublicURL+"/login?error=oauth_state", http.StatusSeeOther)
		return
	}
	var st oauthState
	_ = json.Unmarshal(raw, &st)
	p := s.Providers[st.Provider]
	if p == nil || st.Provider != chi.URLParam(r, "provider") {
		http.Redirect(w, r, s.Cfg.PublicURL+"/login?error=oauth_state", http.StatusSeeOther)
		return
	}
	claims, err := p.Exchange(r.Context(), code, st.Verifier, st.Nonce)
	if err != nil {
		slog.Warn("oauth exchange failed", "provider", st.Provider, "err", err)
		http.Redirect(w, r, s.Cfg.PublicURL+"/login?error=oauth_failed", http.StatusSeeOther)
		return
	}
	if st.Provider == "apple" && claims.Name == "" {
		claims.Name = auth.AppleUserName(r.PostForm) // only present on first authorisation
	}
	dest, err := s.resolveOAuthUser(r.Context(), w, r, st.Provider, claims)
	if err != nil {
		s.fail(w, err)
		return
	}
	if st.Next != "" && strings.HasSuffix(dest, "/dashboard") {
		dest = s.Cfg.PublicURL + st.Next
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// resolveOAuthUser never silently merges an OAuth login into an existing password account.
func (s *Server) resolveOAuthUser(ctx context.Context, w http.ResponseWriter, r *http.Request, provider string, c *auth.Claims) (string, error) {
	if id, err := s.Users.IdentityBySubject(ctx, provider, c.Subject); err == nil {
		u, err := s.Users.UserByID(ctx, id.UserID)
		if err != nil {
			return "", err
		}
		if err := s.issueSession(ctx, w, r, u); err != nil {
			return "", err
		}
		return s.afterLogin(ctx, u), nil
	}
	existing, err := s.Users.UserByEmail(ctx, strings.ToLower(c.Email))
	if err == nil {
		// Existing account with the same email: require confirmation before linking.
		tok := platform.RandomToken(24)
		b, _ := json.Marshal(map[string]string{"provider": provider, "subject": c.Subject, "email": c.Email, "user_id": existing.ID.String()})
		s.KV.Put("link:"+tok, b, 10*time.Minute)
		return s.Cfg.PublicURL + "/login?link=" + tok + "&provider=" + provider, nil
	}
	if s.Cfg.LegalApproved {
		return s.Cfg.PublicURL + "/signup?error=legal_signup_requires_email", nil
	}
	name := strings.TrimSpace(c.Name)
	if name == "" {
		name = strings.Split(c.Email, "@")[0]
	}
	u := &store.User{Email: strings.ToLower(c.Email), EmailVerified: true, FullName: name, AuthOrigin: provider, Status: "active"}
	if err := s.Users.CreateUser(ctx, u); err != nil {
		return "", err
	}
	if err := s.Users.LinkIdentity(ctx, store.OAuthIdentity{UserID: u.ID, Provider: provider, ProviderSubject: c.Subject, ProviderEmail: c.Email}); err != nil {
		return "", err
	}
	if err := s.issueSession(ctx, w, r, u); err != nil {
		return "", err
	}
	return s.Cfg.PublicURL + "/onboarding", nil
}

func (s *Server) afterLogin(ctx context.Context, u *store.User) string {
	if _, err := s.Users.OrgForUser(ctx, u.ID); err == nil {
		return s.Cfg.AppURL + "/dashboard"
	}
	return s.Cfg.PublicURL + "/onboarding"
}

func (s *Server) linkConfirm(w http.ResponseWriter, r *http.Request) {
	var in struct {
		LinkToken string `json:"link_token"`
		Password  string `json:"password"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	raw, ok := s.KV.Take("link:" + in.LinkToken)
	if !ok {
		platform.Error(w, 410, "link_expired")
		return
	}
	var p map[string]string
	_ = json.Unmarshal(raw, &p)
	uid, _ := uuid.Parse(p["user_id"])
	u, err := s.Users.UserByID(r.Context(), uid)
	if err != nil || !auth.VerifyOrDummy(in.Password, u.PasswordHash) {
		platform.Error(w, 401, "invalid_credentials")
		return
	}
	if err := s.Users.LinkIdentity(r.Context(), store.OAuthIdentity{UserID: u.ID, Provider: p["provider"], ProviderSubject: p["subject"], ProviderEmail: p["email"]}); err != nil && !errors.Is(err, store.ErrConflict) {
		s.fail(w, err)
		return
	}
	if err := s.issueSession(r.Context(), w, r, u); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"status": "linked", "redirect": s.afterLogin(r.Context(), u)})
}

/* ------------------------------------------------------------------ */
/* onboarding                                                          */
/* ------------------------------------------------------------------ */

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func (s *Server) onboardingComplete(w http.ResponseWriter, r *http.Request) {
	u, _, err := s.currentUser(r)
	if err != nil {
		platform.Error(w, 401, "unauthenticated")
		return
	}
	var in struct {
		Role        string  `json:"role"`
		OrgName     string  `json:"org_name"`
		CountryCode string  `json:"country_code"`
		TaxID       *string `json:"gstin_or_eori"`
		Sector      string  `json:"sector"`
		Plan        string  `json:"plan"`
		Sandbox     bool    `json:"sandbox"`
	}
	if platform.DecodeLoose(r, &in) != nil || len(strings.TrimSpace(in.OrgName)) < 2 || len(in.CountryCode) != 2 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	orgType := map[string]string{"msme": "msme", "eu_importer": "eu_importer", "transporter": "transporter", "buyer": "buyer"}[in.Role]
	if orgType == "" {
		platform.Error(w, 400, "invalid_role")
		return
	}
	tier := map[string]string{"starter": "starter", "growth": "growth", "importer_hub": "importer_hub", "transporter_free": "starter"}[in.Plan]
	if tier == "" {
		tier = "starter"
	}
	tax := ""
	if in.TaxID != nil {
		tax = strings.ToUpper(strings.TrimSpace(*in.TaxID))
	}
	slug := strings.Trim(slugRe.ReplaceAllString(strings.ToLower(in.OrgName), "-"), "-")
	if len(slug) > 32 {
		slug = slug[:32]
	}
	org := &store.Org{Name: strings.TrimSpace(in.OrgName), OrgType: orgType, CountryCode: strings.ToUpper(in.CountryCode), TaxID: tax, Tier: tier, Sector: in.Sector, InboundSlug: slug + "-" + platform.RandomHex(2)}
	if err := s.Users.CreateOrg(r.Context(), org, u.ID, "owner"); err != nil {
		s.fail(w, err)
		return
	}
	// Rotate the session on privilege elevation (now an org owner).
	if err := s.issueSession(r.Context(), w, r, u); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 201, map[string]any{"org_id": org.ID, "tier": tier, "redirect": s.Cfg.AppURL + "/dashboard"})
}

/* ------------------------------------------------------------------ */
/* hidden PGP vault                                                    */
/* ------------------------------------------------------------------ */

// The public SvelteKit gate authenticates this header with X-Zevra-Gate.
// Vercel replaces X-Forwarded-For on the server-to-server hop, so the
// original caller must travel in a separate header.
func vaultClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Zenres-Client-IP")); net.ParseIP(ip) != nil {
		return ip
	}
	return platform.ClientIP(r)
}

func (s *Server) vaultChallenge(w http.ResponseWriter, r *http.Request) {
	ip := vaultClientIP(r)
	if !s.Limiter.Allow("vault-ch:"+ip, 5, 15*time.Minute) {
		platform.Error(w, 429, "rate_limited")
		return
	}
	var in struct {
		Identifier string `json:"identifier"`
	}
	if platform.DecodeLoose(r, &in) != nil || len(in.Identifier) < 3 || len(in.Identifier) > 80 {
		platform.Error(w, 400, "invalid_input")
		return
	}
	out, err := s.Vault.IssueChallenge(r.Context(), in.Identifier, ip)
	if err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, out)
}

func (s *Server) vaultVerify(w http.ResponseWriter, r *http.Request) {
	var in pgp.VerifyIn
	if platform.DecodeLoose(r, &in) != nil {
		platform.Error(w, 400, "invalid_input")
		return
	}
	out, err := s.Vault.Verify(r.Context(), in, vaultClientIP(r), r.UserAgent())
	if err != nil {
		status := 401
		if errors.Is(err, pgp.ErrChallengeExpired) || errors.Is(err, pgp.ErrTooManyAttempts) || errors.Is(err, pgp.ErrReplay) {
			status = 410
		}
		platform.Error(w, status, err.Error())
		return
	}
	code := platform.RandomToken(32)
	if err := s.Vault.Store.CreateHandoff(r.Context(), platform.SHA256(code), out.Token, s.Now().Add(60*time.Second)); err != nil {
		s.fail(w, err)
		return
	}
	platform.JSON(w, 200, map[string]any{"handoff_url": s.Cfg.AppURL + "/ops-vault/handoff?code=" + code})
}

func (s *Server) vaultExchange(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code string `json:"code"`
		IP   string `json:"ip"`
		UA   string `json:"ua"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		notFound(w)
		return
	}
	tok, err := s.Vault.Store.ConsumeHandoff(r.Context(), platform.SHA256(in.Code), s.Now())
	if err != nil {
		notFound(w)
		return
	}
	if _, err := s.Vault.ValidateSession(r.Context(), tok, in.IP, in.UA); err != nil {
		notFound(w)
		return
	}
	platform.JSON(w, 200, map[string]string{"token": tok})
}

func (s *Server) vaultSession(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
		IP    string `json:"ip"`
		UA    string `json:"ua"`
	}
	if platform.DecodeLoose(r, &in) != nil {
		notFound(w)
		return
	}
	k, err := s.Vault.ValidateSession(r.Context(), in.Token, in.IP, in.UA)
	if err != nil {
		notFound(w)
		return
	}
	platform.JSON(w, 200, map[string]any{"handle": k.Handle, "fingerprint": k.Fingerprint, "role": k.Role})
}

func (s *Server) vaultLogout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
	}
	if platform.DecodeLoose(r, &in) != nil || in.Token == "" {
		notFound(w)
		return
	}
	if err := s.Vault.Store.RevokeAdminSession(r.Context(), platform.SHA256(in.Token)); err != nil {
		s.fail(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	slog.Error("auth-api", "err", err)
	platform.Error(w, 500, "internal_error")
}
