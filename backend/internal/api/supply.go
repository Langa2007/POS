package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/core/fiscal"
	"github.com/Langa2007/POS/backend/internal/core/tax"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func supplyScope(r *http.Request) (tenancy.Scope, error) { return tenancy.Require(r.Context()) }

type taxRateInput struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	BasisPoints   int64  `json:"basis_points"`
	Exempt        bool   `json:"exempt"`
	CategoryID    string `json:"category_id"`
	EffectiveFrom string `json:"effective_from"`
	EffectiveTo   string `json:"effective_to"`
}

func (s *Server) taxRates(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	on := r.URL.Query().Get("on")
	if on == "" {
		on = time.Now().Format("2006-01-02")
	}
	if _, err = time.Parse("2006-01-02", on); err != nil {
		writeError(w, 400, "tax.invalid_date", "on must be YYYY-MM-DD")
		return
	}
	var rates []map[string]any
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT id::text,code,name,basis_points,exempt,COALESCE(category_id::text,''),effective_from,effective_to FROM tax_rates WHERE tenant_id=$1 AND effective_from<=$2::date AND (effective_to IS NULL OR effective_to >= $2::date) ORDER BY code,category_id NULLS FIRST`, scope.TenantID, on)
		if e != nil {
			return e
		}
		defer rows.Close()
		rates = make([]map[string]any, 0)
		for rows.Next() {
			var id, code, name, category string
			var bp int64
			var exempt bool
			var from time.Time
			var to *time.Time
			if e = rows.Scan(&id, &code, &name, &bp, &exempt, &category, &from, &to); e != nil {
				return e
			}
			rates = append(rates, map[string]any{"id": id, "code": code, "name": name, "basis_points": bp, "exempt": exempt, "category_id": category, "effective_from": from.Format("2006-01-02"), "effective_to": to})
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "tax.unavailable", "Tax rates could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"rates": rates, "effective_on": on})
}

func (s *Server) putTaxRate(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in taxRateInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&in) != nil || !validIdentifier(in.Code, 40) || strings.TrimSpace(in.Name) == "" || in.BasisPoints < 0 || in.BasisPoints > 10000 {
		writeError(w, 400, "tax.invalid_rate", "Provide a code, name and rate from 0 to 10000 basis points")
		return
	}
	from, e := time.Parse("2006-01-02", in.EffectiveFrom)
	if e != nil {
		writeError(w, 400, "tax.invalid_date", "effective_from must be YYYY-MM-DD")
		return
	}
	var to any
	if in.EffectiveTo != "" {
		parsed, e := time.Parse("2006-01-02", in.EffectiveTo)
		if e != nil || parsed.Before(from) {
			writeError(w, 400, "tax.invalid_date", "effective_to must be on or after effective_from")
			return
		}
		to = in.EffectiveTo
	}
	id, _ := ids.NewUUIDv7()
	auditID, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(r.Context(), `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, scope.TenantID+"/tax/"+in.Code); e != nil {
			return e
		}
		var overlap bool
		e := tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM tax_rates WHERE tenant_id=$1 AND code=$2 AND category_id IS NOT DISTINCT FROM NULLIF($3,'')::uuid AND effective_from<=COALESCE($5::date,'infinity'::date) AND COALESCE(effective_to,'infinity'::date)>=$4::date)`, scope.TenantID, in.Code, in.CategoryID, from, to).Scan(&overlap)
		if e != nil {
			return e
		}
		if overlap {
			return errors.New("tax rate dates overlap an existing rate")
		}
		if _, e = tx.Exec(r.Context(), `INSERT INTO tax_rates(id,tenant_id,code,name,basis_points,exempt,category_id,effective_from,effective_to) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,'')::uuid,$8,$9)`, id, scope.TenantID, in.Code, strings.TrimSpace(in.Name), in.BasisPoints, in.Exempt, in.CategoryID, from, to); e != nil {
			return e
		}
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'tax.rate_created','tax_rate',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		if strings.Contains(err.Error(), "overlap") {
			writeError(w, 409, "tax.rate_overlap", err.Error())
		} else {
			writeError(w, 400, "tax.rate_create_failed", "Tax rate could not be saved; verify its category and dates")
		}
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "code": in.Code, "effective_from": in.EffectiveFrom, "effective_to": in.EffectiveTo})
}

func (s *Server) calculateTax(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in struct {
		On               string     `json:"effective_on"`
		Inclusion        string     `json:"inclusion"`
		Rounding         string     `json:"rounding"`
		RoundPerDocument bool       `json:"round_per_document"`
		Lines            []tax.Line `json:"lines"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || len(in.Lines) == 0 || len(in.Lines) > 500 {
		writeError(w, 400, "tax.invalid_calculation", "Provide 1 to 500 tax lines")
		return
	}
	if in.On == "" {
		in.On = time.Now().Format("2006-01-02")
	}
	if _, err = time.Parse("2006-01-02", in.On); err != nil {
		writeError(w, 400, "tax.invalid_date", "effective_on must be YYYY-MM-DD")
		return
	}
	cfg := tax.Config{Rates: map[string]tax.Rate{}, RoundPerDocument: in.RoundPerDocument}
	if in.Inclusion == "inclusive" {
		cfg.Inclusion = tax.Inclusive
	} else if in.Inclusion != "" && in.Inclusion != "exclusive" {
		writeError(w, 400, "tax.invalid_calculation", "inclusion must be inclusive or exclusive")
		return
	}
	switch in.Rounding {
	case "", "half_up":
		cfg.Rounding = tax.HalfUp
	case "half_even":
		cfg.Rounding = tax.HalfEven
	case "down":
		cfg.Rounding = tax.Down
	default:
		writeError(w, 400, "tax.invalid_calculation", "rounding must be half_up, half_even, or down")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT DISTINCT ON(code) code,basis_points,exempt FROM tax_rates WHERE tenant_id=$1 AND category_id IS NULL AND effective_from<=$2::date AND (effective_to IS NULL OR effective_to >= $2::date) ORDER BY code,effective_from DESC`, scope.TenantID, in.On)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var code string
			var bp int64
			var ex bool
			if e = rows.Scan(&code, &bp, &ex); e != nil {
				return e
			}
			cfg.Rates[code] = tax.Rate{Code: code, BasisPoint: bp, Exempt: ex}
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "tax.unavailable", "Tax rates could not be loaded")
		return
	}
	result, err := cfg.Calculate(in.Lines)
	if err != nil {
		writeError(w, 400, "tax.calculation_failed", err.Error())
		return
	}
	writeJSON(w, 200, result)
}

func validIdentifier(v string, max int) bool {
	if len(v) == 0 || len(v) > max {
		return false
	}
	for _, c := range v {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

type supplierInput struct {
	Name         string `json:"name"`
	TaxNumber    string `json:"tax_number"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	PaymentTerms int    `json:"payment_terms_days"`
}

