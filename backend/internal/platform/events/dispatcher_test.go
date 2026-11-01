package events

import (
	"context"
	"errors"
	"os"
	"testing"

	kernelEvents "github.com/Langa2007/POS/backend/internal/kernel/events"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

func TestTransactionalOutboxDispatchAndRetry(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL outbox integration test")
	}
	ctx := context.Background()
	store, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	tenantID, _ := ids.NewUUIDv7()
	businessID, _ := ids.NewUUIDv7()
	event, err := kernelEvents.New(tenantID, "test.business.created", map[string]string{"business_id": businessID})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,'outbox test')`, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'Outbox test business')`, businessID, tenantID); err != nil {
			return err
		}
		return kernelEvents.Write(ctx, tx, event)
	}); err != nil {
		t.Fatal(err)
	}
	attempt := 0
	registry := NewRegistry()
	if err := registry.Register(event.Type, func(ctx context.Context, tx pgx.Tx, e kernelEvents.Event) error {
		attempt++
		if _, err := tx.Exec(ctx, `INSERT INTO audit_log(id,tenant_id,action,entity_type,entity_id) VALUES($1,$2,'test.dispatch','business',$3)`, mustID(t), tenantID, businessID); err != nil {
			return err
		}
		if attempt == 1 {
			return errors.New("simulated transient failure")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dispatcher := NewDispatcher(store, registry)
	if count, err := dispatcher.DispatchTenant(ctx, tenantID, 10); err == nil || count != 0 {
		t.Fatalf("first dispatch=(%d,%v), expected retryable error", count, err)
	}
	var auditCount, attempts int
	if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE tenant_id=$1 AND entity_id=$2`, tenantID, businessID).Scan(&auditCount); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT attempts FROM outbox WHERE id=$1`, event.ID).Scan(&attempts)
	}); err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 || attempts != 1 {
		t.Fatalf("failed handler effects/attempts=(%d,%d), expected (0,1)", auditCount, attempts)
	}
	if count, err := dispatcher.DispatchTenant(ctx, tenantID, 10); err != nil || count != 1 {
		t.Fatalf("second dispatch=(%d,%v)", count, err)
	}
	if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE tenant_id=$1 AND entity_id=$2`, tenantID, businessID).Scan(&auditCount); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT attempts FROM outbox WHERE id=$1 AND delivered_at IS NOT NULL`, event.ID).Scan(&attempts)
	}); err != nil {
		t.Fatal(err)
	}
	if auditCount != 1 || attempts != 2 {
		t.Fatalf("successful handler effects/attempts=(%d,%d), expected (1,2)", auditCount, attempts)
	}
}
func mustID(t *testing.T) string {
	t.Helper()
	id, err := ids.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
