package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Langa2007/POS/backend/internal/core/sales"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type itemInput struct {
	Name         string `json:"name"`
	SKU          string `json:"sku"`
	Barcode      string `json:"barcode"`
	CategoryID   string `json:"category_id"`
	PriceMinor   int64  `json:"price_minor"`
	CostMinor    int64  `json:"cost_minor"`
	StockOnHand  int64  `json:"stock_on_hand"`
	ReorderLevel int64  `json:"reorder_level"`
	Kind         string `json:"kind"`
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if value, parseErr := strconv.Atoi(raw); parseErr == nil && value > 0 && value <= 1000 {
			limit = value
		}
	}
	items := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT ci.id::text,ci.name,ci.sku,COALESCE(ci.barcode,''),ci.price_minor,ci.cost_minor,ci.currency,COALESCE(bi.quantity_on_hand,ci.stock_on_hand),ci.reorder_level,COALESCE(ci.category_id::text,''),ci.kind FROM catalog_items ci LEFT JOIN branch_inventory bi ON bi.tenant_id=ci.tenant_id AND bi.item_id=ci.id AND bi.branch_id=NULLIF($4,'')::uuid WHERE ci.tenant_id=$1 AND ci.active AND ($2='' OR lower(ci.name) LIKE lower('%'||$2||'%') OR ci.sku ILIKE '%'||$2||'%' OR ci.barcode=$2) ORDER BY ci.name LIMIT $3`, scope.TenantID, query, limit, scope.BranchID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, name, sku, barcode, currency, category, kind string
			var price, cost, stock, reorder int64
			if err := rows.Scan(&id, &name, &sku, &barcode, &price, &cost, &currency, &stock, &reorder, &category, &kind); err != nil {
				return err
			}
			items = append(items, map[string]any{"id": id, "name": name, "sku": sku, "barcode": barcode, "price_minor": price, "cost_minor": cost, "currency": currency, "stock_on_hand": stock, "reorder_level": reorder, "category_id": category, "kind": kind, "low_stock": stock <= reorder})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list catalog", "error", err)
		writeError(w, 500, "catalog.unavailable", "Catalog could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"items": items, "next_cursor": nil})
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var input itemInput
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if json.NewDecoder(r.Body).Decode(&input) != nil || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.SKU) == "" || input.PriceMinor < 0 || input.StockOnHand < 0 || input.CostMinor < 0 {
		writeError(w, 400, "catalog.invalid_item", "Name, SKU, non-negative price, cost, and stock are required")
		return
	}
	if input.Kind == "" {
		input.Kind = "stocked"
	}
	if input.Kind != "stocked" && input.Kind != "service" && input.Kind != "non_inventory" {
		writeError(w, 400, "catalog.invalid_item", "Unsupported item kind")
		return
	}
	itemID, _ := ids.NewUUIDv7()
	auditID, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO catalog_items(id,tenant_id,category_id,name,sku,barcode,kind,price_minor,cost_minor,stock_on_hand,reorder_level) VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,NULLIF($6,''),$7,$8,$9,$10,$11)`, itemID, scope.TenantID, input.CategoryID, strings.TrimSpace(input.Name), strings.TrimSpace(input.SKU), strings.TrimSpace(input.Barcode), input.Kind, input.PriceMinor, input.CostMinor, input.StockOnHand, input.ReorderLevel)
		if err != nil {
			return err
		}
		if err = publishCatalogItemChange(r, tx, scope.TenantID, itemID); err != nil {
			return err
		}
		branchID := scope.BranchID
		if branchID == "" {
			_ = tx.QueryRow(r.Context(), `SELECT id::text FROM branches WHERE tenant_id=$1 ORDER BY created_at,id LIMIT 1`, scope.TenantID).Scan(&branchID)
		}
		if input.Kind == "stocked" && branchID != "" {
			if _, err = tx.Exec(r.Context(), `INSERT INTO branch_inventory(tenant_id,branch_id,item_id,quantity_on_hand) VALUES($1,$2,$3,$4) ON CONFLICT(tenant_id,branch_id,item_id) DO UPDATE SET quantity_on_hand=excluded.quantity_on_hand,updated_at=now()`, scope.TenantID, branchID, itemID, input.StockOnHand); err != nil { return err }
		}
		if input.StockOnHand > 0 {
			movementID, _ := ids.NewUUIDv7()
			if _, err = tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason) VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,'opening_balance')`, movementID, scope.TenantID, branchID, itemID, input.StockOnHand); err != nil {
				return err
			}
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'catalog.item_created','catalog_item',$4,$5)`, auditID, scope.TenantID, scope.UserID, itemID, mustJSON(input))
		return err
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "catalog.duplicate", "An item with this SKU or barcode already exists")
			return
		}
		s.logger.Error("create catalog item", "error", err)
		writeError(w, 400, "catalog.create_failed", "Item could not be created. Check the product details and try again.")
		return
	}
	writeJSON(w, 201, map[string]any{"id": itemID, "name": input.Name, "sku": input.SKU, "price_minor": input.PriceMinor, "stock_on_hand": input.StockOnHand})
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil { writeError(w, 401, "auth.required", "A valid bearer token is required"); return }
	itemID := r.PathValue("itemID")
	if !uuidPattern.MatchString(itemID) { writeError(w, 400, "catalog.invalid_item", "A valid item ID is required"); return }
	var input itemInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&input) != nil || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.SKU) == "" || input.PriceMinor < 0 || input.CostMinor < 0 || input.StockOnHand < 0 || input.ReorderLevel < 0 {
		writeError(w, 400, "catalog.invalid_item", "Name, SKU, price, cost, stock, and reorder level must be valid")
		return
	}
	if input.Kind == "" { input.Kind = "stocked" }
	if input.Kind != "stocked" && input.Kind != "service" && input.Kind != "non_inventory" {
		writeError(w, 400, "catalog.invalid_item", "Unsupported item kind"); return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, execErr := tx.Exec(r.Context(), `UPDATE catalog_items SET name=$1,sku=$2,barcode=NULLIF($3,''),price_minor=$4,cost_minor=$5,stock_on_hand=$6,reorder_level=$7,kind=$8 WHERE tenant_id=$9 AND id=$10 AND active`, strings.TrimSpace(input.Name), strings.TrimSpace(input.SKU), strings.TrimSpace(input.Barcode), input.PriceMinor, input.CostMinor, input.StockOnHand, input.ReorderLevel, input.Kind, scope.TenantID, itemID)
		if execErr != nil { return execErr }
		if tag.RowsAffected() == 0 { return pgx.ErrNoRows }
		branchID := scope.BranchID
		if branchID == "" { _ = tx.QueryRow(r.Context(), `SELECT id::text FROM branches WHERE tenant_id=$1 ORDER BY created_at,id LIMIT 1`, scope.TenantID).Scan(&branchID) }
		if input.Kind == "stocked" && branchID != "" {
			_, execErr = tx.Exec(r.Context(), `INSERT INTO branch_inventory(tenant_id,branch_id,item_id,quantity_on_hand) VALUES($1,$2,$3,$4) ON CONFLICT(tenant_id,branch_id,item_id) DO UPDATE SET quantity_on_hand=excluded.quantity_on_hand,updated_at=now()`, scope.TenantID, branchID, itemID, input.StockOnHand)
			if execErr != nil { return execErr }
		}
		return publishCatalogItemChange(r, tx, scope.TenantID, itemID)
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { writeError(w, 409, "catalog.duplicate", "An item with this SKU or barcode already exists"); return }
		if errors.Is(err, pgx.ErrNoRows) { writeError(w, 404, "catalog.not_found", "Product was not found"); return }
		s.logger.Error("update catalog item", "error", err)
		writeError(w, 400, "catalog.update_failed", "Product could not be updated"); return
	}
	writeJSON(w, 200, map[string]any{"id": itemID, "name": input.Name, "sku": input.SKU, "price_minor": input.PriceMinor, "stock_on_hand": input.StockOnHand, "reorder_level": input.ReorderLevel})
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil { writeError(w, 401, "auth.required", "A valid bearer token is required"); return }
	itemID := r.PathValue("itemID")
	if !uuidPattern.MatchString(itemID) { writeError(w, 400, "catalog.invalid_item", "A valid item ID is required"); return }
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, execErr := tx.Exec(r.Context(), `UPDATE catalog_items SET active=false WHERE tenant_id=$1 AND id=$2 AND active`, scope.TenantID, itemID)
		if execErr != nil { return execErr }
		if tag.RowsAffected() == 0 { return pgx.ErrNoRows }
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) { writeError(w, 404, "catalog.not_found", "Product was not found"); return }
	if err != nil { s.logger.Error("delete catalog item", "error", err); writeError(w, 400, "catalog.delete_failed", "Product could not be deleted"); return }
	w.WriteHeader(http.StatusNoContent)
}

