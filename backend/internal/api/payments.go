package api

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/adapters/mpesa"
	"github.com/Langa2007/POS/backend/internal/core/payments"
	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

type tenderInput struct {
	Method      string `json:"method"`
	AmountMinor int64  `json:"amount_minor"`
	Reference   string `json:"reference"`
}

type allocatePaymentRequest struct {
	Tenders []tenderInput `json:"tenders"`
}

func (s *Server) allocatePayment(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	saleID := r.PathValue("saleID")
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, 400, "request.idempotency_required", "Idempotency-Key is required")
		return
	}
	var input allocatePaymentRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&input) != nil || !uuidPattern.MatchString(saleID) || len(input.Tenders) != 1 {
		writeError(w, 400, "payments.invalid_allocation", "Provide a sale and payment tenders")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var existingSale string
		var existingAmount int64
		if err := tx.QueryRow(r.Context(), `SELECT sale_id::text,amount_minor FROM payment_tenders WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, key+"-cash").Scan(&existingSale, &existingAmount); err == nil {
			if len(input.Tenders) != 1 || existingSale != saleID || existingAmount != input.Tenders[0].AmountMinor {
				return errors.New("idempotency key was already used for a different allocation")
			}
			return nil
		} else if err != pgx.ErrNoRows {
			return err
		}
		var total, paid int64
		if err := tx.QueryRow(r.Context(), `SELECT total_minor FROM sales WHERE tenant_id=$1 AND id=$2 AND status='pending_payment' FOR UPDATE`, scope.TenantID, saleID).Scan(&total); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, scope.TenantID, saleID).Scan(&paid); err != nil {
			return err
		}
		due := total - paid
		tenders := make([]payments.Tender, len(input.Tenders))
		for i, tender := range input.Tenders {
			if tender.Method != "cash" {
				return errors.New("manual allocation supports cash only; use provider callbacks for electronic tenders")
			}
			tenders[i] = payments.Tender{Method: tender.Method, AmountMinor: tender.AmountMinor, Reference: tender.Reference}
		}
		if err := payments.ValidateAllocation(due, tenders); err != nil {
			return err
		}
		for _, tender := range input.Tenders {
			tenderID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,method,amount_minor,provider_reference,idempotency_key) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7) ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, tenderID, scope.TenantID, saleID, tender.Method, tender.AmountMinor, tender.Reference, key+"-"+tender.Method); err != nil {
				return err
			}
			if tender.Method == "cash" {
				_, err := tx.Exec(r.Context(), `INSERT INTO cash_tenders(id,tenant_id,sale_id,amount_minor,received_minor,change_minor) VALUES($1,$2,$3,$4,$4,0) ON CONFLICT DO NOTHING`, tenderID, scope.TenantID, saleID, tender.AmountMinor)
				if err != nil {
					return err
				}
			}
		}
		journalID, _ := ids.NewUUIDv7()
		if _, err := tx.Exec(r.Context(), `INSERT INTO ledger_journals(id,tenant_id,reference_type,reference_id,description) VALUES($1,$2,'sale_payment',$3,'Cash payment allocation')`, journalID, scope.TenantID, key); err != nil {
			return err
		}
		cashAccount, revenueAccount, err := ensureLedgerAccounts(r, tx, scope.TenantID)
		if err != nil {
			return err
		}
		for _, posting := range []struct {
			account string
			debit   int64
			credit  int64
		}{{cashAccount, due, 0}, {revenueAccount, 0, due}} {
			lineID, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO ledger_lines(id,tenant_id,journal_id,account_id,debit_minor,credit_minor) VALUES($1,$2,$3,$4,$5,$6)`, lineID, scope.TenantID, journalID, posting.account, posting.debit, posting.credit); err != nil {
				return err
			}
		}
		if paid+due >= total {
			_, err := tx.Exec(r.Context(), `UPDATE sales SET status='completed' WHERE tenant_id=$1 AND id=$2 AND status='pending_payment'`, scope.TenantID, saleID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeError(w, 400, "payments.allocation_failed", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"sale_id": saleID, "status": "captured"})
}

