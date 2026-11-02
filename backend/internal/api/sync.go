package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/core/sales"
	syncdomain "github.com/Langa2007/POS/backend/internal/core/sync"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

type syncOperationInput struct {
	ClientID         string          `json:"client_id"`
	TerminalID       string          `json:"terminal_id"`
	EntityType       string          `json:"entity_type"`
	EntityID         string          `json:"entity_id"`
	Operation        string          `json:"operation"`
	Payload          json.RawMessage `json:"payload"`
	ClientOccurredAt time.Time       `json:"client_occurred_at"`
}
type syncPushInput struct {
	Operations []syncOperationInput `json:"operations"`
}

var offlineReceiptPattern = regexp.MustCompile(`^[A-Z0-9]{1,12}-[A-Z0-9]{1,12}-[0-9]{8}$`)

func (s *Server) syncPush(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var input syncPushInput
	r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
	if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.Operations) == 0 || len(input.Operations) > 1000 {
		writeError(w, 400, "sync.invalid_batch", "Provide between 1 and 1000 operations")
		return
	}
	ops := make([]syncdomain.Operation, len(input.Operations))
	for i, op := range input.Operations {
		ops[i] = syncdomain.Operation{ClientID: op.ClientID, TerminalID: op.TerminalID, EntityType: op.EntityType, EntityID: op.EntityID, Operation: op.Operation, Payload: op.Payload, ClientOccurredAt: op.ClientOccurredAt}
	}
	ordered := syncdomain.SortOperations(ops)
	acks := make([]syncdomain.Ack, 0, len(ordered))
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		for _, op := range ordered {
			ack, err := s.processSyncOperation(r, tx, scope, op)
			if err != nil {
				return err
			}
			acks = append(acks, ack)
		}
		return nil
	})
	if err != nil {
		s.logger.Error("sync push", "error", err)
		writeError(w, 500, "sync.push_failed", "Sync batch could not be processed")
		return
	}
	writeJSON(w, 200, map[string]any{"acks": acks})
}

func operationHash(op syncdomain.Operation) []byte {
	body, _ := json.Marshal(struct {
		TerminalID, EntityType, EntityID, Operation string
		Payload                                     json.RawMessage
		Occurred                                    time.Time
	}{op.TerminalID, op.EntityType, op.EntityID, op.Operation, op.Payload, op.ClientOccurredAt.UTC()})
	sum := sha256.Sum256(body)
	return sum[:]
}