type saleInput struct {
	BranchID      string `json:"branch_id"`
	ShiftID       string `json:"shift_id"`
	CustomerName  string `json:"customer_name"`
	PaymentMethod string `json:"payment_method"`
	ReceivedMinor int64  `json:"received_minor"`
	Lines         []struct {
		ItemID        string `json:"item_id"`
		Quantity      int64  `json:"quantity"`
		DiscountMinor int64  `json:"discount_minor"`
	} `json:"lines"`
}

func (s *Server) createSale(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, 400, "request.idempotency_required", "Idempotency-Key is required")
		return
	}
	var input saleInput
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.Lines) == 0 {
		writeError(w, 400, "sales.invalid_sale", "At least one sale line is required")
		return
	}
	saleID, _ := ids.NewUUIDv7()
	paymentReference, err := newPaymentReference()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	saleStatus := "completed"
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var existing string
		if tx.QueryRow(r.Context(), `SELECT id::text,payment_reference,status FROM sales WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, key).Scan(&existing, &paymentReference, &saleStatus) == nil {
			saleID = existing
			return nil
		}
		branchID := input.BranchID
		if branchID == "" {
			branchID = scope.BranchID
		}
		if branchID == "" && tx.QueryRow(r.Context(), `SELECT id::text FROM branches WHERE tenant_id=$1 ORDER BY created_at LIMIT 1`, scope.TenantID).Scan(&branchID) != nil {
			return errors.New("branch is required")
		}
		if scope.BranchID != "" && scope.BranchID != branchID { return errors.New("branch is outside the authenticated scope") }
		shiftID := input.ShiftID
		if shiftID == "" && tx.QueryRow(r.Context(), `SELECT id::text FROM shifts WHERE tenant_id=$1 AND user_id=$2 AND status='open'`, scope.TenantID, scope.UserID).Scan(&shiftID) != nil {
			return errors.New("an open shift is required")
		}
		var total int64
		type line struct {
			id                          string
			qty, price, discount, total int64
		}
		lines := make([]line, 0, len(input.Lines))
		for _, raw := range input.Lines {
			var itemID, kind, currency string
			var price, stock int64
			if raw.Quantity <= 0 || tx.QueryRow(r.Context(), `SELECT id::text,price_minor,stock_on_hand,kind,currency FROM catalog_items WHERE tenant_id=$1 AND id=$2 AND active FOR UPDATE`, scope.TenantID, raw.ItemID).Scan(&itemID, &price, &stock, &kind, &currency) != nil {
				return errors.New("item not found")
			}
			if kind == "stocked" {
				if err := tx.QueryRow(r.Context(), `SELECT quantity_on_hand FROM branch_inventory WHERE tenant_id=$1 AND branch_id=$2 AND item_id=$3 FOR UPDATE`, scope.TenantID, branchID, itemID).Scan(&stock); err != nil { return errors.New("item is not stocked at this branch") }
				if stock < raw.Quantity { return errors.New("insufficient stock") }
			}
			lineTotal, calcErr := sales.Line{Quantity: raw.Quantity, UnitPriceMinor: price, DiscountMinor: raw.DiscountMinor}.TotalMinor()
			if calcErr != nil {
				return calcErr
			}
			total += lineTotal
			lines = append(lines, line{raw.ItemID, raw.Quantity, price, raw.DiscountMinor, lineTotal})
		}
		method := input.PaymentMethod
		if method == "" {
			method = "cash"
		}
		if method != "cash" && method != "mpesa_stk" {
			return errors.New("payment_method must be cash or mpesa_stk")
		}
		if (method == "cash" && input.ReceivedMinor < total) || (method == "mpesa_stk" && input.ReceivedMinor != 0) {
			return errors.New("cash sales require received_minor at least total; M-Pesa sales require received_minor zero")
		}
		status := "completed"
		if method == "mpesa_stk" {
			status = "pending_payment"
		}
		saleStatus = status
		if _, err := tx.Exec(r.Context(), `INSERT INTO sales(id,tenant_id,branch_id,cashier_id,shift_id,customer_name,total_minor,idempotency_key,status,payment_reference) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10)`, saleID, scope.TenantID, branchID, scope.UserID, shiftID, strings.TrimSpace(input.CustomerName), total, key, status, paymentReference); err != nil {
			return err
		}
		for _, item := range lines {
			lineID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO sale_lines(id,tenant_id,sale_id,item_id,quantity,unit_price_minor,discount_minor,line_total_minor) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, lineID, scope.TenantID, saleID, item.id, item.qty, item.price, item.discount, item.total); err != nil {
				return err
			}
			if _, err := tx.Exec(r.Context(), `UPDATE catalog_items SET stock_on_hand=stock_on_hand-$1 WHERE tenant_id=$2 AND id=$3 AND kind='stocked'`, item.qty, scope.TenantID, item.id); err != nil { return err }
			if _, err := tx.Exec(r.Context(), `UPDATE branch_inventory SET quantity_on_hand=quantity_on_hand-$1,updated_at=now() WHERE tenant_id=$2 AND branch_id=$3 AND item_id=$4`, item.qty, scope.TenantID, branchID, item.id); err != nil {
				return err
			}
			movementID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'sale',$6)`, movementID, scope.TenantID, branchID, item.id, -item.qty, saleID); err != nil {
				return err
			}
			if err := publishCatalogItemChange(r, tx, scope.TenantID, item.id); err != nil {
				return err
			}
		}
		if method == "cash" {
			tenderID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO cash_tenders(id,tenant_id,sale_id,amount_minor,received_minor,change_minor) VALUES($1,$2,$3,$4,$5,$6)`, tenderID, scope.TenantID, saleID, total, input.ReceivedMinor, input.ReceivedMinor-total); err != nil {
				return err
			}
			if _, err := tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,method,amount_minor,idempotency_key) VALUES($1,$2,$3,'cash',$4,$5)`, tenderID, scope.TenantID, saleID, total, key+"-cash"); err != nil {
				return err
			}
			journalID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO ledger_journals(id,tenant_id,reference_type,reference_id,description) VALUES($1,$2,'sale',$3,'Cash sale')`, journalID, scope.TenantID, saleID); err != nil {
				return err
			}
			cashAccount, revenueAccount, err := ensureLedgerAccounts(r, tx, scope.TenantID)
			if err != nil {
				return err
			}
			for _, posting := range []struct {
				account       string
				debit, credit int64
			}{{cashAccount, total, 0}, {revenueAccount, 0, total}} {
				lineID, _ := ids.NewUUIDv7()
				if _, err := tx.Exec(r.Context(), `INSERT INTO ledger_lines(id,tenant_id,journal_id,account_id,debit_minor,credit_minor) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, journalID, posting.account, posting.debit, posting.credit); err != nil {
					return err
				}
			}
		}
		auditID, _ := ids.NewUUIDv7()
		outboxID, _ := ids.NewUUIDv7()
		if _, err := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'sale.completed','sale',$4,jsonb_build_object('total_minor',$5))`, auditID, scope.TenantID, scope.UserID, saleID, total); err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO outbox(id,tenant_id,event_type,payload) VALUES($1,$2,'sale.completed.v1',jsonb_build_object('sale_id',$3,'total_minor',$4))`, outboxID, scope.TenantID, saleID, total)
		return err
	})
	if err != nil {
		writeError(w, 400, "sales.finalize_failed", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"sale_id": saleID, "payment_reference": paymentReference, "status": saleStatus})
}