func (s *Server) suppliers(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	if r.Method == http.MethodPost {
		var in supplierInput
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&in) != nil || strings.TrimSpace(in.Name) == "" || in.PaymentTerms < 0 || in.PaymentTerms > 3650 {
			writeError(w, 400, "suppliers.invalid", "Supplier name and valid payment terms are required")
			return
		}
		id, _ := ids.NewUUIDv7()
		err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(r.Context(), `INSERT INTO suppliers(id,tenant_id,name,tax_number,phone,email,payment_terms_days) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7)`, id, scope.TenantID, strings.TrimSpace(in.Name), strings.TrimSpace(in.TaxNumber), strings.TrimSpace(in.Phone), strings.TrimSpace(in.Email), in.PaymentTerms)
			return e
		})
		if err != nil {
			writeError(w, 400, "suppliers.create_failed", "Supplier could not be saved")
			return
		}
		writeJSON(w, 201, map[string]any{"id": id, "name": in.Name, "status": "active"})
		return
	}
	items := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT id::text,name,COALESCE(tax_number,''),COALESCE(phone,''),COALESCE(email,''),payment_terms_days,status FROM suppliers WHERE tenant_id=$1 ORDER BY name`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var id, n, taxNo, phone, email, status string
			var terms int
			if e = rows.Scan(&id, &n, &taxNo, &phone, &email, &terms, &status); e != nil {
				return e
			}
			items = append(items, map[string]any{"id": id, "name": n, "tax_number": taxNo, "phone": phone, "email": email, "payment_terms_days": terms, "status": status})
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "suppliers.unavailable", "Suppliers could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"suppliers": items})
}

type purchaseLineInput struct {
	ItemID        string `json:"item_id"`
	Quantity      int64  `json:"quantity"`
	UnitCostMinor int64  `json:"unit_cost_minor"`
}
type purchaseOrderInput struct {
	SupplierID  string              `json:"supplier_id"`
	BranchID    string              `json:"branch_id"`
	OrderNumber string              `json:"order_number"`
	Currency    string              `json:"currency"`
	ExpectedAt  string              `json:"expected_at"`
	Lines       []purchaseLineInput `json:"lines"`
}

func (s *Server) purchaseOrders(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	if r.Method == http.MethodGet {
		result := make([]map[string]any, 0)
		err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
			rows, e := tx.Query(r.Context(), `SELECT id::text,supplier_id::text,branch_id::text,order_number,status,currency,expected_at,total_minor FROM purchase_orders WHERE tenant_id=$1 AND ($2='' OR branch_id=$2::uuid) ORDER BY created_at DESC LIMIT 500`, scope.TenantID, scope.BranchID)
			if e != nil {
				return e
			}
			defer rows.Close()
			for rows.Next() {
				var id, supplier, branch, num, status, currency string
				var expected *time.Time
				var total int64
				if e = rows.Scan(&id, &supplier, &branch, &num, &status, &currency, &expected, &total); e != nil {
					return e
				}
				result = append(result, map[string]any{"id": id, "supplier_id": supplier, "branch_id": branch, "order_number": num, "status": status, "currency": strings.TrimSpace(currency), "expected_at": expected, "total_minor": total})
			}
			return rows.Err()
		})
		if err != nil {
			writeError(w, 500, "purchasing.unavailable", "Purchase orders could not be loaded")
			return
		}
		writeJSON(w, 200, map[string]any{"purchase_orders": result})
		return
	}
	var in purchaseOrderInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || !uuidPattern.MatchString(in.SupplierID) || !uuidPattern.MatchString(in.BranchID) || !validIdentifier(in.OrderNumber, 40) || len(in.Lines) == 0 || len(in.Lines) > 500 {
		writeError(w, 400, "purchasing.invalid_order", "Provide supplier, branch, order number, and 1 to 500 lines")
		return
	}
	if scope.BranchID != "" && scope.BranchID != in.BranchID {
		writeError(w, 403, "auth.branch_denied", "Branch is outside the authenticated scope")
		return
	}
	if in.Currency == "" {
		in.Currency = "KES"
	}
	if len(in.Currency) != 3 {
		writeError(w, 400, "purchasing.invalid_currency", "Currency must be a 3-letter code")
		return
	}
	var total int64
	for _, line := range in.Lines {
		if !uuidPattern.MatchString(line.ItemID) || line.Quantity <= 0 || line.UnitCostMinor < 0 || line.Quantity > 0 && line.UnitCostMinor > 0 && line.Quantity > int64(^uint64(0)>>1)/line.UnitCostMinor {
			writeError(w, 400, "purchasing.invalid_line", "Purchase line is invalid or overflows")
			return
		}
		v := line.Quantity * line.UnitCostMinor
		if total > int64(^uint64(0)>>1)-v {
			writeError(w, 400, "purchasing.total_overflow", "Purchase order total is too large")
			return
		}
		total += v
	}
	id, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var expected any
		if in.ExpectedAt != "" {
			d, e := time.Parse("2006-01-02", in.ExpectedAt)
			if e != nil {
				return e
			}
			expected = d
		}
		if _, e := tx.Exec(r.Context(), `INSERT INTO purchase_orders(id,tenant_id,supplier_id,branch_id,order_number,currency,expected_at,total_minor,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, scope.TenantID, in.SupplierID, in.BranchID, in.OrderNumber, in.Currency, expected, total, scope.UserID); e != nil {
			return e
		}
		for _, line := range in.Lines {
			lineID, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO purchase_order_lines(id,tenant_id,purchase_order_id,item_id,ordered_quantity,unit_cost_minor) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, id, line.ItemID, line.Quantity, line.UnitCostMinor); e != nil {
				return e
			}
		}
		auditID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'purchasing.po_created','purchase_order',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 400, "purchasing.create_failed", "Purchase order could not be created; verify supplier, branch and items")
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "order_number": in.OrderNumber, "status": "draft", "total_minor": total})
}