type intentRequest struct {
	SaleID           string `json:"sale_id"`
	AmountMinor      int64  `json:"amount_minor"`
	Phone            string `json:"phone"`
	AccountReference string `json:"account_reference"`
}

func (s *Server) initiatePayment(w http.ResponseWriter, r *http.Request) {
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
	var input intentRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&input) != nil || !uuidPattern.MatchString(input.SaleID) || input.AmountMinor <= 0 || !validPhone(input.Phone) {
		writeError(w, 400, "payments.invalid_intent", "Provide sale, positive amount, and a valid phone number")
		return
	}
	intentID, _ := ids.NewUUIDv7()
	var existing struct {
		ID, Status, CheckoutID, SaleID, Phone string
		Amount                                int64
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(r.Context(), `SELECT id::text,status,COALESCE(provider_reference,''),sale_id::text,amount_minor,phone FROM payment_intents WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, key).Scan(&existing.ID, &existing.Status, &existing.CheckoutID, &existing.SaleID, &existing.Amount, &existing.Phone)
		return err
	})
	if err == nil {
		if existing.SaleID != input.SaleID || existing.Amount != input.AmountMinor || normalizePhoneForComparison(existing.Phone) != normalizePhoneForComparison(input.Phone) {
			writeError(w, 409, "payments.idempotency_conflict", "Idempotency-Key was already used for a different payment")
			return
		}
		writeJSON(w, 200, map[string]any{"intent_id": existing.ID, "status": existing.Status, "checkout_request_id": existing.CheckoutID})
		return
	}
	if err != pgx.ErrNoRows {
		writeError(w, 500, "payments.unavailable", "Payment service is unavailable")
		return
	}
	provider, _, providerErr := s.loadMPesaConfig(r, scope.TenantID)
	if providerErr != nil {
		writeError(w, 503, "payments.provider_unavailable", "M-Pesa is not configured or available")
		return
	}
	var due int64
	var saleReference string
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var total, paid, reserved int64
		if err := tx.QueryRow(r.Context(), `SELECT total_minor,payment_reference FROM sales WHERE tenant_id=$1 AND id=$2 AND status IN ('pending_payment','completed') FOR UPDATE`, scope.TenantID, input.SaleID).Scan(&total, &saleReference); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, scope.TenantID, input.SaleID).Scan(&paid); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_intents WHERE tenant_id=$1 AND sale_id=$2 AND status='pending'`, scope.TenantID, input.SaleID).Scan(&reserved); err != nil {
			return err
		}
		due = total - paid - reserved
		if input.AmountMinor > due || due <= 0 {
			return errors.New("payment amount exceeds outstanding sale balance")
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO payment_intents(id,tenant_id,sale_id,amount_minor,provider,method,phone,status,idempotency_key) VALUES($1,$2,$3,$4,'mpesa','mpesa_stk',$5,'pending',$6)`, intentID, scope.TenantID, input.SaleID, input.AmountMinor, input.Phone, key)
		return err
	})
	if err != nil {
		writeError(w, 400, "payments.invalid_amount", err.Error())
		return
	}
	accountReference := input.AccountReference
	if accountReference != "" && accountReference != saleReference {
		writeError(w, 400, "payments.invalid_reference", "Account reference must match the sale payment reference")
		return
	}
	accountReference = saleReference
	response, providerErr := provider.InitiateSTK(r.Context(), mpesa.STKRequest{Phone: input.Phone, AmountMinor: input.AmountMinor, AccountReference: accountReference, Description: "POS sale"})
	if providerErr != nil {
		_ = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
			_, updateErr := tx.Exec(r.Context(), `UPDATE payment_intents SET status='failed',failure_code='provider_initiation_failed',updated_at=now() WHERE tenant_id=$1 AND id=$2 AND status='pending'`, scope.TenantID, intentID)
			return updateErr
		})
		writeError(w, 502, "payments.provider_error", "M-Pesa could not initiate the payment")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `UPDATE payment_intents SET provider_reference=$1,merchant_request_id=$2,raw_provider_response=$3,updated_at=now() WHERE tenant_id=$4 AND id=$5 AND status='pending'`, response.CheckoutRequestID, response.MerchantRequestID, mustJSON(response), scope.TenantID, intentID)
		return err
	})
	if err != nil {
		writeError(w, 400, "payments.initiation_failed", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"intent_id": intentID, "status": payments.Pending, "checkout_request_id": response.CheckoutRequestID})
}

func (s *Server) mpesaCallback(w http.ResponseWriter, r *http.Request) {
	tenantID, callbackToken, tokenErr := callbackCredentials(r.PathValue("token"))
	if tokenErr != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		writeError(w, 400, "payments.invalid_callback", "Invalid callback payload")
		return
	}
	provider, _, err := s.loadMPesaConfig(r, tenantID)
	if err != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	if err = s.verifyMPesaCallbackToken(r, tenantID, callbackToken); err != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	callback, err := provider.ParseCallback(body)
	if err != nil {
		writeError(w, 400, "payments.invalid_callback", "Invalid M-Pesa callback payload")
		return
	}
	err = s.db.TenantTx(r.Context(), tenantID, func(tx pgx.Tx) error {
		callbackID, _ := ids.NewUUIDv7()
		payload := mustJSON(callback)
		tag, err := tx.Exec(r.Context(), `INSERT INTO payment_callbacks(id,tenant_id,provider,callback_key,payload,processed_at) VALUES($1,$2,'mpesa',$3,$4,now()) ON CONFLICT (tenant_id,provider,callback_key) DO NOTHING`, callbackID, tenantID, callback.CheckoutRequestID, payload)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		var intentID, status, saleID, intentPhone string
		var amount int64
		if err := tx.QueryRow(r.Context(), `SELECT id::text,status,sale_id::text,amount_minor,phone FROM payment_intents WHERE tenant_id=$1 AND provider_reference=$2 FOR UPDATE`, tenantID, callback.CheckoutRequestID).Scan(&intentID, &status, &saleID, &amount, &intentPhone); err != nil {
			return err
		}
		if callback.ResultCode == 0 && (callback.AmountMinor != amount || callback.ReceiptNumber == "" || normalizePhoneForComparison(callback.Phone) != normalizePhoneForComparison(intentPhone)) {
			return errors.New("provider callback payment details failed verification")
		}
		target := payments.Captured
		if callback.ResultCode != 0 {
			target = payments.Failed
		}
		if err := payments.Transition(payments.Status(status), target); err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(), `UPDATE payment_intents SET status=$1,raw_payload=$2,callback_received_at=now(),captured_at=CASE WHEN $1='captured' THEN now() ELSE captured_at END,updated_at=now(),failure_code=NULLIF($3,'') WHERE tenant_id=$4 AND id=$5`, target, payload, callback.ResultDescription, tenantID, intentID)
		if err != nil || target != payments.Captured {
			return err
		}
		tenderID, _ := ids.NewUUIDv7()
		if _, err = tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,intent_id,method,amount_minor,provider_reference,idempotency_key) VALUES($1,$2,$3,$4,'mpesa_stk',$5,$6,$7) ON CONFLICT(tenant_id,idempotency_key) DO NOTHING`, tenderID, tenantID, saleID, intentID, amount, callback.ReceiptNumber, "mpesa-intent-"+intentID); err != nil {
			return err
		}
		if err = postPaymentJournal(r, tx, tenantID, intentID, saleID, amount, callback.ReceiptNumber); err != nil {
			return err
		}
		var saleTotal, paid int64
		if err = tx.QueryRow(r.Context(), `SELECT total_minor FROM sales WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenantID, saleID).Scan(&saleTotal); err != nil {
			return err
		}
		if err = tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, tenantID, saleID).Scan(&paid); err != nil {
			return err
		}
		if paid >= saleTotal {
			_, err = tx.Exec(r.Context(), `UPDATE sales SET status='completed' WHERE tenant_id=$1 AND id=$2 AND status='pending_payment'`, tenantID, saleID)
		}
		return err
	})
	if err != nil {
		writeError(w, 400, "payments.callback_failed", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"ResultCode": "0", "ResultDesc": "Accepted"})
}

func (s *Server) mpesaC2BCallback(w http.ResponseWriter, r *http.Request) {
	tenantID, token, err := callbackCredentials(r.PathValue("token"))
	if err != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	if err = s.verifyMPesaCallbackToken(r, tenantID, token); err != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	var payload struct {
		TransID       string          `json:"TransID"`
		TransAmount   json.RawMessage `json:"TransAmount"`
		BillRefNumber string          `json:"BillRefNumber"`
		MSISDN        string          `json:"MSISDN"`
		TransTime     string          `json:"TransTime"`
	}
	if json.NewDecoder(r.Body).Decode(&payload) != nil || payload.TransID == "" || payload.BillRefNumber == "" {
		writeError(w, 400, "payments.invalid_callback", "Invalid M-Pesa C2B callback")
		return
	}
	amountText := strings.Trim(string(payload.TransAmount), `"`)
	amount, err := parseMpesaAmount(amountText)
	if err != nil || amount <= 0 {
		writeError(w, 400, "payments.invalid_callback", "Invalid M-Pesa C2B amount")
		return
	}
	err = s.db.TenantTx(r.Context(), tenantID, func(tx pgx.Tx) error {
		callbackID, _ := ids.NewUUIDv7()
		tag, e := tx.Exec(r.Context(), `INSERT INTO payment_callbacks(id,tenant_id,provider,callback_key,payload,processed_at) VALUES($1,$2,'mpesa',$3,$4,now()) ON CONFLICT DO NOTHING`, callbackID, tenantID, "c2b-"+payload.TransID, mustJSON(payload))
		if e != nil || tag.RowsAffected() == 0 {
			return e
		}
		var total, paid int64
		var saleID string
		if e = tx.QueryRow(r.Context(), `SELECT id::text,total_minor FROM sales WHERE tenant_id=$1 AND payment_reference=$2 AND status IN ('pending_payment','completed') FOR UPDATE`, tenantID, payload.BillRefNumber).Scan(&saleID, &total); e != nil {
			return storeUnallocated(r, tx, tenantID, payload.TransID, amount, payload.BillRefNumber, payload)
		}
		if e = tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, tenantID, saleID).Scan(&paid); e != nil {
			return e
		}
		if amount > total-paid {
			return storeUnallocated(r, tx, tenantID, payload.TransID, amount, payload.BillRefNumber, payload)
		}
		intentID, _ := ids.NewUUIDv7()
		if _, e = tx.Exec(r.Context(), `INSERT INTO payment_intents(id,tenant_id,sale_id,amount_minor,provider,method,status,provider_reference,idempotency_key,captured_at,raw_payload) VALUES($1,$2,$3,$4,'mpesa','mpesa_c2b','captured',$5,$6,now(),$7)`, intentID, tenantID, saleID, amount, payload.TransID, "c2b-"+payload.TransID, mustJSON(payload)); e != nil {
			return e
		}
		tenderID, _ := ids.NewUUIDv7()
		if _, e = tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,intent_id,method,amount_minor,provider_reference,idempotency_key) VALUES($1,$2,$3,$4,'mpesa_c2b',$5,$6,$7)`, tenderID, tenantID, saleID, intentID, amount, payload.TransID, "c2b-"+payload.TransID); e != nil {
			return e
		}
		if e = postPaymentJournal(r, tx, tenantID, intentID, saleID, amount, payload.TransID); e != nil {
			return e
		}
		if paid+amount >= total {
			_, e = tx.Exec(r.Context(), `UPDATE sales SET status='completed' WHERE tenant_id=$1 AND id=$2 AND status='pending_payment'`, tenantID, saleID)
		}
		return e
	})
	if err != nil {
		writeError(w, 400, "payments.c2b_failed", "C2B payment could not be processed")
		return
	}
	writeJSON(w, 200, map[string]string{"ResultCode": "0", "ResultDesc": "Accepted"})
}

func (s *Server) mpesaC2BValidation(w http.ResponseWriter, r *http.Request) {
	tenantID, token, err := callbackCredentials(r.PathValue("token"))
	if err != nil || s.verifyMPesaCallbackToken(r, tenantID, token) != nil {
		writeError(w, 404, "payments.callback_not_found", "Callback endpoint not found")
		return
	}
	writeJSON(w, 200, map[string]string{"ResultCode": "0", "ResultDesc": "Accepted"})
}

func (s *Server) verifyMPesaCallbackToken(r *http.Request, tenantID, token string) error {
	var hash []byte
	err := s.db.TenantTx(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT callback_token_hash FROM mpesa_configurations WHERE tenant_id=$1 AND enabled`, tenantID).Scan(&hash)
	})
	if err != nil || !callbackTokenValid(hash, token) {
		return errors.New("not found")
	}
	return nil
}