func publishCatalogItemChange(r *http.Request, tx pgx.Tx, tenantID, itemID string) error {
	var id, name, sku, barcode, currency, kind string
	var price, stock, reorder int64
	if err := tx.QueryRow(r.Context(), `SELECT id::text,name,sku,COALESCE(barcode,''),currency,kind,price_minor,stock_on_hand,reorder_level FROM catalog_items WHERE tenant_id=$1 AND id=$2`, tenantID, itemID).Scan(&id, &name, &sku, &barcode, &currency, &kind, &price, &stock, &reorder); err != nil {
		return err
	}
	payload := map[string]any{"id": id, "name": name, "sku": sku, "barcode": barcode, "currency": strings.TrimSpace(currency), "kind": kind, "price_minor": price, "stock_on_hand": stock, "reorder_level": reorder}
	_, err := tx.Exec(r.Context(), `INSERT INTO sync_changes(tenant_id,entity_type,entity_id,operation,payload) VALUES($1,'master.catalog_item',$2,'update',$3)`, tenantID, itemID, mustJSON(payload))
	return err
}

func newPaymentReference() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}

func (s *Server) openShift(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct {
		BranchID          string `json:"branch_id"`
		OpeningFloatMinor int64  `json:"opening_float_minor"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || input.OpeningFloatMinor < 0 {
		writeError(w, 400, "shift.invalid", "Opening float must be non-negative")
		return
	}
	shiftID, _ := ids.NewUUIDv7()
	branchID := input.BranchID
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if branchID == "" {
			branchID = scope.BranchID
		}
		if branchID == "" {
			_ = tx.QueryRow(r.Context(), `SELECT id::text FROM branches WHERE tenant_id=$1 ORDER BY created_at LIMIT 1`, scope.TenantID).Scan(&branchID)
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO shifts(id,tenant_id,branch_id,user_id,opening_float_minor,status) VALUES($1,$2,$3,$4,$5,'open')`, shiftID, scope.TenantID, branchID, scope.UserID, input.OpeningFloatMinor)
		return err
	})
	if err != nil {
		writeError(w, 400, "shift.open_failed", "An open shift already exists or branch is invalid")
		return
	}
	writeJSON(w, 201, map[string]any{"shift_id": shiftID, "branch_id": branchID, "status": "open"})
}

