// Command api runs the Zevra core platform API (app.zevra.io backend).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadysbasement/zevra/services/core-api/internal/api"
	"github.com/shadysbasement/zevra/services/core-api/internal/twin"
	"github.com/shadysbasement/zevra/services/internal/platform"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("db config", "err", err)
		os.Exit(1)
	}
	// Connect as the RLS-bound application role so policies always apply.
	if role := platform.Env("DB_APP_ROLE", "zevra_app"); role != "" {
		cfg.ConnConfig.RuntimeParams["role"] = role
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	srv := &api.Server{
		Cfg: api.Config{
			AppOrigin:      platform.Env("PUBLIC_APP_URL", "http://localhost:5174"),
			WorkerURL:      platform.Env("WORKER_URL", "http://localhost:8090"),
			AuthAPIURL:     platform.Env("AUTH_API_URL", "http://localhost:8081"),
			InternalSecret: os.Getenv("INTERNAL_API_SECRET"),
			InboundSecret:  os.Getenv("MAIL_INBOUND_SECRET"),
			TypstBin:       platform.Env("TYPST_BIN", "typst"),
			TemplatesDir:   platform.Env("TYPST_TEMPLATES", "templates/typst"),
		},
		DB: pool, Broker: twin.NewBroker(), Limiter: platform.NewMemLimiter(),
	}
	addr := platform.Env("CORE_API_ADDR", ":8082")
	httpSrv := &http.Server{Addr: addr, Handler: srv.Routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		slog.Info("zevra-core-api listening", "addr", addr)
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