func storeUnallocated(r *http.Request, tx pgx.Tx, tenantID, receipt string, amount int64, reference string, payload any) error {
	id, _ := ids.NewUUIDv7()
	_, err := tx.Exec(r.Context(), `INSERT INTO unallocated_payments(id,tenant_id,provider,provider_reference,amount_minor,account_reference,payload) VALUES($1,$2,'mpesa',$3,$4,$5,$6) ON CONFLICT DO NOTHING`, id, tenantID, receipt, amount, reference, mustJSON(payload))
	return err
}

func (s *Server) allocateUnallocatedPayment(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	id := r.PathValue("paymentID")
	if !uuidPattern.MatchString(id) {
		writeError(w, 404, "payments.unallocated_not_found", "Unallocated payment was not found")
		return
	}
	var input struct {
		SaleID string `json:"sale_id"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || !uuidPattern.MatchString(input.SaleID) {
		writeError(w, 400, "payments.invalid_allocation", "A valid sale_id is required")
		return
	}
	var intentID string
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var receipt string
		var amount int64
		var status string
		if err := tx.QueryRow(r.Context(), `SELECT provider_reference,amount_minor,status FROM unallocated_payments WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, scope.TenantID, id).Scan(&receipt, &amount, &status); err != nil {
			return err
		}
		if status != "unallocated" {
			return errors.New("payment is already resolved")
		}
		var total, paid int64
		if err := tx.QueryRow(r.Context(), `SELECT total_minor FROM sales WHERE tenant_id=$1 AND id=$2 AND status IN ('pending_payment','completed') FOR UPDATE`, scope.TenantID, input.SaleID).Scan(&total); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, scope.TenantID, input.SaleID).Scan(&paid); err != nil {
			return err
		}
		if amount > total-paid {
			return errors.New("unallocated amount exceeds sale balance")
		}
		intentID, _ = ids.NewUUIDv7()
		if _, err := tx.Exec(r.Context(), `INSERT INTO payment_intents(id,tenant_id,sale_id,amount_minor,provider,method,status,provider_reference,idempotency_key,captured_at) VALUES($1,$2,$3,$4,'mpesa','mpesa_c2b','captured',$5,$6,now())`, intentID, scope.TenantID, input.SaleID, amount, receipt, "unallocated-"+id); err != nil {
			return err
		}
		tenderID, _ := ids.NewUUIDv7()
		if _, err := tx.Exec(r.Context(), `INSERT INTO payment_tenders(id,tenant_id,sale_id,intent_id,method,amount_minor,provider_reference,idempotency_key) VALUES($1,$2,$3,$4,'mpesa_c2b',$5,$6,$7)`, tenderID, scope.TenantID, input.SaleID, intentID, amount, receipt, "unallocated-"+id); err != nil {
			return err
		}
		if err := postPaymentJournal(r, tx, scope.TenantID, intentID, input.SaleID, amount, receipt); err != nil {
			return err
		}
		if paid+amount >= total {
			if _, err := tx.Exec(r.Context(), `UPDATE sales SET status='completed' WHERE tenant_id=$1 AND id=$2 AND status='pending_payment'`, scope.TenantID, input.SaleID); err != nil {
				return err
			}
		}
		_, err := tx.Exec(r.Context(), `UPDATE unallocated_payments SET status='allocated' WHERE tenant_id=$1 AND id=$2`, scope.TenantID, id)
		return err
	})
	if err != nil {
		writeError(w, 400, "payments.unallocated_allocation_failed", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "allocated", "intent_id": intentID})
}