func (s *Server) approvePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("orderID")
	if !uuidPattern.MatchString(id) {
		writeError(w, 400, "request.invalid_id", "Invalid order id")
		return
	}
	var status string
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, e := tx.Exec(r.Context(), `UPDATE purchase_orders SET status='approved' WHERE tenant_id=$1 AND id=$2 AND status='draft' AND ($3='' OR branch_id=$3::uuid)`, scope.TenantID, id, scope.BranchID)
		if e != nil {
			return e
		}
		if tag.RowsAffected() != 1 {
			return errors.New("order is unavailable or not a draft")
		}
		status = "approved"
		auditID, _ := ids.NewUUIDv7()
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id) VALUES($1,$2,$3,'purchasing.po_approved','purchase_order',$4)`, auditID, scope.TenantID, scope.UserID, id)
		return e
	})
	if err != nil {
		writeError(w, 409, "purchasing.approval_conflict", "Purchase order is missing or cannot be approved")
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "status": status})
}

type receiveLineInput struct {
	PurchaseOrderLineID string `json:"purchase_order_line_id"`
	Quantity            int64  `json:"quantity"`
	UnitCostMinor       int64  `json:"unit_cost_minor"`
}

func (s *Server) receivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	orderID := r.PathValue("orderID")
	var in struct {
		ReceiptNumber string             `json:"receipt_number"`
		Lines         []receiveLineInput `json:"lines"`
	}
	if !uuidPattern.MatchString(orderID) || json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || !validIdentifier(in.ReceiptNumber, 40) || len(in.Lines) == 0 {
		writeError(w, 400, "receiving.invalid", "Provide a receipt number and received lines")
		return
	}
	receiptID, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var branch, status string
		e := tx.QueryRow(r.Context(), `SELECT branch_id::text,status FROM purchase_orders WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, scope.TenantID, orderID).Scan(&branch, &status)
		if e != nil {
			return e
		}
		if scope.BranchID != "" && branch != scope.BranchID {
			return errors.New("branch denied")
		}
		if status != "approved" && status != "partially_received" {
			return errors.New("order is not receivable")
		}
		if _, e = tx.Exec(r.Context(), `INSERT INTO goods_receipts(id,tenant_id,purchase_order_id,branch_id,receipt_number,received_by) VALUES($1,$2,$3,$4,$5,$6)`, receiptID, scope.TenantID, orderID, branch, in.ReceiptNumber, scope.UserID); e != nil {
			return e
		}
		seen := map[string]bool{}
		for _, line := range in.Lines {
			if !uuidPattern.MatchString(line.PurchaseOrderLineID) || line.Quantity <= 0 || line.UnitCostMinor < 0 || seen[line.PurchaseOrderLineID] {
				return errors.New("invalid or duplicate receipt line")
			}
			seen[line.PurchaseOrderLineID] = true
			var item string
			var ordered, received, poCost int64
			if e = tx.QueryRow(r.Context(), `SELECT item_id::text,ordered_quantity,received_quantity,unit_cost_minor FROM purchase_order_lines WHERE tenant_id=$1 AND id=$2 AND purchase_order_id=$3 FOR UPDATE`, scope.TenantID, line.PurchaseOrderLineID, orderID).Scan(&item, &ordered, &received, &poCost); e != nil {
				return e
			}
			if line.Quantity > ordered-received {
				return errors.New("receipt exceeds outstanding ordered quantity")
			}
			cost := line.UnitCostMinor
			if cost == 0 {
				cost = poCost
			}
			if _, e = tx.Exec(r.Context(), `UPDATE purchase_order_lines SET received_quantity=received_quantity+$1 WHERE tenant_id=$2 AND id=$3`, line.Quantity, scope.TenantID, line.PurchaseOrderLineID); e != nil {
				return e
			}
			lineID, _ := ids.NewUUIDv7()
			if _, e = tx.Exec(r.Context(), `INSERT INTO goods_receipt_lines(id,tenant_id,goods_receipt_id,purchase_order_line_id,item_id,quantity,unit_cost_minor) VALUES($1,$2,$3,$4,$5,$6,$7)`, lineID, scope.TenantID, receiptID, line.PurchaseOrderLineID, item, line.Quantity, cost); e != nil {
				return e
			}
			if _, e = tx.Exec(r.Context(), `UPDATE catalog_items SET stock_on_hand=stock_on_hand+$1,cost_minor=$2 WHERE tenant_id=$3 AND id=$4`, line.Quantity, cost, scope.TenantID, item); e != nil {
				return e
			}
			if _, e = tx.Exec(r.Context(), `INSERT INTO branch_inventory(tenant_id,branch_id,item_id,quantity_on_hand) VALUES($1,$2,$3,$4) ON CONFLICT(tenant_id,branch_id,item_id) DO UPDATE SET quantity_on_hand=branch_inventory.quantity_on_hand+excluded.quantity_on_hand,updated_at=now()`, scope.TenantID, branch, item, line.Quantity); e != nil { return e }
			movement, _ := ids.NewUUIDv7()
			if _, e = tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'purchase_receipt',$6)`, movement, scope.TenantID, branch, item, line.Quantity, receiptID); e != nil {
				return e
			}
			layer, _ := ids.NewUUIDv7()
			if _, e = tx.Exec(r.Context(), `INSERT INTO inventory_cost_layers(id,tenant_id,item_id,branch_id,source_type,source_id,quantity_remaining,unit_cost_minor) VALUES($1,$2,$3,$4,'goods_receipt',$5,$6,$7)`, layer, scope.TenantID, item, branch, receiptID, line.Quantity, cost); e != nil {
				return e
			}
		}
		var left bool
		if e = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM purchase_order_lines WHERE tenant_id=$1 AND purchase_order_id=$2 AND received_quantity<ordered_quantity)`, scope.TenantID, orderID).Scan(&left); e != nil {
			return e
		}
		next := "received"
		if left {
			next = "partially_received"
		}
		if _, e = tx.Exec(r.Context(), `UPDATE purchase_orders SET status=$1 WHERE tenant_id=$2 AND id=$3`, next, scope.TenantID, orderID); e != nil {
			return e
		}
		auditID, _ := ids.NewUUIDv7()
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'purchasing.receipt_posted','goods_receipt',$4,jsonb_build_object('purchase_order_id',$5::text))`, auditID, scope.TenantID, scope.UserID, receiptID, orderID)
		return e
	})
	if err != nil {
		writeError(w, 409, "receiving.post_failed", "Receipt could not be posted; check order status, branch, and outstanding quantities")
		return
	}
	writeJSON(w, 201, map[string]any{"id": receiptID, "purchase_order_id": orderID, "status": "posted"})
}

func (s *Server) reorderSuggestions(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var result []map[string]any
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT id::text,name,sku,stock_on_hand,reorder_level,GREATEST(reorder_level-stock_on_hand,0) suggested_quantity FROM catalog_items WHERE tenant_id=$1 AND active AND kind='stocked' AND stock_on_hand<=reorder_level ORDER BY name`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		result = make([]map[string]any, 0)
		for rows.Next() {
			var id, name, sku string
			var stock, level, suggest int64
			if e = rows.Scan(&id, &name, &sku, &stock, &level, &suggest); e != nil {
				return e
			}
			result = append(result, map[string]any{"item_id": id, "name": name, "sku": sku, "stock_on_hand": stock, "reorder_level": level, "suggested_quantity": suggest})
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "purchasing.reorder_unavailable", "Reorder suggestions could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"suggestions": result})
}

