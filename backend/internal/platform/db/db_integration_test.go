package db

import (
	"context"
	"os"
	"testing"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
)

func TestTenantRLSAndMigrations(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL migration/RLS integration test")
	}
	ctx := context.Background()
	store, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	var unprotected string
	err = store.pool.QueryRow(ctx, `SELECT c.table_name FROM information_schema.columns c JOIN pg_class p ON p.relname=c.table_name JOIN pg_namespace n ON n.oid=p.relnamespace WHERE c.table_schema=current_schema() AND c.column_name='tenant_id' AND n.nspname=current_schema() AND NOT p.relrowsecurity LIMIT 1`).Scan(&unprotected)
	if err != pgx.ErrNoRows {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("tenant table %q has RLS disabled", unprotected)
	}
	tenantA, _ := ids.NewUUIDv7()
	tenantB, _ := ids.NewUUIDv7()
	business, _ := ids.NewUUIDv7()
	for _, tenantID := range []string{tenantA, tenantB} {
		if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,$2)`, tenantID, "integration test")
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.TenantTx(ctx, tenantA, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'tenant A business')`, business, tenantA)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.pool.QueryRow(ctx, `SELECT count(*) FROM businesses WHERE id=$1`, business).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("unscoped query returned %d rows", count)
	}
	if err := store.TenantTx(ctx, tenantB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM businesses WHERE id=$1`, business).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			t.Fatalf("tenant B saw %d tenant A rows", count)
		}
		_, err := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'cross-tenant write')`, mustNewID(t), tenantA)
		if err == nil {
			t.Fatal("cross-tenant insert succeeded")
		}
		return err
	}); err == nil {
		t.Fatal("tenant transaction did not reject cross-tenant insert")
	}
	if err := store.TenantTx(ctx, tenantB, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO branches(id,tenant_id,business_id,name) VALUES($1,$2,$3,'cross-tenant branch')`, mustNewID(t), tenantB, business)
		if err == nil {
			t.Fatal("cross-tenant business relationship succeeded")
		}
		return err
	}); err == nil {
		t.Fatal("tenant transaction did not reject cross-tenant relationship")
	}
}

func mustNewID(t *testing.T) string {
	t.Helper()
	id, err := ids.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
