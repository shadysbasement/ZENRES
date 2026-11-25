package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// OIDCProvider implements Google and Apple Sign-In with Authorization Code + PKCE (S256).
// ID tokens are verified against the provider JWKS (aud, iss, exp, signature) by go-oidc.
type OIDCProvider struct {
	name     string
	cfg      oauth2.Config
	verifier *oidc.IDTokenVerifier
	apple    *appleSecret
	extra    []oauth2.AuthCodeOption
}

func (p *OIDCProvider) Name() string { return p.name }

func NewGoogle(ctx context.Context, clientID, clientSecret, redirect string) (*OIDCProvider, error) {
	prov, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	return &OIDCProvider{
		name:     "google",
		cfg:      oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirect, Endpoint: prov.Endpoint(), Scopes: []string{oidc.ScopeOpenID, "email", "profile"}},
		verifier: prov.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

// appleSecret mints the ES256 client-secret JWT Apple requires, rotating it before expiry.
type appleSecret struct {
	teamID, keyID, clientID string
	key                     *ecdsa.PrivateKey
	cached                  string
	exp                     time.Time
}

func (a *appleSecret) get() (string, error) {
	if a.cached != "" && time.Until(a.exp) > time.Hour {
		return a.cached, nil
	}
	now := time.Now()
	exp := now.Add(24 * time.Hour * 30)
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": a.teamID, "iat": now.Unix(), "exp": exp.Unix(), "aud": "https://appleid.apple.com", "sub": a.clientID,
	})
	tok.Header["kid"] = a.keyID
	s, err := tok.SignedString(a.key)
	if err != nil {
		return "", err
	}
	a.cached, a.exp = s, exp
	return s, nil
}

func NewApple(ctx context.Context, clientID, teamID, keyID, privateKeyPEM, redirect string) (*OIDCProvider, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("apple: invalid private key PEM")
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apple: key is not ECDSA P-256")
	}
	prov, err := oidc.NewProvider(ctx, "https://appleid.apple.com")
	if err != nil {
		return nil, err
	}
	return &OIDCProvider{
		name:     "apple",
		cfg:      oauth2.Config{ClientID: clientID, RedirectURL: redirect, Endpoint: prov.Endpoint(), Scopes: []string{"name", "email"}},
		verifier: prov.Verifier(&oidc.Config{ClientID: clientID}),
		apple:    &appleSecret{teamID: teamID, keyID: keyID, clientID: clientID, key: ec},
		extra:    []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("response_mode", "form_post")},
	}, nil
}

func (p *OIDCProvider) AuthURL(state, nonce, challenge string) (string, error) {
	opts := append([]oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}, p.extra...)
	return p.cfg.AuthCodeURL(state, opts...), nil
}

func (p *OIDCProvider) Exchange(ctx context.Context, code, verifier, nonce string) (*Claims, error) {
	cfg := p.cfg
	if p.apple != nil {
		s, err := p.apple.get()
		if err != nil {
			return nil, err
		}
		cfg.ClientSecret = s
	}
	tok, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token")
	}
	idt, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, err
	}
	if idt.Nonce != nonce {
		return nil, errors.New("nonce mismatch")
	}
	var c struct {
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"` // Apple sends "true" as a string
		Name          string `json:"name"`
	}
	if err := idt.Claims(&c); err != nil {
		return nil, err
	}
	verified := c.EmailVerified == true || c.EmailVerified == "true"
	if !verified {
		return nil, errors.New("email_not_verified")
	}
	return &Claims{Subject: idt.Subject, Email: c.Email, EmailVerified: verified, Name: c.Name}, nil
}

// AppleUserName parses the `user` JSON Apple posts on FIRST authorisation only.
func AppleUserName(form url.Values) string {
	type n struct {
		Name struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"name"`
	}
	var v n
	if raw := form.Get("user"); raw != "" {
		_ = jsonUnmarshal([]byte(raw), &v)
	}
	if v.Name.FirstName == "" && v.Name.LastName == "" {
		return ""
	}
	return v.Name.FirstName + " " + v.Name.LastName
}
