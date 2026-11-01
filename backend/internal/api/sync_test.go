package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

func TestSyncPushAppliesCashSaleAtomicallyAndIdempotently(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to exercise offline sale sync against PostgreSQL")
	}
	ctx := context.Background()
	store, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	tenantID, _ := ids.NewUUIDv7()
	businessID, _ := ids.NewUUIDv7()
	branchID, _ := ids.NewUUIDv7()
	userID, _ := ids.NewUUIDv7()
	roleID, _ := ids.NewUUIDv7()
	shiftID, _ := ids.NewUUIDv7()
	itemID, _ := ids.NewUUIDv7()
	hash, err := auth.HashPassword("sync test password")
	if err != nil {
		t.Fatal(err)
	}
	err = store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,'sync integration')`, tenantID); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'Sync test')`, businessID, tenantID); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO branches(id,tenant_id,business_id,name) VALUES($1,$2,$3,'Branch')`, branchID, tenantID, businessID); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO users(id,tenant_id,email,password_hash) VALUES($1,$2,$3,$4)`, userID, tenantID, tenantID+"@sync.test", hash); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO roles(id,tenant_id,name) VALUES($1,$2,'owner')`, roleID, tenantID); e != nil {
			return e
		}
		for _, permission := range []string{"sync.push", "sync.pull"} {
			if _, e := tx.Exec(ctx, `INSERT INTO permissions(id,description) VALUES($1,'sync test') ON CONFLICT DO NOTHING`, permission); e != nil {
				return e
			}
			if _, e := tx.Exec(ctx, `INSERT INTO role_permissions(tenant_id,role_id,permission_id) VALUES($1,$2,$3)`, tenantID, roleID, permission); e != nil {
				return e
			}
		}
		if _, e := tx.Exec(ctx, `INSERT INTO user_roles(tenant_id,user_id,role_id) VALUES($1,$2,$3)`, tenantID, userID, roleID); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO shifts(id,tenant_id,branch_id,user_id,opening_float_minor,status) VALUES($1,$2,$3,$4,0,'open')`, shiftID, tenantID, branchID, userID); e != nil {
			return e
		}
		if _, e := tx.Exec(ctx, `INSERT INTO catalog_items(id,tenant_id,name,sku,kind,price_minor,cost_minor,stock_on_hand) VALUES($1,$2,'Tea','SYNC-TEA','stocked',1250,500,1)`, itemID, tenantID); e != nil {
			return e
		}
		movementID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(ctx, `INSERT INTO stock_movements(id,tenant_id,item_id,quantity_delta,reason) VALUES($1,$2,$3,1,'opening_balance')`, movementID, tenantID, itemID)
		return e
	})
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewTokenManager([]byte(strings.Repeat("s", 32)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	srv := New(store, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	token, err := tokens.Issue(tenantID, userID)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(time.Second)
	clientID := "10000000-0000-4000-8000-000000000001"
	saleID, _ := ids.NewUUIDv7()
	secondClientID := "20000000-0000-4000-8000-000000000001"
	secondSaleID, _ := ids.NewUUIDv7()
	operation := func(client, sale string, qty int) map[string]any {
		return map[string]any{"client_id": client, "entity_type": "sale", "entity_id": sale, "operation": "create", "client_occurred_at": now, "payload": map[string]any{"branch_id": branchID, "shift_id": shiftID, "receipt_number": "T-ABCDEF1234-00000001", "offline_since": now.Add(-time.Hour), "received_minor": 1250, "lines": []any{map[string]any{"item_id": itemID, "quantity": qty, "unit_price_minor": 1250, "discount_minor": 0}}}}
	}
	body, _ := json.Marshal(map[string]any{"operations": []any{operation(clientID, saleID, 1), operation(secondClientID, secondSaleID, 1)}})
	push := func(raw []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/push", strings.NewReader(string(raw)))
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		return res
	}
	response := push(body)
	if response.Code != 200 {
		t.Fatalf("sync push status=%d body=%s", response.Code, response.Body.String())
	}
	var result struct {
		Acks []struct {
			ClientID string `json:"client_id"`
			Status   string `json:"status"`
			Error    *struct {
				Code string `json:"code"`
			} `json:"error"`
		} `json:"acks"`
	}
	if err = json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Acks) != 2 || result.Acks[0].Status != "accepted" || result.Acks[1].Status != "rejected" || result.Acks[1].Error == nil || result.Acks[1].Error.Code != "sync.insufficient_stock" {
		t.Fatalf("acks=%+v response=%s", result.Acks, response.Body.String())
	}
	replay := push(mustMarshal(map[string]any{"operations": []any{operation(clientID, saleID, 1)}}))
	if replay.Code != 200 {
		t.Fatalf("replay status=%d body=%s", replay.Code, replay.Body.String())
	}
	pullReq := httptest.NewRequest(http.MethodGet, "/api/v1/sync/pull?cursor=0&limit=100", nil)
	pullReq.Header.Set("Authorization", "Bearer "+token)
	pullRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(pullRes, pullReq)
	if pullRes.Code != 200 {
		t.Fatalf("pull status=%d body=%s", pullRes.Code, pullRes.Body.String())
	}
	var pull struct {
		Changes []struct {
			EntityType string         `json:"entity_type"`
			EntityID   string         `json:"entity_id"`
			Payload    map[string]any `json:"payload"`
		} `json:"changes"`
	}
	if err = json.Unmarshal(pullRes.Body.Bytes(), &pull); err != nil {
		t.Fatal(err)
	}
	foundStockChange := false
	for _, change := range pull.Changes {
		if change.EntityType == "master.catalog_item" && change.EntityID == itemID && change.Payload["stock_on_hand"] == float64(0) {
			foundStockChange = true
		}
	}
	if !foundStockChange {
		t.Fatalf("pull omitted updated item stock: %s", pullRes.Body.String())
	}
	err = store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		var salesCount, stock int64
		var receipt string
		var debits, credits int64
		if e := tx.QueryRow(ctx, `SELECT count(*) FROM sales WHERE tenant_id=$1`, tenantID).Scan(&salesCount); e != nil {
			return e
		}
		if e := tx.QueryRow(ctx, `SELECT stock_on_hand FROM catalog_items WHERE tenant_id=$1 AND id=$2`, tenantID, itemID).Scan(&stock); e != nil {
			return e
		}
		if e := tx.QueryRow(ctx, `SELECT receipt_number FROM sales WHERE tenant_id=$1 AND id=$2`, tenantID, saleID).Scan(&receipt); e != nil {
			return e
		}
		if e := tx.QueryRow(ctx, `SELECT COALESCE(sum(debit_minor),0),COALESCE(sum(credit_minor),0) FROM ledger_lines WHERE tenant_id=$1`, tenantID).Scan(&debits, &credits); e != nil {
			return e
		}
		if salesCount != 1 || stock != 0 || debits != credits || debits != 1250 || receipt != "T-ABCDEF1234-00000001" {
			return fmt.Errorf("sales=%d stock=%d debits=%d credits=%d receipt=%s", salesCount, stock, debits, credits, receipt)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func mustMarshal(value any) []byte { data, _ := json.Marshal(value); return data }