func (s *Server) supplySummary(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var stockValue, openPO, payables int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if e := tx.QueryRow(r.Context(), `SELECT COALESCE(sum(stock_on_hand*cost_minor),0) FROM catalog_items WHERE tenant_id=$1 AND active AND kind='stocked'`, scope.TenantID).Scan(&stockValue); e != nil {
			return e
		}
		if e := tx.QueryRow(r.Context(), `SELECT COALESCE(sum(total_minor),0) FROM purchase_orders WHERE tenant_id=$1 AND status IN ('approved','partially_received')`, scope.TenantID).Scan(&openPO); e != nil {
			return e
		}
		return tx.QueryRow(r.Context(), `SELECT COALESCE(sum(i.total_minor-COALESCE(p.paid,0)),0) FROM supplier_invoices i LEFT JOIN (SELECT tenant_id,invoice_id,sum(amount_minor) paid FROM supplier_payments GROUP BY tenant_id,invoice_id) p ON p.tenant_id=i.tenant_id AND p.invoice_id=i.id WHERE i.tenant_id=$1 AND i.status IN ('open','partially_paid')`, scope.TenantID).Scan(&payables)
	})
	if err != nil {
		writeError(w, 500, "purchasing.report_unavailable", "Supply summary could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"inventory_value_minor": stockValue, "open_purchase_orders_minor": openPO, "payables_minor": payables})
}

type stockTakeLineInput struct {
	ItemID  string `json:"item_id"`
	Counted int64  `json:"counted_quantity"`
}

func (s *Server) createStockTake(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in struct {
		BranchID string               `json:"branch_id"`
		Blind    bool                 `json:"blind"`
		Lines    []stockTakeLineInput `json:"lines"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || !uuidPattern.MatchString(in.BranchID) || len(in.Lines) == 0 || len(in.Lines) > 2000 {
		writeError(w, 400, "inventory.invalid_count", "Provide a branch and 1 to 2000 item counts")
		return
	}
	if scope.BranchID != "" && scope.BranchID != in.BranchID {
		writeError(w, 403, "auth.branch_denied", "Branch is outside the authenticated scope")
		return
	}
	id, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(r.Context(), `INSERT INTO stock_takes(id,tenant_id,branch_id,status,blind) VALUES($1,$2,$3,'pending_approval',$4)`, id, scope.TenantID, in.BranchID, in.Blind); e != nil {
			return e
		}
		seen := map[string]bool{}
		for _, line := range in.Lines {
			if !uuidPattern.MatchString(line.ItemID) || line.Counted < 0 || seen[line.ItemID] {
				return errors.New("invalid or duplicate count")
			}
			seen[line.ItemID] = true
			var expected int64
			if e := tx.QueryRow(r.Context(), `SELECT quantity_on_hand FROM branch_inventory WHERE tenant_id=$1 AND branch_id=$2 AND item_id=$3 FOR UPDATE`, scope.TenantID, in.BranchID, line.ItemID).Scan(&expected); e != nil {
				return e
			}
			lineID, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO stock_take_lines(id,tenant_id,stock_take_id,item_id,expected_quantity,counted_quantity) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, id, line.ItemID, expected, line.Counted); e != nil {
				return e
			}
		}
		auditID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'inventory.stock_take_started','stock_take',$4,jsonb_build_object('blind',$5,'line_count',$6))`, auditID, scope.TenantID, scope.UserID, id, in.Blind, len(in.Lines))
		return e
	})
	if err != nil {
		writeError(w, 400, "inventory.count_failed", "Stock count could not be saved; verify branch and item IDs")
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "pending_approval", "blind": in.Blind})
}

