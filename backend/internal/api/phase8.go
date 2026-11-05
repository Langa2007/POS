package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) consolidatedReport(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var rowsOut []map[string]any
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT b.id::text, b.name,
			       COALESCE(SUM(CASE WHEN s.status='completed' THEN s.total_minor ELSE 0 END),0),
			       COUNT(s.id)
			FROM branches b
			LEFT JOIN sales s ON s.tenant_id=b.tenant_id AND s.branch_id=b.id
			WHERE b.tenant_id=$1 AND ($2='' OR b.id::text=$2)
			GROUP BY b.id,b.name ORDER BY b.name`, scope.TenantID, scope.BranchID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, name string
			var total, count int64
			if err := rows.Scan(&id, &name, &total, &count); err != nil {
				return err
			}
			rowsOut = append(rowsOut, map[string]any{"branch_id": id, "branch_name": name, "sales_total_minor": total, "sales_count": count})
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"branches": rowsOut})
}

func (s *Server) createPayout(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct {
		BeneficiaryType string `json:"beneficiary_type"`
		BeneficiaryID   string `json:"beneficiary_id"`
		Provider        string `json:"provider"`
		AmountMinor     int64  `json:"amount_minor"`
		IdempotencyKey  string `json:"idempotency_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.AmountMinor <= 0 || input.Provider == "" || input.IdempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "request.invalid", "provider, amount_minor and idempotency_key are required")
		return
	}
	id, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal.error", "Could not create payout")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `
			INSERT INTO payouts(id,tenant_id,beneficiary_type,beneficiary_id,provider,amount_minor,idempotency_key,requested_by)
			VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,$8)`,
			id, scope.TenantID, input.BeneficiaryType, input.BeneficiaryID, input.Provider, input.AmountMinor, input.IdempotencyKey, scope.UserID)
		return err
	})
	if err != nil {
		writeError(w, http.StatusConflict, "payout.duplicate_or_invalid", "Payout could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": "pending"})
}

func (s *Server) approvePayout(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	payoutID := r.PathValue("payoutID")
	var approved bool
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(), `
			UPDATE payouts SET status='approved', approved_by=$1, approved_at=now()
			WHERE tenant_id=$2 AND id=$3 AND status='pending' AND requested_by<>$1`, scope.UserID, scope.TenantID, payoutID)
		approved = tag.RowsAffected() == 1
		return err
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal.error", "Payout approval failed")
		return
	}
	if !approved {
		writeError(w, http.StatusConflict, "payout.approval_denied", "Payout is missing, already processed, or requires another approver")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": payoutID, "status": "approved", "approved_at": time.Now().UTC()})
}