func parseMpesaAmount(raw string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("invalid amount")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, errors.New("invalid amount")
	}
	fraction := "00"
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) > 2 {
			return 0, errors.New("too many decimals")
		}
		fraction += strings.Repeat("0", 2-len(fraction))
	}
	minor, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, err
	}
	if whole > (math.MaxInt64-minor)/100 {
		return 0, errors.New("amount is too large")
	}
	return whole*100 + minor, nil
}

func (s *Server) refundPayment(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	saleID := r.PathValue("saleID")
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, 400, "request.idempotency_required", "Idempotency-Key is required")
		return
	}
	var input struct {
		AmountMinor int64  `json:"amount_minor"`
		Reason      string `json:"reason"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || !uuidPattern.MatchString(saleID) || input.AmountMinor <= 0 || strings.TrimSpace(input.Reason) == "" {
		writeError(w, 400, "refund.invalid", "Positive amount and reason are required")
		return
	}
	refundID, _ := ids.NewUUIDv7()
	var priorAmount int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT id::text,amount_minor FROM payment_refunds WHERE tenant_id=$1 AND idempotency_key=$2`, scope.TenantID, key).Scan(&refundID, &priorAmount); err == nil {
			if priorAmount != input.AmountMinor {
				return errors.New("idempotency key was used for a different refund")
			}
			return nil
		} else if err != pgx.ErrNoRows {
			return err
		}
		var captured, refunded int64
		var total int64
		if err := tx.QueryRow(r.Context(), `SELECT total_minor FROM sales WHERE tenant_id=$1 AND id=$2 AND status='completed' FOR UPDATE`, scope.TenantID, saleID).Scan(&total); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_tenders WHERE tenant_id=$1 AND sale_id=$2`, scope.TenantID, saleID).Scan(&captured); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payment_refunds WHERE tenant_id=$1 AND sale_id=$2 AND status IN ('pending','completed')`, scope.TenantID, saleID).Scan(&refunded); err != nil {
			return err
		}
		if captured > total || input.AmountMinor > captured-refunded {
			return errors.New("refund exceeds captured balance")
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO payment_refunds(id,tenant_id,sale_id,amount_minor,reason,status,idempotency_key) VALUES($1,$2,$3,$4,$5,'pending',$6)`, refundID, scope.TenantID, saleID, input.AmountMinor, strings.TrimSpace(input.Reason), key)
		return err
	})
	if err != nil {
		writeError(w, 400, "refund.failed", err.Error())
		return
	}
	writeJSON(w, 202, map[string]any{"refund_id": refundID, "status": "pending_approval_or_provider", "amount_minor": input.AmountMinor})
}

func (s *Server) paymentList(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	type row struct {
		ID        string    `json:"id"`
		Amount    int64     `json:"amount_minor"`
		Method    string    `json:"method"`
		Status    string    `json:"status"`
		Reference string    `json:"provider_reference"`
		Created   time.Time `json:"created_at"`
	}
	result := make([]row, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT id::text,amount_minor,method,status,COALESCE(provider_reference,''),created_at FROM payment_intents WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 100`, scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var value row
			if err := rows.Scan(&value.ID, &value.Amount, &value.Method, &value.Status, &value.Reference, &value.Created); err != nil {
				return err
			}
			result = append(result, value)
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "payments.unavailable", "Payments could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"payments": result})
}

