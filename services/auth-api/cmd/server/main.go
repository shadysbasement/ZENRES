// Command server runs the Zevra auth & identity API (zevra-auth-api).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/shadysbasement/zevra/services/auth-api/internal/auth"
	"github.com/shadysbasement/zevra/services/auth-api/internal/httpapi"
	"github.com/shadysbasement/zevra/services/auth-api/internal/pgp"
	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publicURL := platform.Env("PUBLIC_SITE_URL", "http://localhost:5173")
	appURL := platform.Env("PUBLIC_APP_URL", "http://localhost:5174")

	var users store.Users
	var vaultStore store.Vault
	mem := store.NewMemory()
	var kv store.KV = mem // swap for Redis in multi-replica deployments
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" && platform.Env("STORE", "postgres") == "postgres" {
		pg, err := store.NewPostgres(ctx, dsn)
		if err != nil {
			slog.Error("postgres", "err", err)
			os.Exit(1)
		}
		users, vaultStore = pg, pg
	} else {
		slog.Warn("STORE=memory — data is ephemeral (development only)")
		users, vaultStore = mem, mem
		// Development convenience: preload one admin public key so the hidden vault can be tested locally.
		if f := os.Getenv("ADMIN_DEV_KEY_FILE"); f != "" {
			if raw, err := os.ReadFile(f); err == nil {
				if fpr, k, err := pgp.FingerprintOf(string(raw)); err == nil {
					pub, _ := k.GetArmoredPublicKey()
					_ = mem.InsertAdminKey(ctx, &store.AdminKey{Handle: "zv-dev-admin", Fingerprint: fpr, ArmoredPublicKey: pub, Role: "super_admin", Status: "active"})
					slog.Info("dev admin key loaded", "fingerprint", fpr)
				}
			}
		}
	}

	vault, err := pgp.New(vaultStore, platform.Env("VAULT_DOMAIN", "zevra.io"))
	if err != nil {
		slog.Error("vault", "err", err)
		os.Exit(1)
	}

	rpID := platform.Env("WEBAUTHN_RP_ID", "localhost")
	wa, err := webauthn.New(&webauthn.Config{RPID: rpID, RPDisplayName: "Zevra", RPOrigins: strings.Split(platform.Env("WEBAUTHN_ORIGINS", publicURL+","+appURL), ",")})
	if err != nil {
		slog.Error("webauthn", "err", err)
		os.Exit(1)
	}

	var mailer auth.Mailer = auth.LogMailer{}
	if addr := os.Getenv("SMTP_ADDR"); addr != "" {
		mailer = auth.SMTPMailer{Addr: addr, From: platform.Env("SMTP_FROM", "no-reply@zevra.io"), User: os.Getenv("SMTP_USER"), Pass: os.Getenv("SMTP_PASS")}
	}

	providers := map[string]auth.IdentityProvider{}
	callback := platform.Env("AUTH_PUBLIC_CALLBACK_BASE", publicURL+"/api/auth/oauth")
	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		if g, err := auth.NewGoogle(ctx, id, os.Getenv("GOOGLE_CLIENT_SECRET"), callback+"/google/callback"); err == nil {
			providers["google"] = g
		} else {
			slog.Warn("google oidc disabled", "err", err)
		}
	}
	if id := os.Getenv("APPLE_CLIENT_ID"); id != "" {
		if a, err := auth.NewApple(ctx, id, os.Getenv("APPLE_TEAM_ID"), os.Getenv("APPLE_KEY_ID"), os.Getenv("APPLE_PRIVATE_KEY"), callback+"/apple/callback"); err == nil {
			providers["apple"] = a
		} else {
			slog.Warn("apple oidc disabled", "err", err)
		}
	}
	if os.Getenv("ENABLE_PROTON_OIDC") == "true" {
		providers["proton_oidc"] = auth.ProtonOIDCStub{Enabled: true}
	}

	srv := &httpapi.Server{
		Cfg: httpapi.Config{
			PublicURL:      publicURL,
			AppURL:         appURL,
			CookieDomain:   os.Getenv("COOKIE_DOMAIN"),
			CookieSecure:   platform.Env("COOKIE_SECURE", "true") != "false",
			GateSecret:     os.Getenv("ADMIN_GATE_SECRET"),
			InternalSecret: os.Getenv("INTERNAL_API_SECRET"),
		},
		Users: users, KV: kv, Vault: vault, Limiter: platform.NewMemLimiter(), Mail: mailer, Providers: providers, WebAuthn: wa, Now: time.Now,
	}

	addr := platform.Env("AUTH_API_ADDR", ":8081")
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	httpSrv := &http.Server{Addr: addr, Handler: srv.Routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		slog.Info("zevra-auth-api listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shut)
}
