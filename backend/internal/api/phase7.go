package api

import (
	"encoding/json"
	"net/http"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/verticals"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listVerticals(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	type result struct {
		ID           string          `json:"id"`
		Version      string          `json:"version"`
		Enabled      bool            `json:"enabled"`
		Capabilities []string        `json:"capabilities"`
		Manifest     json.RawMessage `json:"manifest"`
	}
	enabled := map[string]bool{}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, e := tx.Query(r.Context(), `SELECT module_id, enabled FROM tenant_modules WHERE tenant_id=$1`, scope.TenantID)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			var value bool
			if e = rows.Scan(&id, &value); e != nil {
				return e
			}
			enabled[id] = value
		}
		return rows.Err()
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal.error", "Request could not be completed")
		return
	}
	out := make([]result, 0)
	for _, m := range verticals.All() {
		out = append(out, result{ID: m.ID(), Version: m.Version(), Enabled: enabled[m.ID()], Capabilities: m.Capabilities(), Manifest: json.RawMessage(m.Manifest())})
	}
	writeJSON(w, http.StatusOK, map[string]any{"verticals": out})
}

func (s *Server) setVerticalEnabled(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	m, ok := verticals.ByID(r.PathValue("verticalID"))
	if !ok {
		writeError(w, http.StatusNotFound, "vertical.not_found", "Unknown vertical")
		return
	}
	var request struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "request.invalid", "Invalid payload")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `
			INSERT INTO tenant_modules(tenant_id,module_id,enabled,enabled_at)
			VALUES($1,$2,$3,now())
			ON CONFLICT (tenant_id,module_id) DO UPDATE SET enabled=EXCLUDED.enabled, enabled_at=now()`,
			scope.TenantID, m.ID(), request.Enabled)
		if err != nil {
			return err
		}
		for _, capability := range m.Capabilities() {
			if _, err = tx.Exec(r.Context(), `
				INSERT INTO tenant_capabilities(tenant_id,capability_id,enabled)
				VALUES($1,$2,$3)
				ON CONFLICT (tenant_id,capability_id) DO UPDATE SET enabled=EXCLUDED.enabled`,
				scope.TenantID, capability, request.Enabled); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": m.ID(), "enabled": request.Enabled})
}