func (s *Server) paymentIntentStatus(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	intentID := r.PathValue("intentID")
	if !uuidPattern.MatchString(intentID) {
		writeError(w, 404, "payments.not_found", "Payment intent was not found")
		return
	}
	var result struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount_minor"`
		Reference string `json:"provider_reference"`
		Failure   string `json:"failure_code,omitempty"`
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT id::text,status,amount_minor,COALESCE(provider_reference,''),COALESCE(failure_code,'') FROM payment_intents WHERE tenant_id=$1 AND id=$2`, scope.TenantID, intentID).Scan(&result.ID, &result.Status, &result.Amount, &result.Reference, &result.Failure)
	})
	if err != nil {
		writeError(w, 404, "payments.not_found", "Payment intent was not found")
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) trialBalance(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	type row struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Debit  int64  `json:"debit_minor"`
		Credit int64  `json:"credit_minor"`
	}
	result := make([]row, 0)
	var debit, credit int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT a.code,a.name,COALESCE(SUM(l.debit_minor),0),COALESCE(SUM(l.credit_minor),0) FROM ledger_accounts a LEFT JOIN ledger_lines l ON l.tenant_id=a.tenant_id AND l.account_id=a.id WHERE a.tenant_id=$1 GROUP BY a.id ORDER BY a.code`, scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var value row
			if err := rows.Scan(&value.Code, &value.Name, &value.Debit, &value.Credit); err != nil {
				return err
			}
			debit += value.Debit
			credit += value.Credit
			result = append(result, value)
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, 500, "ledger.unavailable", "Ledger could not be loaded")
		return
	}
	writeJSON(w, 200, map[string]any{"accounts": result, "total_debit_minor": debit, "total_credit_minor": credit, "balanced": debit == credit})
}