func (s *Server) processSyncOperation(r *http.Request, root pgx.Tx, scope tenancy.Scope, op syncdomain.Operation) (syncdomain.Ack, error) {
	ack := syncdomain.Ack{ClientID: op.ClientID, EntityID: op.EntityID}
	hash := operationHash(op)
	if err := syncdomain.ValidateOperation(op, time.Now()); err != nil {
		return s.recordSyncRejection(r, root, scope.TenantID, op, hash, "sync.invalid_operation", err.Error())
	}
	if !uuidPattern.MatchString(op.ClientID) || !uuidPattern.MatchString(op.EntityID) {
		return s.recordSyncRejection(r, root, scope.TenantID, op, hash, "sync.invalid_id", "client_id and entity_id must be UUIDs")
	}
	var priorStatus, priorCode, priorMessage string
	var priorHash []byte
	err := root.QueryRow(r.Context(), `SELECT status,COALESCE(error_code,''),COALESCE(error_message,''),operation_hash FROM sync_operations WHERE tenant_id=$1 AND client_id=$2`, scope.TenantID, op.ClientID).Scan(&priorStatus, &priorCode, &priorMessage, &priorHash)
	if err == nil {
		if priorHash != nil && !bytes.Equal(hash, priorHash) {
			ack.Status = "rejected"
			ack.Error = &syncdomain.Error{Code: "sync.idempotency_conflict", Message: "client_id was already used for a different operation"}
			return ack, nil
		}
		ack.Status = priorStatus
		if priorStatus == "rejected" {
			ack.Error = &syncdomain.Error{Code: priorCode, Message: priorMessage}
		}
		return ack, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ack, err
	}
	child, err := root.Begin(r.Context())
	if err != nil {
		return ack, err
	}
	code, message, applyErr := s.applyOfflineOperation(r, child, scope, op)
	if applyErr != nil {
		_ = child.Rollback(r.Context())
		if strings.HasSuffix(code, "_failed") || code == "sync.policy_unavailable" {
			s.logger.Error("offline operation rejected by storage error", "code", code, "error", applyErr)
			message = "The server could not safely save this sale. Retry sync or contact support."
		}
		return s.recordSyncRejection(r, root, scope.TenantID, op, hash, code, message)
	}
	id, _ := ids.NewUUIDv7()
	if _, err = child.Exec(r.Context(), `INSERT INTO sync_operations(id,tenant_id,client_id,terminal_id,entity_type,entity_id,operation,payload,client_occurred_at,status,operation_hash) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9,'accepted',$10)`, id, scope.TenantID, op.ClientID, op.TerminalID, op.EntityType, op.EntityID, op.Operation, op.Payload, op.ClientOccurredAt, hash); err != nil {
		_ = child.Rollback(r.Context())
		return ack, err
	}
	if err = child.Commit(r.Context()); err != nil {
		return ack, err
	}
	ack.Status = "accepted"
	ack.Result = map[string]any{"sale_id": op.EntityID, "message": message}
	return ack, nil
}

func (s *Server) recordSyncRejection(r *http.Request, tx pgx.Tx, tenantID string, op syncdomain.Operation, hash []byte, code, message string) (syncdomain.Ack, error) {
	ack := syncdomain.Ack{ClientID: op.ClientID, EntityID: op.EntityID, Status: "rejected", Error: &syncdomain.Error{Code: code, Message: message}}
	if !uuidPattern.MatchString(op.ClientID) {
		return ack, nil
	}
	id, _ := ids.NewUUIDv7()
	terminalID := op.TerminalID
	if terminalID != "" && !uuidPattern.MatchString(terminalID) {
		terminalID = ""
	}
	_, err := tx.Exec(r.Context(), `INSERT INTO sync_operations(id,tenant_id,client_id,terminal_id,entity_type,entity_id,operation,payload,client_occurred_at,status,error_code,error_message,operation_hash) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,COALESCE($8,'{}'::jsonb),$9,'rejected',$10,$11,$12) ON CONFLICT(tenant_id,client_id) DO NOTHING`, id, tenantID, op.ClientID, terminalID, op.EntityType, op.EntityID, op.Operation, op.Payload, op.ClientOccurredAt, code, message, hash)
	return ack, err
}

type offlineSalePayload struct {
	BranchID      string    `json:"branch_id"`
	ShiftID       string    `json:"shift_id"`
	ReceiptNumber string    `json:"receipt_number"`
	OfflineSince  time.Time `json:"offline_since"`
	ReceivedMinor int64     `json:"received_minor"`
	Lines         []struct {
		ItemID         string `json:"item_id"`
		Quantity       int64  `json:"quantity"`
		DiscountMinor  int64  `json:"discount_minor"`
		UnitPriceMinor int64  `json:"unit_price_minor"`
	} `json:"lines"`
}

