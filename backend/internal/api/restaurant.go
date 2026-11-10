package api

import (
	"net/http"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) restaurantMenu(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	s.catalog(w, r)
}

func (s *Server) restaurantKitchenTickets(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	// Kitchen tickets are populated by the order workflow when kitchen routing is enabled.
	writeJSON(w, http.StatusOK, map[string]any{"tickets": []map[string]any{}})
}

func (s *Server) restaurantTables(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tables": []map[string]any{}})
}

func (s *Server) updateRestaurantTable(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	writeError(w, http.StatusNotImplemented, "restaurant.tables_unavailable", "Table updates are not available. Please contact your administrator")
}

func (s *Server) restaurantSummary(w http.ResponseWriter, r *http.Request) {
	if _, err := tenancy.Require(r.Context()); err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "Please sign in again or contact your administrator")
		return
	}
	var report struct {
		Sales int64 `json:"sales_count"`
		Total int64 `json:"total_minor"`
	}
	scope, _ := tenancy.Require(r.Context())
	err := s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT COUNT(*), COALESCE(SUM(total_minor), 0) FROM sales WHERE tenant_id=$1 AND status='completed' AND created_at::date=current_date`, scope.TenantID).Scan(&report.Sales, &report.Total)
	})
	if err != nil {
		s.logger.Error("load restaurant summary", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "restaurant.summary_unavailable", "Restaurant summary is unavailable. Please contact your administrator")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sales_count": report.Sales, "total_minor": report.Total, "tables": 0, "covers": 0, "kitchen_tickets": 0})
}
