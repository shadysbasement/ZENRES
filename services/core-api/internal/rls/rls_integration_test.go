package rls

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run against a disposable PostgreSQL database with migrations 001 and 002 applied.
// The test deliberately uses the same zevra_app role as the core API.
func TestTenantPolicyIsolation(t *testing.T) {
	dsn := os.Getenv("RLS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set RLS_TEST_DATABASE_URL to a disposable migrated test database")
	}
	adminCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(adminCfg.ConnConfig.Database, "test") {
		t.Fatal("RLS_TEST_DATABASE_URL must name a disposable database containing 'test'")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.NewWithConfig(ctx, adminCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	appCfg := adminCfg.Copy()
	appCfg.ConnConfig.RuntimeParams["role"] = "zevra_app"
	app, err := pgxpool.NewWithConfig(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	createOrg := func(name string) (uuid.UUID, uuid.UUID) {
		t.Helper()
		var orgID, plantID uuid.UUID
		slug := "rls-test-" + uuid.NewString()
		if err := admin.QueryRow(ctx, `INSERT INTO organizations (name,org_type,inbound_mail_slug) VALUES ($1,'msme',$2) RETURNING id`, name, slug).Scan(&orgID); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_, _ = admin.Exec(cleanupCtx, `DELETE FROM organizations WHERE id=$1`, orgID)
		})
		if err := admin.QueryRow(ctx, `INSERT INTO plants (org_id,name,city) VALUES ($1,$2,'Pune') RETURNING id`, orgID, name).Scan(&plantID); err != nil {
			t.Fatal(err)
		}
		return orgID, plantID
	}
	orgA, plantA := createOrg("RLS test A")
	orgB, plantB := createOrg("RLS test B")

	checkVisible := func(org, want uuid.UUID) {
		t.Helper()
		err := Tx(ctx, app, org, func(tx pgx.Tx) error {
			rows, err := tx.Query(ctx, `SELECT id FROM plants WHERE id IN ($1,$2)`, plantA, plantB)
			if err != nil {
				return err
			}
			defer rows.Close()
			var got []uuid.UUID
			for rows.Next() {
				var id uuid.UUID
				if err := rows.Scan(&id); err != nil {
					return err
				}
				got = append(got, id)
			}
			if err := rows.Err(); err != nil {
				return err
			}
			if want == uuid.Nil && len(got) != 0 || want != uuid.Nil && (len(got) != 1 || got[0] != want) {
				t.Errorf("org %s saw plant IDs %v; want %s", org, got, want)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	checkVisible(orgA, plantA)
	checkVisible(orgB, plantB)
	checkVisible(uuid.Nil, uuid.Nil)

	if err := Tx(ctx, app, orgA, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `UPDATE plants SET city='Delhi' WHERE id=$1`, plantB)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 0 {
			t.Errorf("cross-tenant update affected %d rows", result.RowsAffected())
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := Tx(ctx, app, orgA, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO plants (org_id,name,city) VALUES ($1,'Forbidden','Pune')`, orgB)
		return err
	}); err == nil {
		t.Fatal("cross-tenant insert unexpectedly succeeded")
	}
	var city string
	if err := admin.QueryRow(ctx, `SELECT city FROM plants WHERE id=$1`, plantB).Scan(&city); err != nil || city != "Pune" {
		t.Fatalf("other tenant plant changed: city=%q err=%v", city, err)
	}
	var setting *string
	if err := app.QueryRow(ctx, `SELECT current_setting('app.current_org_id', true)`).Scan(&setting); err != nil {
		t.Fatal(err)
	}
	if setting != nil && *setting != "" {
		t.Fatalf("transaction-local org leaked to next request: %q", *setting)
	}
}