func (s *Server) reconcile(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}

	var entries []struct {
		ProviderReference string `json:"provider_reference"`
		AmountMinor       int64  `json:"amount_minor"`
		StatementDate     string `json:"statement_date"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024)).Decode(&entries) != nil {
		writeError(w, 400, "reconciliation.invalid", "A JSON statement is required")
		return
	}
	matched, variances := 0, 0
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		for _, entry := range entries {
			var expected int64
			_ = tx.QueryRow(r.Context(), `SELECT amount_minor FROM payment_intents WHERE tenant_id=$1 AND provider_reference=$2 AND status='captured'`, scope.TenantID, entry.ProviderReference).Scan(&expected)
			variance := entry.AmountMinor - expected
			status := "matched"
			if variance != 0 {
				status = "variance"
				variances++
			} else {
				matched++
			}
			id, _ := ids.NewUUIDv7()
			if _, err := tx.Exec(r.Context(), `INSERT INTO reconciliation_records(id,tenant_id,provider,provider_reference,expected_minor,actual_minor,variance_minor,status,statement_date) VALUES($1,$2,'mpesa',$3,$4,$5,$6,$7,NULLIF($8,'')::date)`, id, scope.TenantID, entry.ProviderReference, expected, entry.AmountMinor, variance, status, entry.StatementDate); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeError(w, 400, "reconciliation.failed", err.Error())
		return
	}
	writeJSON(w, 200, map[string]int{"matched": matched, "variances": variances})
}

func (s *Server) createCreditAccount(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct {
		CustomerName string `json:"customer_name"`
		LimitMinor   int64  `json:"limit_minor"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input) != nil || strings.TrimSpace(input.CustomerName) == "" || input.LimitMinor < 0 {
		writeError(w, 400, "credit.invalid_account", "Customer name and non-negative limit are required")
		return
	}
	accountID, _ := ids.NewUUIDv7()
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO credit_accounts(id,tenant_id,customer_name,limit_minor) VALUES($1,$2,$3,$4)`, accountID, scope.TenantID, strings.TrimSpace(input.CustomerName), input.LimitMinor)
		return err
	})
	if err != nil {
		writeError(w, 400, "credit.create_failed", "Credit account could not be created")
		return
	}
	writeJSON(w, 201, map[string]any{"account_id": accountID, "balance_minor": 0, "limit_minor": input.LimitMinor})
}

func (s *Server) creditStatement(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	accountID := r.PathValue("accountID")
	var account struct {
		ID           string `json:"id"`
		CustomerName string `json:"customer_name"`
		Limit        int64  `json:"limit_minor"`
		Balance      int64  `json:"balance_minor"`
		Status       string `json:"status"`
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT id::text,customer_name,limit_minor,balance_minor,status FROM credit_accounts WHERE tenant_id=$1 AND id=$2`, scope.TenantID, accountID).Scan(&account.ID, &account.CustomerName, &account.Limit, &account.Balance, &account.Status)
	})
	if err != nil {
		writeError(w, 404, "credit.not_found", "Credit account was not found")
		return
	}
	writeJSON(w, 200, account)
}

