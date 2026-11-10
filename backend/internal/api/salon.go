package api

import (
	"net/http"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listSalonAppointments(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}

	date := r.URL.Query().Get("date")
	appointments := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT id::text, service_type, staff_id::text, starts_at, ends_at, status
			FROM service_appointments
			WHERE tenant_id = $1 AND ($2 = '' OR starts_at::date = NULLIF($2, '')::date)
			ORDER BY starts_at`, scope.TenantID, date)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, serviceType, staffID, status string
			var startsAt, endsAt any
			if err := rows.Scan(&id, &serviceType, &staffID, &startsAt, &endsAt, &status); err != nil {
				return err
			}
			appointments = append(appointments, map[string]any{
				"id": id, "service_type": serviceType, "staff_id": staffID,
				"starts_at": startsAt, "ends_at": endsAt, "status": status,
			})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list salon appointments", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "internal.error", "Appointments could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"appointments": appointments})
}

func (s *Server) updateSalonAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	switch status {
	case "booked", "completed", "cancelled", "no_show":
	default:
		writeError(w, http.StatusBadRequest, "appointment.status_invalid", "Unsupported appointment status")
		return
	}

	tag, err := s.db.Pool().Exec(r.Context(), `
		UPDATE service_appointments
		SET status = $1
		WHERE tenant_id = $2 AND id = $3::uuid`, status, scope.TenantID, r.PathValue("appointmentID"))
	if err != nil {
		s.logger.Error("update salon appointment", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusBadRequest, "appointment.invalid", "Appointment could not be updated")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "appointment.not_found", "Appointment not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": r.PathValue("appointmentID"), "status": status})
}
