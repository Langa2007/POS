package api

import (
	"net/http"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listSalonPackages(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	packages := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT id::text, name, description, price_minor, visits, duration_minutes, active
			FROM salon_service_packages
			WHERE tenant_id = $1 AND active
			ORDER BY name`, scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, name, description string
			var price int64
			var visits, duration int
			var active bool
			if err := rows.Scan(&id, &name, &description, &price, &visits, &duration, &active); err != nil {
				return err
			}
			packages = append(packages, map[string]any{"id": id, "name": name, "description": description, "price_minor": price, "visits": visits, "duration_minutes": duration, "active": active})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list salon packages", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "packages.unavailable", "Service packages could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"packages": packages})
}

func (s *Server) listSalonStaff(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	staff := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT id::text, email FROM users WHERE tenant_id=$1 AND active ORDER BY email`, scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, email string
			if err := rows.Scan(&id, &email); err != nil {
				return err
			}
			staff = append(staff, map[string]any{"id": id, "name": email})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list salon staff", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "staff.unavailable", "Salon staff could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"staff": staff})
}

func (s *Server) salonDashboard(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var bookings, inService int
	var revenue, commission int64
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		today := time.Now().UTC().Format("2006-01-02")
		if err := tx.QueryRow(r.Context(), `
			SELECT count(*) FILTER (WHERE status IN ('booked','completed','cancelled','no_show')),
			       count(*) FILTER (WHERE status='booked')
			FROM service_appointments WHERE tenant_id=$1 AND starts_at::date=$2::date`,
			scope.TenantID, today).Scan(&bookings, &inService); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `
			SELECT COALESCE(sum(total_minor),0) FROM sales
			WHERE tenant_id=$1 AND status='completed' AND created_at::date=$2::date`,
			scope.TenantID, today).Scan(&revenue); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `
			SELECT COALESCE(sum(commission_minor),0) FROM salon_commission_periods
			WHERE tenant_id=$1 AND period_start <= $2::date AND period_end >= $2::date`,
			scope.TenantID, today).Scan(&commission)
	})
	if err != nil {
		s.logger.Error("load salon dashboard", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "dashboard.unavailable", "Salon dashboard could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bookings": bookings, "in_service": inService, "revenue_minor": revenue, "commission_minor": commission})
}

func (s *Server) listSalonCommissions(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	commissions := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT p.staff_id::text, u.email, p.service_count, p.sales_minor, p.commission_minor
			FROM salon_commission_periods p
			JOIN users u ON u.tenant_id=p.tenant_id AND u.id=p.staff_id
			WHERE p.tenant_id=$1
			ORDER BY u.email`, scope.TenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var staffID, name string
			var services int
			var sales, commission int64
			if err := rows.Scan(&staffID, &name, &services, &sales, &commission); err != nil {
				return err
			}
			commissions = append(commissions, map[string]any{"staff_id": staffID, "name": name, "services": services, "sales_minor": sales, "commission_minor": commission})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list salon commissions", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "commissions.unavailable", "Commissions could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"commissions": commissions})
}
