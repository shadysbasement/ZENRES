// Package rls runs every tenant query inside a transaction with SET LOCAL app.current_org_id,
// so PostgreSQL row-level security policies isolate tenants even if a query forgets a WHERE clause.
package rls

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Opt func(ctx context.Context, tx pgx.Tx) error

// Marketplace enables the corridor-wide read policy on open freight listings.
func Marketplace(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `SELECT set_config('app.marketplace', 'on', true)`)
	return err
}

// PublicResolver enables the read-only QR passport policy (no org context).
func PublicResolver(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `SELECT set_config('app.public_resolver', 'on', true)`)
	return err
}

func Tx(ctx context.Context, db *pgxpool.Pool, org uuid.UUID, fn func(pgx.Tx) error, opts ...Opt) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if org != uuid.Nil {
		if _, err := tx.Exec(ctx, `SELECT set_config('app.current_org_id', $1, true)`, org.String()); err != nil {
			return err
		}
	}
	for _, o := range opts {
		if err := o(ctx, tx); err != nil {
			return err
		}
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