func (s *Server) closeShift(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	shiftID := r.PathValue("shiftID")
	closing, parseErr := strconv.ParseInt(r.URL.Query().Get("closing_cash_minor"), 10, 64)
	if !uuidPattern.MatchString(shiftID) || parseErr != nil || closing < 0 {
		writeError(w, 400, "shift.invalid", "Provide a valid shift and closing cash")
		return
	}
	var variance int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var opening int64
		if err := tx.QueryRow(r.Context(), `SELECT opening_float_minor FROM shifts WHERE tenant_id=$1 AND id=$2 AND user_id=$3 AND status='open' FOR UPDATE`, scope.TenantID, shiftID, scope.UserID).Scan(&opening); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT $1 + COALESCE(SUM(ct.amount_minor),0) - $2 FROM shifts sh LEFT JOIN sales sa ON sa.tenant_id=sh.tenant_id AND sa.shift_id=sh.id AND sa.status='completed' LEFT JOIN cash_tenders ct ON ct.tenant_id=sa.tenant_id AND ct.sale_id=sa.id WHERE sh.tenant_id=$3 AND sh.id=$4 GROUP BY sh.id`, opening, closing, scope.TenantID, shiftID).Scan(&variance); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(), `UPDATE shifts SET closing_cash_minor=$1,variance_minor=$2,status='closed',closed_at=now() WHERE tenant_id=$3 AND id=$4`, closing, variance, scope.TenantID, shiftID)
		return err
	})
	if err != nil {
		writeError(w, 400, "shift.close_failed", "Shift could not be closed")
		return
	}
	writeJSON(w, 200, map[string]any{"shift_id": shiftID, "status": "closed", "variance_minor": variance})
}

func (s *Server) dailyReport(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var report struct {
		Sales int64 `json:"sales_count"`
		Total int64 `json:"total_minor"`
		Items int64 `json:"items_sold"`
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT COUNT(*),COALESCE(SUM(total_minor),0),COALESCE((SELECT SUM(quantity) FROM sale_lines sl JOIN sales sx ON sx.tenant_id=sl.tenant_id AND sx.id=sl.sale_id WHERE sx.tenant_id=$1 AND sx.status='completed' AND sx.created_at::date=current_date),0) FROM sales WHERE tenant_id=$1 AND status='completed' AND created_at::date=current_date`, scope.TenantID).Scan(&report.Sales, &report.Total, &report.Items)
	})
	if err != nil {
		writeError(w, 500, "reports.unavailable", "Report could not be loaded")
		return
	}
	writeJSON(w, 200, report)
}

func mustJSON(value any) []byte { body, _ := json.Marshal(value); return body }