func (s *Server) approveStockTake(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("takeID")
	if !uuidPattern.MatchString(id) {
		writeError(w, 400, "request.invalid_id", "Invalid stock take id")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var branch string
		if e := tx.QueryRow(r.Context(), `SELECT branch_id::text FROM stock_takes WHERE tenant_id=$1 AND id=$2 AND status='pending_approval' FOR UPDATE`, scope.TenantID, id).Scan(&branch); e != nil {
			return e
		}
		if scope.BranchID != "" && scope.BranchID != branch {
			return errors.New("branch denied")
		}
		rows, e := tx.Query(r.Context(), `SELECT id::text,item_id::text,expected_quantity,counted_quantity FROM stock_take_lines WHERE tenant_id=$1 AND stock_take_id=$2 ORDER BY item_id FOR UPDATE`, scope.TenantID, id)
		if e != nil {
			return e
		}
		type adjust struct {
			line, item        string
			expected, counted int64
		}
		var lines []adjust
		for rows.Next() {
			var v adjust
			if e = rows.Scan(&v.line, &v.item, &v.expected, &v.counted); e != nil {
				rows.Close()
				return e
			}
			lines = append(lines, v)
		}
		if e = rows.Err(); e != nil {
			rows.Close()
			return e
		}
		rows.Close()
		for _, v := range lines {
			delta := v.counted - v.expected
			if delta == 0 {
				continue
			}
			if _, e = tx.Exec(r.Context(), `UPDATE catalog_items SET stock_on_hand=stock_on_hand+$1 WHERE tenant_id=$2 AND id=$3`, delta, scope.TenantID, v.item); e != nil {
				return e
			}
			if _, e = tx.Exec(r.Context(), `UPDATE branch_inventory SET quantity_on_hand=quantity_on_hand+$1,updated_at=now() WHERE tenant_id=$2 AND branch_id=$3 AND item_id=$4`, delta, scope.TenantID, branch, v.item); e != nil { return e }
			movement, _ := ids.NewUUIDv7()
			if _, e = tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'stock_take_adjustment',$6)`, movement, scope.TenantID, branch, v.item, delta, id); e != nil {
				return e
			}
		}
		_, e = tx.Exec(r.Context(), `UPDATE stock_takes SET status='posted',approved_by=$1,approved_at=now() WHERE tenant_id=$2 AND id=$3`, scope.UserID, scope.TenantID, id)
		if e != nil {
			return e
		}
		auditID, _ := ids.NewUUIDv7()
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id) VALUES($1,$2,$3,'inventory.stock_take_posted','stock_take',$4)`, auditID, scope.TenantID, scope.UserID, id)
		return e
	})
	if err != nil {
		writeError(w, 409, "inventory.approval_conflict", "Stock take is missing, outside branch scope, or already posted")
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "status": "posted"})
}

type transferLineInput struct {
	ItemID   string `json:"item_id"`
	Quantity int64  `json:"quantity"`
}