func (s *Server) applyOfflineOperation(r *http.Request, tx pgx.Tx, scope tenancy.Scope, op syncdomain.Operation) (string, string, error) {
	if op.EntityType != "sale" || op.Operation != syncdomain.OperationCreate {
		return "sync.unsupported_operation", "", errors.New("only offline cash sale creation is currently supported")
	}
	if scope.TerminalID != "" && op.TerminalID != scope.TerminalID {
		return "sync.terminal_mismatch", "", errors.New("operation terminal does not match the authenticated terminal")
	}
	var input offlineSalePayload
	if err := json.Unmarshal(op.Payload, &input); err != nil || len(input.Lines) == 0 {
		return "sync.invalid_sale", "", errors.New("offline sale payload is invalid")
	}
	if !uuidPattern.MatchString(input.BranchID) || !uuidPattern.MatchString(input.ShiftID) {
		return "sync.invalid_sale", "", errors.New("offline sale requires a valid cached branch and shift")
	}
	if !offlineReceiptPattern.MatchString(input.ReceiptNumber) {
		return "sync.invalid_receipt", "Offline receipt number is invalid. Reopen the terminal session and try again.", errors.New("offline receipt number is invalid")
	}
	if scope.BranchID != "" && input.BranchID != scope.BranchID {
		return "sync.branch_mismatch", "", errors.New("offline sale branch does not match the authenticated branch")
	}
	var policy struct {
		Hours         int
		MaxMinor      int64
		AllowNegative bool
	}
	err := tx.QueryRow(r.Context(), `SELECT COALESCE(p.max_offline_hours,24),COALESCE(p.max_sale_minor,500000),COALESCE(p.allow_negative_stock,false) FROM (SELECT 1) x LEFT JOIN offline_sync_policies p ON p.tenant_id=$1`, scope.TenantID).Scan(&policy.Hours, &policy.MaxMinor, &policy.AllowNegative)
	if err != nil {
		return "sync.policy_unavailable", "", err
	}
	now := time.Now().UTC()
	if input.OfflineSince.IsZero() || input.OfflineSince.After(op.ClientOccurredAt) || now.Sub(input.OfflineSince) > time.Duration(policy.Hours)*time.Hour {
		return "sync.offline_window_exceeded", "", errors.New("offline operation is outside the allowed offline window")
	}
	var shiftBranch, shiftUser string
	var opened time.Time
	var closed *time.Time
	err = tx.QueryRow(r.Context(), `SELECT branch_id::text,user_id::text,opened_at,closed_at FROM shifts WHERE tenant_id=$1 AND id=$2`, scope.TenantID, input.ShiftID).Scan(&shiftBranch, &shiftUser, &opened, &closed)
	if err != nil || shiftBranch != input.BranchID || shiftUser != scope.UserID || op.ClientOccurredAt.Before(opened) || closed != nil && op.ClientOccurredAt.After(*closed) {
		return "sync.shift_invalid", "", errors.New("cached shift is unavailable or the sale timestamp is outside the shift")
	}
	if op.TerminalID != "" {
		var active bool
		if err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM terminals WHERE tenant_id=$1 AND id=$2 AND revoked_at IS NULL AND branch_id=$3)`, scope.TenantID, op.TerminalID, input.BranchID).Scan(&active); err != nil || !active {
			return "sync.terminal_revoked", "", errors.New("terminal is invalid or revoked")
		}
	}
	type computedLine struct {
		id                                               string
		qty, price, discount, total, stock, currentPrice int64
		kind                                             string
		changedPrice                                     bool
	}
	lines := make([]computedLine, 0, len(input.Lines))
	seenItems := make(map[string]struct{}, len(input.Lines))
	var total int64
	for _, raw := range input.Lines {
		if !uuidPattern.MatchString(raw.ItemID) || raw.Quantity <= 0 || raw.UnitPriceMinor < 0 || raw.DiscountMinor != 0 {
			return "sync.invalid_sale", "", errors.New("offline sale line is invalid")
		}
		if _, duplicate := seenItems[raw.ItemID]; duplicate {
			return "sync.invalid_sale", "Offline sale has duplicate item lines", errors.New("offline sale has duplicate item lines")
		}
		seenItems[raw.ItemID] = struct{}{}
		var currentPrice, stock int64
		var kind string
		err = tx.QueryRow(r.Context(), `SELECT price_minor,stock_on_hand,kind FROM catalog_items WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, scope.TenantID, raw.ItemID).Scan(&currentPrice, &stock, &kind)
		if err != nil {
			return "sync.item_missing", "", errors.New("an item in the offline sale no longer exists")
		}
		if kind == "stocked" { if err = tx.QueryRow(r.Context(), `SELECT quantity_on_hand FROM branch_inventory WHERE tenant_id=$1 AND branch_id=$2 AND item_id=$3 FOR UPDATE`, scope.TenantID, input.BranchID, raw.ItemID).Scan(&stock); err != nil { return "sync.item_missing", "", errors.New("item is not stocked at this branch") } }
		if kind == "stocked" && stock < raw.Quantity && !policy.AllowNegative {
			return "sync.insufficient_stock", "", fmt.Errorf("insufficient stock for item %s", raw.ItemID)
		}
		if raw.Quantity > 0 && raw.UnitPriceMinor > 0 && raw.Quantity > math.MaxInt64/raw.UnitPriceMinor {
			return "sync.invalid_sale", "", errors.New("sale line amount overflow")
		}
		lineTotal, calcErr := sales.Line{Quantity: raw.Quantity, UnitPriceMinor: raw.UnitPriceMinor, DiscountMinor: raw.DiscountMinor}.TotalMinor()
		if calcErr != nil {
			return "sync.invalid_sale", "", calcErr
		}
		if lineTotal > math.MaxInt64-total {
			return "sync.invalid_sale", "", errors.New("sale total overflow")
		}
		total += lineTotal
		lines = append(lines, computedLine{raw.ItemID, raw.Quantity, raw.UnitPriceMinor, raw.DiscountMinor, lineTotal, stock, currentPrice, kind, currentPrice != raw.UnitPriceMinor})
	}
	if total <= 0 || total > policy.MaxMinor || input.ReceivedMinor < total {
		return "sync.sale_limit", "", errors.New("offline cash received or sale total violates the offline sale limit")
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO sales(id,tenant_id,branch_id,cashier_id,shift_id,total_minor,idempotency_key,status,created_at,receipt_number) VALUES($1,$2,$3,$4,$5,$6,$7,'completed',$8,$9)`, op.EntityID, scope.TenantID, input.BranchID, scope.UserID, input.ShiftID, total, "offline:"+op.ClientID, op.ClientOccurredAt, input.ReceiptNumber); err != nil {
		return "sync.sale_conflict", "", errors.New("offline sale conflicts with an existing sale")
	}
	for _, line := range lines {
		lineID, _ := ids.NewUUIDv7()
		if _, err = tx.Exec(r.Context(), `INSERT INTO sale_lines(id,tenant_id,sale_id,item_id,quantity,unit_price_minor,discount_minor,line_total_minor) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, lineID, scope.TenantID, op.EntityID, line.id, line.qty, line.price, line.discount, line.total); err != nil {
			return "sync.sale_write_failed", "", err
		}
		if line.kind == "stocked" {
			if _, err = tx.Exec(r.Context(), `UPDATE catalog_items SET stock_on_hand=stock_on_hand-$1 WHERE tenant_id=$2 AND id=$3`, line.qty, scope.TenantID, line.id); err != nil {
				return "sync.stock_write_failed", "", err
			}
			if _, err = tx.Exec(r.Context(), `UPDATE branch_inventory SET quantity_on_hand=quantity_on_hand-$1,updated_at=now() WHERE tenant_id=$2 AND branch_id=$3 AND item_id=$4`, line.qty, scope.TenantID, input.BranchID, line.id); err != nil { return "sync.stock_write_failed", "", err }
			movementID, _ := ids.NewUUIDv7()
			if _, err = tx.Exec(r.Context(), `INSERT INTO stock_movements(id,tenant_id,branch_id,item_id,quantity_delta,reason,reference_id) VALUES($1,$2,$3,$4,$5,'offline_sale',$6)`, movementID, scope.TenantID, input.BranchID, line.id, -line.qty, op.EntityID); err != nil {
				return "sync.stock_write_failed", "", err
			}
		}
		if line.changedPrice {
			auditID, _ := ids.NewUUIDv7()
			if _, err = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,before_data,after_data) VALUES($1,$2,$3,'offline.sale_price_variance','sale_line',$4,jsonb_build_object('current_price_minor',$5::bigint),jsonb_build_object('offline_price_minor',$6::bigint,'sale_id',$7::text))`, auditID, scope.TenantID, scope.UserID, lineID, line.currentPrice, line.price, op.EntityID); err != nil {
				return "sync.audit_failed", "", err
			}
		}
	}
	tenderID, _ := ids.NewUUIDv7()
	if _, err = tx.Exec(r.Context(), `INSERT INTO cash_tenders(id,tenant_id,sale_id,amount_minor,received_minor,change_minor) VALUES($1,$2,$3,$4,$5,$6)`, tenderID, scope.TenantID, op.EntityID, total, input.ReceivedMinor, input.ReceivedMinor-total); err != nil {
		return "sync.tender_failed", "", err
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,method,amount_minor,idempotency_key) VALUES($1,$2,$3,'cash',$4,$5)`, tenderID, scope.TenantID, op.EntityID, total, "offline:"+op.ClientID); err != nil {
		return "sync.tender_failed", "", err
	}
	journalID, _ := ids.NewUUIDv7()
	if _, err = tx.Exec(r.Context(), `INSERT INTO ledger_journals(id,tenant_id,reference_type,reference_id,description,posted_at) VALUES($1,$2,'sale',$3,'Offline cash sale',$4)`, journalID, scope.TenantID, op.EntityID, op.ClientOccurredAt); err != nil {
		return "sync.ledger_failed", "", err
	}
	cashAccount, revenueAccount, err := ensureLedgerAccounts(r, tx, scope.TenantID)
	if err != nil {
		return "sync.ledger_failed", "", err
	}
	for _, posting := range []struct {
		account       string
		debit, credit int64
	}{{cashAccount, total, 0}, {revenueAccount, 0, total}} {
		lineID, _ := ids.NewUUIDv7()
		if _, err = tx.Exec(r.Context(), `INSERT INTO ledger_lines(id,tenant_id,journal_id,account_id,debit_minor,credit_minor) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, journalID, posting.account, posting.debit, posting.credit); err != nil {
			return "sync.ledger_failed", "", err
		}
	}
	auditID, _ := ids.NewUUIDv7()
	outboxID, _ := ids.NewUUIDv7()
	if _, err = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'sale.completed.offline','sale',$4,jsonb_build_object('total_minor',$5::bigint,'client_id',$6::text,'client_occurred_at',$7::timestamptz))`, auditID, scope.TenantID, scope.UserID, op.EntityID, total, op.ClientID, op.ClientOccurredAt); err != nil {
		return "sync.audit_failed", "", err
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO outbox(id,tenant_id,event_type,payload,occurred_at) VALUES($1,$2,'sale.completed.v1',jsonb_build_object('sale_id',$3::text,'total_minor',$4::bigint,'offline',true),$5)`, outboxID, scope.TenantID, op.EntityID, total, op.ClientOccurredAt); err != nil {
		return "sync.outbox_failed", "", err
	}
	for _, line := range lines {
		if line.kind != "stocked" {
			continue
		}
		var item map[string]any
		var id, name, sku, barcode, currency, kind string
		var price, stock, reorder int64
		err = tx.QueryRow(r.Context(), `SELECT id::text,name,sku,COALESCE(barcode,''),currency,kind,price_minor,stock_on_hand,reorder_level FROM catalog_items WHERE tenant_id=$1 AND id=$2`, scope.TenantID, line.id).Scan(&id, &name, &sku, &barcode, &currency, &kind, &price, &stock, &reorder)
		if err != nil {
			return "sync.catalog_change_failed", "", err
		}
		item = map[string]any{"id": id, "name": name, "sku": sku, "barcode": barcode, "currency": strings.TrimSpace(currency), "kind": kind, "price_minor": price, "stock_on_hand": stock, "reorder_level": reorder}
		if _, err = tx.Exec(r.Context(), `INSERT INTO sync_changes(tenant_id,entity_type,entity_id,operation,payload,occurred_at) VALUES($1,'master.catalog_item',$2,'update',$3,$4)`, scope.TenantID, id, mustJSON(item), op.ClientOccurredAt); err != nil {
			return "sync.catalog_change_failed", "", err
		}
	}
	return "", "offline sale accepted", nil
}

func (s *Server) syncPull(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	cursor, limit := int64(0), 500
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		cursor, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || cursor < 0 {
			writeError(w, 400, "sync.invalid_cursor", "cursor must be a non-negative integer")
			return
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, e := strconv.Atoi(raw)
		if e != nil || parsed < 1 || parsed > 1000 {
			writeError(w, 400, "sync.invalid_limit", "limit must be between 1 and 1000")
			return
		}
		limit = parsed
	}
	changes := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT cursor,entity_type,entity_id,operation,payload,occurred_at FROM sync_changes WHERE tenant_id=$1 AND cursor>$2 ORDER BY cursor LIMIT $3`, scope.TenantID, cursor, limit)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var next int64
			var typ, id, op string
			var payload []byte
			var occurred time.Time
			if e = rows.Scan(&next, &typ, &id, &op, &payload, &occurred); e != nil {
				return e
			}
			var decoded any
			if e = json.NewDecoder(bytes.NewReader(payload)).Decode(&decoded); e != nil {
				return e
			}
			changes = append(changes, map[string]any{"cursor": next, "entity_type": typ, "entity_id": id, "operation": op, "payload": decoded, "occurred_at": occurred})
			cursor = next
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("sync pull", "error", err)
		writeError(w, 500, "sync.pull_failed", "Sync changes could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"changes": changes, "next_cursor": cursor})
}

func (s *Server) getSyncPolicy(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var hours int
	var max int64
	var negative bool
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(r.Context(), `INSERT INTO offline_sync_policies(tenant_id) VALUES($1) ON CONFLICT DO NOTHING`, scope.TenantID); e != nil {
			return e
		}
		return tx.QueryRow(r.Context(), `SELECT max_offline_hours,max_sale_minor,allow_negative_stock FROM offline_sync_policies WHERE tenant_id=$1`, scope.TenantID).Scan(&hours, &max, &negative)
	})
	if err != nil {
		writeError(w, 500, "sync.policy_unavailable", "Offline policy could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"max_offline_hours": hours, "max_sale_minor": max, "allow_negative_stock": negative})
}

func (s *Server) updateSyncPolicy(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct {
		Hours    int   `json:"max_offline_hours"`
		Max      int64 `json:"max_sale_minor"`
		Negative bool  `json:"allow_negative_stock"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || input.Hours < 1 || input.Hours > 168 || input.Max < 1 {
		writeError(w, 400, "sync.policy_invalid", "Offline duration must be 1–168 hours and max sale amount must be positive")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(r.Context(), `INSERT INTO offline_sync_policies(tenant_id,max_offline_hours,max_sale_minor,allow_negative_stock,updated_at) VALUES($1,$2,$3,$4,now()) ON CONFLICT(tenant_id) DO UPDATE SET max_offline_hours=excluded.max_offline_hours,max_sale_minor=excluded.max_sale_minor,allow_negative_stock=excluded.allow_negative_stock,updated_at=now()`, scope.TenantID, input.Hours, input.Max, input.Negative)
		return e
	})
	if err != nil {
		writeError(w, 400, "sync.policy_save_failed", "Offline policy could not be saved")
		return
	}
	writeJSON(w, 200, map[string]any{"max_offline_hours": input.Hours, "max_sale_minor": input.Max, "allow_negative_stock": input.Negative})
}
