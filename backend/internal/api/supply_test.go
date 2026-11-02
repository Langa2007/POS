package api

import (
	"context"
	"encoding/json"
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

func TestPhase4TaxAndPurchaseReceivingFlow(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for Phase 4 workflow integration")
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
	tenant, business, branch, user, role, item := newTestID(t), newTestID(t), newTestID(t), newTestID(t), newTestID(t), newTestID(t)
	err = store.TenantTx(ctx, tenant, func(tx pgx.Tx) error {
		for _, q := range []struct {
			sql  string
			args []any
		}{{`INSERT INTO tenants(id,name) VALUES($1,'phase4')`, []any{tenant}}, {`INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'phase4')`, []any{business, tenant}}, {`INSERT INTO branches(id,tenant_id,business_id,name) VALUES($1,$2,$3,'main')`, []any{branch, tenant, business}}, {`INSERT INTO users(id,tenant_id,email,password_hash) VALUES($1,$2,$3,'unused')`, []any{user, tenant, strings.ReplaceAll(user, "-", " ") + "@phase4.test"}}, {`INSERT INTO roles(id,tenant_id,name) VALUES($1,$2,'owner')`, []any{role, tenant}}, {`INSERT INTO user_roles(tenant_id,user_id,role_id) VALUES($1,$2,$3)`, []any{tenant, user, role}}, {`INSERT INTO catalog_items(id,tenant_id,name,sku,price_minor,cost_minor,stock_on_hand) VALUES($1,$2,'Coffee','P4-COFFEE',500,100,0)`, []any{item, tenant}}} {
			if _, e := tx.Exec(ctx, q.sql, q.args...); e != nil {
				return e
			}
		}
		for _, permission := range []string{"tax.read", "tax.manage", "purchasing.read", "purchasing.manage", "purchasing.approve", "purchasing.receive"} {
			if _, e := tx.Exec(ctx, `INSERT INTO permissions(id,description) VALUES($1,'phase4 test') ON CONFLICT DO NOTHING`, permission); e != nil {
				return e
			}
			if _, e := tx.Exec(ctx, `INSERT INTO role_permissions(tenant_id,role_id,permission_id) VALUES($1,$2,$3)`, tenant, role, permission); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewTokenManager([]byte(strings.Repeat("p", 32)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := tokens.Issue(tenant, user)
	srv := New(store, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	do := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		return res
	}
	res := do(http.MethodPost, "/api/v1/tax/rates", `{"code":"VAT16","name":"VAT","basis_points":1600,"effective_from":"2026-01-01"}`)
	if res.Code != 201 {
		t.Fatalf("tax create %d %s", res.Code, res.Body.String())
	}
	res = do(http.MethodPost, "/api/v1/tax/calculate", `{"inclusion":"exclusive","rounding":"half_up","lines":[{"net_minor":1000,"rate_code":"VAT16"}]}`)
	if res.Code != 200 {
		t.Fatalf("tax calc %d %s", res.Code, res.Body.String())
	}
	var taxResult map[string]any
	if err = json.Unmarshal(res.Body.Bytes(), &taxResult); err != nil {
		t.Fatal(err)
	}
	if taxResult["tax_minor"] != float64(160) || taxResult["gross_minor"] != float64(1160) {
		t.Fatalf("tax result=%v", taxResult)
	}
	res = do(http.MethodPost, "/api/v1/suppliers", `{"name":"Acme","payment_terms_days":30}`)
	if res.Code != 201 {
		t.Fatalf("supplier create %d %s", res.Code, res.Body.String())
	}
	var supplier map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &supplier)
	poBody := `{"supplier_id":"` + supplier["id"].(string) + `","branch_id":"` + branch + `","order_number":"PO-P4-1","lines":[{"item_id":"` + item + `","quantity":5,"unit_cost_minor":100}]}`
	res = do(http.MethodPost, "/api/v1/purchase-orders", poBody)
	if res.Code != 201 {
		t.Fatalf("PO create %d %s", res.Code, res.Body.String())
	}
	var po map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &po)
	orderID := po["id"].(string)
	res = do(http.MethodPost, "/api/v1/purchase-orders/"+orderID+"/approve", `{}`)
	if res.Code != 200 {
		t.Fatalf("PO approve %d %s", res.Code, res.Body.String())
	}
	var poLines []struct {
		ID string `json:"id"`
	}
	if err = store.TenantTx(ctx, tenant, func(tx pgx.Tx) error {
		rows, e := tx.Query(ctx, `SELECT id::text FROM purchase_order_lines WHERE tenant_id=$1 AND purchase_order_id=$2`, tenant, orderID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var line struct {
				ID string `json:"id"`
			}
			if e = rows.Scan(&line.ID); e != nil {
				return e
			}
			poLines = append(poLines, line)
		}
		return rows.Err()
	}); err != nil {
		t.Fatal(err)
	}
	for idx, qty := range []int{2, 3} {
		receipt := `{"receipt_number":"GRN-P4-` + string(rune('1'+idx)) + `","lines":[{"purchase_order_line_id":"` + poLines[0].ID + `","quantity":` + string(rune('0'+qty)) + `,"unit_cost_minor":100}]}`
		res = do(http.MethodPost, "/api/v1/purchase-orders/"+orderID+"/receipts", receipt)
		if res.Code != 201 {
			t.Fatalf("receive part %d: %d %s", idx, res.Code, res.Body.String())
		}
	}
	var stock int64
	var status string
	if err = store.TenantTx(ctx, tenant, func(tx pgx.Tx) error {
		if e := tx.QueryRow(ctx, `SELECT stock_on_hand FROM catalog_items WHERE tenant_id=$1 AND id=$2`, tenant, item).Scan(&stock); e != nil {
			return e
		}
		return tx.QueryRow(ctx, `SELECT status FROM purchase_orders WHERE tenant_id=$1 AND id=$2`, tenant, orderID).Scan(&status)
	}); err != nil {
		t.Fatal(err)
	}
	if stock != 5 || status != "received" {
		t.Fatalf("stock=%d status=%s", stock, status)
	}
}

func newTestID(t *testing.T) string {
	t.Helper()
	id, err := ids.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