func (s *Server) stockTransfers(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	if r.Method == http.MethodGet {
		result := make([]map[string]any, 0)
		err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
			rows, e := tx.Query(r.Context(), `SELECT id::text,from_branch_id::text,to_branch_id::text,status,shipped_at,received_at FROM stock_transfers WHERE tenant_id=$1 AND ($2='' OR from_branch_id=$2::uuid OR to_branch_id=$2::uuid) ORDER BY COALESCE(shipped_at,now()) DESC LIMIT 500`, scope.TenantID, scope.BranchID)
			if e != nil {
				return e
			}
			defer rows.Close()
			for rows.Next() {
				var id, from, to, status string
				var shipped, received *time.Time
				if e = rows.Scan(&id, &from, &to, &status, &shipped, &received); e != nil {
					return e
				}
				result = append(result, map[string]any{"id": id, "from_branch_id": from, "to_branch_id": to, "status": status, "shipped_at": shipped, "received_at": received})
			}
			return rows.Err()
		})
		if err != nil {
			writeError(w, 500, "inventory.transfers_unavailable", "Transfers could not be loaded")
			return
		}
		writeJSON(w, 200, map[string]any{"transfers": result})
		return
	}
	var in struct {
		FromBranchID string              `json:"from_branch_id"`
		ToBranchID   string              `json:"to_branch_id"`
		Lines        []transferLineInput `json:"lines"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || !uuidPattern.MatchString(in.FromBranchID) || !uuidPattern.MatchString(in.ToBranchID) || in.FromBranchID == in.ToBranchID || len(in.Lines) == 0 {
		writeError(w, 400, "inventory.invalid_transfer", "Provide different source/destination branches and transfer lines")
		return
	}
	if scope.BranchID != "" && scope.BranchID != in.FromBranchID {
		writeError(w, 403, "auth.branch_denied", "Source branch is outside the authenticated scope")
		return
	}
	id, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(r.Context(), `INSERT INTO stock_transfers(id,tenant_id,from_branch_id,to_branch_id,status,shipped_at,created_by) VALUES($1,$2,$3,$4,'in_transit',now(),$5)`, id, scope.TenantID, in.FromBranchID, in.ToBranchID, scope.UserID); e != nil {
			return e
		}
		seen := map[string]bool{}
		for _, line := range in.Lines {
			if !uuidPattern.MatchString(line.ItemID) || line.Quantity <= 0 || seen[line.ItemID] {
				return errors.New("invalid or repeated transfer item")
			}
			seen[line.ItemID] = true
			var stock, cost int64
			if e := tx.QueryRow(r.Context(), `SELECT bi.quantity_on_hand,ci.cost_minor FROM branch_inventory bi JOIN catalog_items ci ON ci.tenant_id=bi.tenant_id AND ci.id=bi.item_id WHERE bi.tenant_id=$1 AND bi.branch_id=$2 AND bi.item_id=$3 AND ci.active AND ci.kind='stocked' FOR UPDATE OF bi,ci`, scope.TenantID, in.FromBranchID, line.ItemID).Scan(&stock, &cost); e != nil {
				return e
			}
			if stock < line.Quantity {
				return errors.New("insufficient source stock")
			}
			if _, e := tx.Exec(r.Context(), `UPDATE branch_inventory SET quantity_on_hand=quantity_on_hand-$1,updated_at=now() WHERE tenant_id=$2 AND branch_id=$3 AND item_id=$4 AND quantity_on_hand >= $1`, line.Quantity, scope.TenantID, in.FromBranchID, line.ItemID); e != nil {
				return e
			}
			lineID, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO stock_transfer_lines(id,tenant_id,transfer_id,item_id,shipped_quantity) VALUES($1,$2,$3,$4,$5)`, lineID, scope.TenantID, id, line.ItemID, line.Quantity); e != nil {
				return e
			}
			movement, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'transfer_shipped',$6)`, movement, scope.TenantID, in.FromBranchID, line.ItemID, -line.Quantity, id); e != nil {
				return e
			}
		}
		auditID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'inventory.transfer_shipped','stock_transfer',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 409, "inventory.transfer_failed", "Transfer could not be shipped; verify stock and branch IDs")
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "in_transit"})
}

func (s *Server) receiveStockTransfer(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("transferID")
	var in struct {
		Lines []struct {
			ItemID   string `json:"item_id"`
			Received int64  `json:"received_quantity"`
		} `json:"lines"`
	}
	if !uuidPattern.MatchString(id) || json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil {
		writeError(w, 400, "inventory.invalid_transfer_receipt", "Transfer id and received quantities are required")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var dest, status string
		if e := tx.QueryRow(r.Context(), `SELECT to_branch_id::text,status FROM stock_transfers WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, scope.TenantID, id).Scan(&dest, &status); e != nil {
			return e
		}
		if status != "in_transit" || scope.BranchID != "" && scope.BranchID != dest {
			return errors.New("transfer unavailable or branch denied")
		}
		rows, e := tx.Query(r.Context(), `SELECT item_id::text,shipped_quantity FROM stock_transfer_lines WHERE tenant_id=$1 AND transfer_id=$2 ORDER BY item_id FOR UPDATE`, scope.TenantID, id)
		if e != nil {
			return e
		}
		expected := map[string]int64{}
		for rows.Next() {
			var item string
			var qty int64
			if e = rows.Scan(&item, &qty); e != nil {
				rows.Close()
				return e
			}
			expected[item] = qty
		}
		if e = rows.Err(); e != nil {
			rows.Close()
			return e
		}
		rows.Close()
		if len(in.Lines) != len(expected) {
			return errors.New("all transfer lines must be reconciled")
		}
		seen := map[string]bool{}
		for _, line := range in.Lines {
			qty, ok := expected[line.ItemID]
			if !ok || seen[line.ItemID] || line.Received < 0 || line.Received > qty {
				return errors.New("received quantity invalid")
			}
			seen[line.ItemID] = true
			if _, e = tx.Exec(r.Context(), `UPDATE stock_transfer_lines SET received_quantity=$1 WHERE tenant_id=$2 AND transfer_id=$3 AND item_id=$4`, line.Received, scope.TenantID, id, line.ItemID); e != nil {
				return e
			}
			if line.Received > 0 {
				if _, e = tx.Exec(r.Context(), `INSERT INTO branch_inventory(tenant_id,branch_id,item_id,quantity_on_hand) VALUES($1,$2,$3,$4) ON CONFLICT(tenant_id,branch_id,item_id) DO UPDATE SET quantity_on_hand=branch_inventory.quantity_on_hand+excluded.quantity_on_hand,updated_at=now()`, scope.TenantID, dest, line.ItemID, line.Received); e != nil {
					return e
				}
				movement, _ := ids.NewUUIDv7()
				if _, e = tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'transfer_received',$6)`, movement, scope.TenantID, dest, line.ItemID, line.Received, id); e != nil {
					return e
				}
			}
		}
		if _, e = tx.Exec(r.Context(), `UPDATE stock_transfers SET status='received',received_at=now() WHERE tenant_id=$1 AND id=$2`, scope.TenantID, id); e != nil {
			return e
		}
		auditID, _ := ids.NewUUIDv7()
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'inventory.transfer_received','stock_transfer',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 409, "inventory.transfer_receive_failed", "Transfer receipt could not be posted; verify destination and quantities")
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "status": "received"})
}

type supplierInvoiceInput struct {
	SupplierID    string `json:"supplier_id"`
	ReceiptID     string `json:"goods_receipt_id"`
	InvoiceNumber string `json:"invoice_number"`
	Currency      string `json:"currency"`
	Total         int64  `json:"total_minor"`
	DueAt         string `json:"due_at"`
}

func (s *Server) createSupplierInvoice(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in supplierInvoiceInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&in) != nil || !uuidPattern.MatchString(in.SupplierID) || !validIdentifier(in.InvoiceNumber, 60) || in.Total <= 0 {
		writeError(w, 400, "payables.invalid_invoice", "Provide supplier, invoice number and positive total")
		return
	}
	if in.Currency == "" {
		in.Currency = "KES"
	}
	if len(in.Currency) != 3 {
		writeError(w, 400, "payables.invalid_currency", "Currency must be 3 letters")
		return
	}
	id, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var due any
		if in.DueAt != "" {
			d, e := time.Parse("2006-01-02", in.DueAt)
			if e != nil {
				return e
			}
			due = d
		}
		if _, e := tx.Exec(r.Context(), `INSERT INTO supplier_invoices(id,tenant_id,supplier_id,goods_receipt_id,invoice_number,currency,total_minor,due_at) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8)`, id, scope.TenantID, in.SupplierID, in.ReceiptID, in.InvoiceNumber, in.Currency, in.Total, due); e != nil {
			return e
		}
		auditID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'payables.invoice_created','supplier_invoice',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 400, "payables.invoice_create_failed", "Supplier invoice could not be saved; verify supplier and receipt")
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "open", "total_minor": in.Total})
}

func (s *Server) supplierInvoices(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	items := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT i.id::text,i.supplier_id::text,i.invoice_number,i.currency,i.total_minor,COALESCE(sum(p.amount_minor),0),i.due_at,i.status FROM supplier_invoices i LEFT JOIN supplier_payments p ON p.tenant_id=i.tenant_id AND p.invoice_id=i.id WHERE i.tenant_id=$1 GROUP BY i.id ORDER BY i.due_at NULLS LAST,i.created_at`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var id, supplier, num, currency, status string
			var total, paid int64
			var due *time.Time
			if e = rows.Scan(&id, &supplier, &num, &currency, &total, &paid, &due, &status); e != nil {
				return e
			}
			items = append(items, map[string]any{"id": id, "supplier_id": supplier, "invoice_number": num, "currency": strings.TrimSpace(currency), "total_minor": total, "paid_minor": paid, "balance_minor": total - paid, "due_at": due, "status": status})
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "payables.unavailable", "Supplier invoices could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"invoices": items})
}

