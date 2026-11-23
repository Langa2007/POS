package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) membershipDashboard(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}

	counts := map[string]int{}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		for _, resource := range []string{"plans", "members", "attendance", "billing", "freezes", "renewals", "access_logs"} {
			var count int
			if err := tx.QueryRow(r.Context(), `
				SELECT count(*)
				FROM vertical_operation_records
				WHERE tenant_id=$1 AND vertical_id='membership' AND resource=$2`,
				scope.TenantID, resource).Scan(&count); err != nil {
				return err
			}
			counts[resource] = count
		}
		var active int
		if err := tx.QueryRow(r.Context(), `
			SELECT count(*)
			FROM vertical_operation_records
			WHERE tenant_id=$1 AND vertical_id='membership' AND resource='members'
			  AND lower(COALESCE(data->>'status','active')) IN ('active','current')`,
			scope.TenantID).Scan(&active); err != nil {
			return err
		}
		counts["active_members"] = active
		return nil
	})
	if err != nil {
		s.logger.Error("load membership dashboard", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "membership.dashboard_unavailable", "Dashboard could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, counts)
}

func (s *Server) runMembershipBilling(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}

	var request struct {
		Period string `json:"period"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "membership.billing_invalid", "Provide a valid billing period")
		return
	}
	request.Period = request.Period
	if request.Period == "" {
		request.Period = time.Now().UTC().Format("2006-01")
	}

	result := map[string]any{"period": request.Period, "processed": 0, "duplicate": false}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		var duplicate bool
		if err := tx.QueryRow(r.Context(), `
			SELECT EXISTS (
				SELECT 1 FROM vertical_operation_records
				WHERE tenant_id=$1 AND vertical_id='membership' AND resource='billing'
				  AND data->>'run_key'=$2
			)`, scope.TenantID, request.Period).Scan(&duplicate); err != nil {
			return err
		}
		if duplicate {
			result["duplicate"] = true
			return nil
		}

		rows, err := tx.Query(r.Context(), `
			SELECT data
			FROM vertical_operation_records
			WHERE tenant_id=$1 AND vertical_id='membership' AND resource='members'
			  AND lower(COALESCE(data->>'status','active')) IN ('active','current')`,
			scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		processed := 0
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				return err
			}
			member := map[string]any{}
			if err := json.Unmarshal(raw, &member); err != nil {
				return err
			}
			memberName, _ := member["name"].(string)
			planName, _ := member["plan"].(string)
			amount := 0

			var planRaw []byte
			err := tx.QueryRow(r.Context(), `
				SELECT data
				FROM vertical_operation_records
				WHERE tenant_id=$1 AND vertical_id='membership' AND resource='plans'
				  AND data->>'name'=$2
				ORDER BY created_at DESC LIMIT 1`,
				scope.TenantID, planName).Scan(&planRaw)
			if err == nil {
				plan := map[string]any{}
				if json.Unmarshal(planRaw, &plan) == nil {
					switch value := plan["price"].(type) {
					case float64:
						amount = int(value)
					}
				}
			}

			record := map[string]any{
				"member":  memberName,
				"amount":  amount,
				"period":  request.Period,
				"status":  "Pending",
				"run_key": request.Period,
			}
			data, err := json.Marshal(record)
			if err != nil {
				return err
			}
			id, err := ids.NewUUIDv7()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(r.Context(), `
				INSERT INTO vertical_operation_records(id, tenant_id, vertical_id, resource, data)
				VALUES($1,$2,'membership','billing',$3::jsonb)`,
				id, scope.TenantID, string(data)); err != nil {
				return err
			}
			processed++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		result["processed"] = processed
		return nil
	})
	if err != nil {
		s.logger.Error("run membership billing", "tenant_id", scope.TenantID, "period", request.Period, "error", err)
		writeError(w, http.StatusInternalServerError, "membership.billing_failed", "Billing run could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
