package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/verticals/fuelstation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var verticalResources = map[string]map[string][]string{
	"fuelstation": {
		"pumps":      {"pump", "grade", "product"},
		"shifts":     {"pump", "operator", "open", "close", "sales"},
		"dips":       {"tank", "product", "reading", "volume", "operator"},
		"deliveries": {"supplier", "note", "product", "tank", "volume", "operator"},
	},
	"garage": {
		"vehicles":     {"registration", "vehicle", "owner"},
		"jobs":         {"customer", "registration", "vehicle", "issue"},
		"appointments": {"customer", "vehicle", "service", "start", "end"},
		"reminders":    {"vehicle", "customer", "service"},
	},
}

func (s *Server) listVerticalOperationRecords(w http.ResponseWriter, r *http.Request) {
	vertical, resource, ok := verticalResourcePath(r)
	if !ok {
		writeError(w, http.StatusNotFound, "vertical.resource_not_found", "Unsupported vertical resource")
		return
	}
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	items := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT id::text, data, created_at, updated_at
			FROM vertical_operation_records
			WHERE tenant_id=$1 AND vertical_id=$2 AND resource=$3
			ORDER BY created_at DESC`, scope.TenantID, vertical, resource)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			var raw []byte
			var createdAt, updatedAt any
			if err := rows.Scan(&id, &raw, &createdAt, &updatedAt); err != nil {
				return err
			}
			item := map[string]any{}
			if err := json.Unmarshal(raw, &item); err != nil {
				return err
			}
			item["id"], item["created_at"], item["updated_at"] = id, createdAt, updatedAt
			items = append(items, item)
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list vertical records", "tenant_id", scope.TenantID, "vertical", vertical, "resource", resource, "error", err)
		writeError(w, http.StatusInternalServerError, "vertical.records_unavailable", "Records could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{resource: items})
}

func (s *Server) createVerticalOperationRecord(w http.ResponseWriter, r *http.Request) {
	vertical, resource, ok := verticalResourcePath(r)
	if !ok {
		writeError(w, http.StatusNotFound, "vertical.resource_not_found", "Unsupported vertical resource")
		return
	}
	var input map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&input); err != nil || input == nil {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", "Provide a valid record")
		return
	}
	if err := validateVerticalRecord(vertical, resource, input, false); err != nil {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", err.Error())
		return
	}
	if vertical == "fuelstation" && resource == "shifts" {
		if err := reconcileMeterRecord(input); err != nil {
			writeError(w, http.StatusBadRequest, "fuelstation.shift_invalid", err.Error())
			return
		}
	}
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	id, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "id.generate_failed", "Record could not be created")
		return
	}
	delete(input, "id")
	delete(input, "created_at")
	delete(input, "updated_at")
	data, err := json.Marshal(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", "Record could not be encoded")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO vertical_operation_records(id,tenant_id,vertical_id,resource,data) VALUES($1,$2,$3,$4,$5::jsonb)`, id, scope.TenantID, vertical, resource, string(data))
		return err
	})
	if err != nil {
		s.logger.Error("create vertical record", "tenant_id", scope.TenantID, "vertical", vertical, "resource", resource, "error", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23P01") {
			writeError(w, http.StatusConflict, "vertical.record_conflict", "Record conflicts with an existing registration or appointment")
			return
		}
		writeError(w, http.StatusInternalServerError, "vertical.record_create_failed", "Record could not be saved")
		return
	}
	input["id"] = id
	writeJSON(w, http.StatusCreated, input)
}