func (s *Server) paySupplierInvoice(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("invoiceID")
	var in struct {
		IdempotencyKey string `json:"idempotency_key"`
		Amount         int64  `json:"amount_minor"`
		Method         string `json:"method"`
		Reference      string `json:"reference"`
	}
	if !uuidPattern.MatchString(id) || json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&in) != nil || len(strings.TrimSpace(in.IdempotencyKey)) < 8 || in.Amount <= 0 || !validIdentifier(in.Method, 24) {
		writeError(w, 400, "payables.invalid_payment", "Idempotency key, payment method and positive amount are required")
		return
	}
	paymentID, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var invoice string
		var priorAmount int64
		e := tx.QueryRow(r.Context(), `SELECT id::text,amount_minor FROM supplier_payments WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, in.IdempotencyKey).Scan(&invoice, &priorAmount)
		if e == nil {
			if invoice != id || priorAmount != in.Amount {
				return errors.New("idempotency conflict")
			}
			return nil
		}
		if e != pgx.ErrNoRows {
			return e
		}
		var total, paid int64
		if e = tx.QueryRow(r.Context(), `SELECT total_minor FROM supplier_invoices WHERE tenant_id=$1 AND id=$2 AND status IN ('open','partially_paid') FOR UPDATE`, scope.TenantID, id).Scan(&total); e != nil {
			return e
		}
		if e = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(amount_minor),0) FROM supplier_payments WHERE tenant_id=$1 AND invoice_id=$2`, scope.TenantID, id).Scan(&paid); e != nil {
			return e
		}
		if in.Amount > total-paid {
			return errors.New("payment exceeds invoice balance")
		}
		if _, e = tx.Exec(r.Context(), `INSERT INTO supplier_payments(id,tenant_id,invoice_id,idempotency_key,amount_minor,method,reference,created_by) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8)`, paymentID, scope.TenantID, id, in.IdempotencyKey, in.Amount, in.Method, in.Reference, scope.UserID); e != nil {
			return e
		}
		status := "partially_paid"
		if paid+in.Amount == total {
			status = "paid"
		}
		if _, e = tx.Exec(r.Context(), `UPDATE supplier_invoices SET status=$1 WHERE tenant_id=$2 AND id=$3`, status, scope.TenantID, id); e != nil {
			return e
		}
		auditID, _ := ids.NewUUIDv7()
		_, e = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'payables.payment_posted','supplier_invoice',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 409, "payables.payment_failed", "Payment could not be posted; check invoice balance or idempotency key")
		return
	}
	writeJSON(w, 201, map[string]any{"invoice_id": id, "status": "posted"})
}

type supplierReturnLine struct {
	ItemID   string `json:"item_id"`
	Quantity int64  `json:"quantity"`
	UnitCost int64  `json:"unit_cost_minor"`
}

func (s *Server) createSupplierReturn(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in struct {
		SupplierID   string               `json:"supplier_id"`
		BranchID     string               `json:"branch_id"`
		ReturnNumber string               `json:"return_number"`
		Lines        []supplierReturnLine `json:"lines"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in) != nil || !uuidPattern.MatchString(in.SupplierID) || !uuidPattern.MatchString(in.BranchID) || !validIdentifier(in.ReturnNumber, 50) || len(in.Lines) == 0 {
		writeError(w, 400, "purchasing.invalid_return", "Return supplier, branch, number and lines are required")
		return
	}
	if scope.BranchID != "" && scope.BranchID != in.BranchID {
		writeError(w, 403, "auth.branch_denied", "Branch is outside the authenticated scope")
		return
	}
	id, _ := ids.NewUUIDv7()
	var total int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		type row struct {
			item      string
			qty, cost int64
		}
		var checked []row
		seen := map[string]bool{}
		for _, line := range in.Lines {
			if !uuidPattern.MatchString(line.ItemID) || line.Quantity <= 0 || line.UnitCost < 0 || seen[line.ItemID] {
				return errors.New("invalid return line")
			}
			seen[line.ItemID] = true
			var stock int64
			if e := tx.QueryRow(r.Context(), `SELECT quantity_on_hand FROM branch_inventory WHERE tenant_id=$1 AND branch_id=$2 AND item_id=$3 FOR UPDATE`, scope.TenantID, in.BranchID, line.ItemID).Scan(&stock); e != nil {
				return e
			}
			if stock < line.Quantity {
				return errors.New("return exceeds stock on hand")
			}
			if line.Quantity > 0 && line.UnitCost > 0 && line.Quantity > int64(^uint64(0)>>1)/line.UnitCost {
				return errors.New("return total overflow")
			}
			v := line.Quantity * line.UnitCost
			if total > int64(^uint64(0)>>1)-v {
				return errors.New("return total overflow")
			}
			total += v
			checked = append(checked, row{line.ItemID, line.Quantity, line.UnitCost})
		}
		if _, e := tx.Exec(r.Context(), `INSERT INTO supplier_returns(id,tenant_id,supplier_id,branch_id,return_number,total_minor,created_by) VALUES($1,$2,$3,$4,$5,$6,$7)`, id, scope.TenantID, in.SupplierID, in.BranchID, in.ReturnNumber, total, scope.UserID); e != nil {
			return e
		}
		for _, line := range checked {
			lineID, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO supplier_return_lines(id,tenant_id,return_id,item_id,quantity,unit_cost_minor) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, id, line.item, line.qty, line.cost); e != nil {
				return e
			}
			if _, e := tx.Exec(r.Context(), `UPDATE catalog_items SET stock_on_hand=stock_on_hand-$1 WHERE tenant_id=$2 AND id=$3`, line.qty, scope.TenantID, line.item); e != nil {
				return e
			}
			if _, e := tx.Exec(r.Context(), `UPDATE branch_inventory SET quantity_on_hand=quantity_on_hand-$1,updated_at=now() WHERE tenant_id=$2 AND branch_id=$3 AND item_id=$4 AND quantity_on_hand >= $1`, line.qty, scope.TenantID, in.BranchID, line.item); e != nil { return e }
			movement, _ := ids.NewUUIDv7()
			if _, e := tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'supplier_return',$6)`, movement, scope.TenantID, in.BranchID, line.item, -line.qty, id); e != nil {
				return e
			}
		}
		auditID, _ := ids.NewUUIDv7()
		_, e := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'purchasing.supplier_return_posted','supplier_return',$4,$5)`, auditID, scope.TenantID, scope.UserID, id, mustJSON(in))
		return e
	})
	if err != nil {
		writeError(w, 409, "purchasing.return_failed", "Supplier return could not be posted; verify stock and supplier")
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "posted", "total_minor": total})
}

func (s *Server) createFiscalSubmission(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var in struct {
		Type       string `json:"document_type"`
		DocumentID string `json:"document_id"`
		Key        string `json:"idempotency_key"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&in) != nil || !uuidPattern.MatchString(in.DocumentID) || len(strings.TrimSpace(in.Key)) < 8 || !(in.Type == "receipt" || in.Type == "invoice" || in.Type == "credit_note") {
		writeError(w, 400, "fiscal.invalid_submission", "Provide a supported document type, UUID and idempotency key")
		return
	}
	id, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var exists bool
		table := "sales"
		if in.Type == "credit_note" {
			table = "supplier_returns"
		}
		if e := tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM `+table+` WHERE tenant_id=$1 AND id=$2)`, scope.TenantID, in.DocumentID).Scan(&exists); e != nil {
			return e
		}
		if !exists {
			return errors.New("fiscal source document does not exist")
		}
		_, e := tx.Exec(r.Context(), `INSERT INTO fiscal_submissions(id,tenant_id,document_type,document_id,idempotency_key,status) VALUES($1,$2,$3,$4,$5,'pending') ON CONFLICT(tenant_id,idempotency_key) DO NOTHING`, id, scope.TenantID, in.Type, in.DocumentID, in.Key)
		if e != nil {
			return e
		}
		return tx.QueryRow(r.Context(), `SELECT id::text FROM fiscal_submissions WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, in.Key).Scan(&id)
	})
	if err != nil {
		writeError(w, 400, "fiscal.queue_failed", "Fiscal document could not be queued; verify its source record")
		return
	}
	writeJSON(w, 202, map[string]any{"id": id, "status": "pending"})
}