func ensureLedgerAccounts(r *http.Request, tx pgx.Tx, tenantID string) (string, string, error) {
	accounts := []struct{ code, name, kind string }{{"1100", "Cash and mobile money", "asset"}, {"4000", "Sales revenue", "revenue"}}
	idsByCode := make([]string, 0, 2)
	for _, account := range accounts {
		id, _ := ids.NewUUIDv7()
		if _, err := tx.Exec(r.Context(), `INSERT INTO ledger_accounts(id,tenant_id,code,name,account_type) VALUES($1,$2,$3,$4,$5) ON CONFLICT (tenant_id,code) DO NOTHING`, id, tenantID, account.code, account.name, account.kind); err != nil {
			return "", "", err
		}
		var existing string
		if err := tx.QueryRow(r.Context(), `SELECT id::text FROM ledger_accounts WHERE tenant_id=$1 AND code=$2`, tenantID, account.code).Scan(&existing); err != nil {
			return "", "", err
		}
		idsByCode = append(idsByCode, existing)
	}
	return idsByCode[0], idsByCode[1], nil
}

func postPaymentJournal(r *http.Request, tx pgx.Tx, tenantID, intentID, saleID string, amount int64, reference string) error {
	cashAccount, revenueAccount, err := ensureLedgerAccounts(r, tx, tenantID)
	if err != nil {
		return err
	}
	clearingID, _ := ids.NewUUIDv7()
	if _, err = tx.Exec(r.Context(), `INSERT INTO ledger_accounts(id,tenant_id,code,name,account_type) VALUES($1,$2,'1120','M-Pesa clearing','asset') ON CONFLICT(tenant_id,code) DO NOTHING`, clearingID, tenantID); err != nil {
		return err
	}
	if err = tx.QueryRow(r.Context(), `SELECT id::text FROM ledger_accounts WHERE tenant_id=$1 AND code='1120'`, tenantID).Scan(&clearingID); err != nil {
		return err
	}
	journalID, _ := ids.NewUUIDv7()
	if _, err = tx.Exec(r.Context(), `INSERT INTO ledger_journals(id,tenant_id,reference_type,reference_id,description) VALUES($1,$2,'payment_capture',$3,'M-Pesa payment capture') ON CONFLICT(tenant_id,reference_type,reference_id) DO NOTHING`, journalID, tenantID, intentID); err != nil {
		return err
	}
	if err = tx.QueryRow(r.Context(), `SELECT id::text FROM ledger_journals WHERE tenant_id=$1 AND reference_type='payment_capture' AND reference_id=$2`, tenantID, intentID).Scan(&journalID); err != nil {
		return err
	}
	var lineCount int
	if err = tx.QueryRow(r.Context(), `SELECT count(*) FROM ledger_lines WHERE tenant_id=$1 AND journal_id=$2`, tenantID, journalID).Scan(&lineCount); err != nil {
		return err
	}
	if lineCount > 0 {
		return nil
	}
	for _, line := range []struct {
		account       string
		debit, credit int64
	}{{clearingID, amount, 0}, {revenueAccount, 0, amount}} {
		id, _ := ids.NewUUIDv7()
		if _, err = tx.Exec(r.Context(), `INSERT INTO ledger_lines(id,tenant_id,journal_id,account_id,debit_minor,credit_minor) VALUES($1,$2,$3,$4,$5,$6)`, id, tenantID, journalID, line.account, line.debit, line.credit); err != nil {
			return err
		}
	}
	_ = cashAccount
	_ = saleID
	_ = reference
	return nil
}

func normalizePhoneForComparison(phone string) string {
	digits := strings.NewReplacer("+", "", " ", "", "-", "").Replace(phone)
	if strings.HasPrefix(digits, "0") && len(digits) == 10 {
		return "254" + digits[1:]
	}
	if strings.HasPrefix(digits, "7") && len(digits) == 9 {
		return "254" + digits
	}
	if strings.HasPrefix(digits, "1") && len(digits) == 9 {
		return "254" + digits
	}
	return digits
}

func validPhone(phone string) bool {
	if len(phone) < 9 || len(phone) > 15 {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