func (s *Server) updateVerticalOperationRecord(w http.ResponseWriter, r *http.Request) {
	vertical, resource, ok := verticalResourcePath(r)
	if !ok || !uuidPattern.MatchString(r.PathValue("recordID")) {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", "Unsupported resource or record ID")
		return
	}
	var patch map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&patch); err != nil || patch == nil {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", "Provide a valid record update")
		return
	}
	if err := validateVerticalRecord(vertical, resource, patch, true); err != nil {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", err.Error())
		return
	}
	delete(patch, "id")
	delete(patch, "created_at")
	delete(patch, "updated_at")
	data, _ := json.Marshal(patch)
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var raw []byte
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `UPDATE vertical_operation_records SET data=data || $1::jsonb, updated_at=now() WHERE tenant_id=$2 AND vertical_id=$3 AND resource=$4 AND id=$5 RETURNING data`, string(data), scope.TenantID, vertical, resource, r.PathValue("recordID")).Scan(&raw)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "vertical.record_not_found", "Record was not found")
		return
	}
	if err != nil {
		s.logger.Error("update vertical record", "tenant_id", scope.TenantID, "vertical", vertical, "resource", resource, "error", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			writeError(w, http.StatusConflict, "appointment.technician_conflict", "Technician already has an appointment during this time")
			return
		}
		writeError(w, http.StatusInternalServerError, "vertical.record_update_failed", "Record could not be updated")
		return
	}
	result := map[string]any{}
	_ = json.Unmarshal(raw, &result)
	result["id"] = r.PathValue("recordID")
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) deleteVerticalOperationRecord(w http.ResponseWriter, r *http.Request) {
	vertical, resource, ok := verticalResourcePath(r)
	if !ok || !uuidPattern.MatchString(r.PathValue("recordID")) {
		writeError(w, http.StatusBadRequest, "vertical.record_invalid", "Unsupported resource or record ID")
		return
	}
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var deleted bool
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(), `DELETE FROM vertical_operation_records WHERE tenant_id=$1 AND vertical_id=$2 AND resource=$3 AND id=$4`, scope.TenantID, vertical, resource, r.PathValue("recordID"))
		deleted = tag.RowsAffected() > 0
		return err
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vertical.record_delete_failed", "Record could not be deleted")
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "vertical.record_not_found", "Record was not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func verticalResourcePath(r *http.Request) (string, string, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		return "", "", false
	}
	vertical, resource := parts[2], parts[3]
	resources, ok := verticalResources[vertical]
	if !ok {
		return "", "", false
	}
	_, ok = resources[resource]
	return vertical, resource, ok
}

func validateVerticalRecord(vertical, resource string, input map[string]any, partial bool) error {
	required := verticalResources[vertical][resource]
	for _, key := range required {
		value, exists := input[key]
		if !exists {
			if partial {
				continue
			}
			return fmt.Errorf("%s is required", key)
		}
		if text, isText := value.(string); isText && strings.TrimSpace(text) == "" {
			return fmt.Errorf("%s is required", key)
		}
	}
	for _, key := range []string{"open", "close", "deliveries", "sales", "tolerance", "reading", "volume", "odometer", "parts", "labor", "dueKm"} {
		if value, exists := input[key]; exists {
			if number, ok := numericValue(value); ok && number < 0 {
				return fmt.Errorf("%s cannot be negative", key)
			}
		}
	}
	if resource == "shifts" {
		opening, okOpen := numericValue(input["open"])
		closing, okClose := numericValue(input["close"])
		if okOpen && okClose && closing < opening {
			return errors.New("closing meter cannot be lower than opening meter")
		}
	}
	if resource == "appointments" {
		if start, ok := input["start"].(string); ok {
			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				return errors.New("appointment start must be an RFC3339 timestamp")
			}
			if end, okEnd := input["end"].(string); okEnd {
				endTime, err := time.Parse(time.RFC3339, end)
				if err != nil || !endTime.After(startTime) {
					return errors.New("appointment end must be a valid timestamp after start")
				}
			}
		}
	}
	if resource == "jobs" {
		if status, ok := input["status"].(string); ok {
			switch status {
			case "Open", "Diagnosing", "Awaiting parts", "In repair", "Quality check", "Ready", "Collected":
			default:
				return errors.New("unsupported job status")
			}
		}
	}
	if resource == "appointments" {
		if status, ok := input["status"].(string); ok {
			switch status {
			case "Booked", "Arrived", "In service", "Completed", "Cancelled", "No show":
			default:
				return errors.New("unsupported appointment status")
			}
		}
	}
	if resource == "reminders" {
		if status, ok := input["status"].(string); ok {
			switch status {
			case "Due", "Contacted", "Booked", "Completed":
			default:
				return errors.New("unsupported reminder status")
			}
		}
	}
	return nil
}

func numericValue(value any) (float64, bool) {
	number, ok := value.(float64)
	return number, ok
}

func reconcileMeterRecord(input map[string]any) error {
	opening, okOpening := numericValue(input["open"])
	closing, okClosing := numericValue(input["close"])
	deliveries, okDeliveries := numericValue(input["deliveries"])
	sales, okSales := numericValue(input["sales"])
	tolerance, okTolerance := numericValue(input["tolerance"])
	if !okOpening || !okClosing || !okDeliveries || !okSales {
		return errors.New("meter readings, deliveries, and sales must be numeric")
	}
	if !okTolerance {
		tolerance = 0
	}
	volume := closing - opening + deliveries
	input["volume"] = volume
	input["variance"] = volume - sales
	input["ok"] = fuelstation.ReconcileMeter(opening, closing, deliveries, sales, tolerance)
	return nil
}