func (s *Server) retryFiscalSubmission(w http.ResponseWriter, r *http.Request) {
	scope, err := supplyScope(r)
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("submissionID")
	if !uuidPattern.MatchString(id) {
		writeError(w, 400, "request.invalid_id", "Invalid fiscal submission id")
		return
	}
	if s.fiscalSubmitter == nil {
		writeError(w, 503, "fiscal.adapter_unconfigured", "No certified fiscal adapter is configured; submission remains pending")
		return
	}
	var doc fiscal.Document
	var status string
	var payload []byte
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var typ, key string
		var docID string
		if e := tx.QueryRow(r.Context(), `SELECT document_type,document_id::text,idempotency_key,status FROM fiscal_submissions WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, scope.TenantID, id).Scan(&typ, &docID, &key, &status); e != nil {
			return e
		}
		doc = fiscal.Document{TenantID: scope.TenantID, Type: typ, ID: docID, IdempotencyKey: key}
		if typ == "credit_note" {
			return tx.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(jsonb_build_object('item_id',item_id,'quantity',quantity,'unit_cost_minor',unit_cost_minor))::text,'[]') FROM supplier_return_lines WHERE tenant_id=$1 AND return_id=$2`, scope.TenantID, docID).Scan(&payload)
		}
		return tx.QueryRow(r.Context(), `SELECT jsonb_build_object('sale_id',s.id,'total_minor',s.total_minor,'currency',s.currency,'created_at',s.created_at,'lines',(SELECT jsonb_agg(jsonb_build_object('item_id',sl.item_id,'quantity',sl.quantity,'unit_price_minor',sl.unit_price_minor,'line_total_minor',sl.line_total_minor)) FROM sale_lines sl WHERE sl.tenant_id=s.tenant_id AND sl.sale_id=s.id))::text FROM sales s WHERE s.tenant_id=$1 AND s.id=$2`, scope.TenantID, docID).Scan(&payload)
	})
	if err != nil {
		writeError(w, 404, "fiscal.not_found", "Fiscal submission could not be found")
		return
	}
	if status == "submitted" {
		writeJSON(w, 200, map[string]any{"id": id, "status": status})
		return
	}
	doc.Payload = payload
	result, submitErr := s.fiscalSubmitter.Submit(r.Context(), doc)
	var nextStatus, lastError string
	if submitErr != nil {
		nextStatus = "failed"
		lastError = submitErr.Error()
	} else {
		nextStatus = "submitted"
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(r.Context(), `UPDATE fiscal_submissions SET status=$1,attempts=attempts+1,last_error=NULLIF($2,''),control_number=NULLIF($3,''),qr_code=NULLIF($4,''),signature=NULLIF($5,''),submitted_at=CASE WHEN $1='submitted' THEN now() ELSE submitted_at END WHERE tenant_id=$6 AND id=$7`, nextStatus, lastError, result.ControlNumber, result.QRCode, result.Signature, scope.TenantID, id)
		return e
	})
	if err != nil {
		writeError(w, 500, "fiscal.update_failed", "Fiscal result could not be recorded")
		return
	}
	if submitErr != nil {
		writeError(w, 502, "fiscal.provider_failed", "Fiscal provider failed; submission is queued for retry")
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "status": nextStatus, "control_number": result.ControlNumber, "qr_code": result.QRCode, "signature": result.Signature})
}
