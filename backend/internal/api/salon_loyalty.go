package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) listSalonLoyaltyMembers(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	members := make([]map[string]any, 0)
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT id::text, name, phone, membership_no, tier, points, last_visit
			FROM salon_loyalty_members
			WHERE tenant_id = $1
			  AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR phone ILIKE '%' || $2 || '%' OR membership_no ILIKE '%' || $2 || '%')
			ORDER BY name`, scope.TenantID, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, name, phone, membershipNo, tier string
			var points int
			var lastVisit any
			if err := rows.Scan(&id, &name, &phone, &membershipNo, &tier, &points, &lastVisit); err != nil {
				return err
			}
			members = append(members, map[string]any{
				"id": id, "name": name, "phone": phone, "membership_no": membershipNo,
				"tier": tier, "points": points, "last_visit": lastVisit,
			})
		}
		return rows.Err()
	})
	if err != nil {
		s.logger.Error("list salon loyalty members", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "loyalty.unavailable", "Loyalty members could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (s *Server) createSalonLoyaltyMember(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&input) != nil ||
		strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Phone) == "" {
		writeError(w, http.StatusBadRequest, "loyalty.invalid_member", "Name and phone are required")
		return
	}
	id, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "id.generate_failed", "Member could not be created")
		return
	}
	membershipNo := "SPA-" + strings.ToUpper(id[:8])
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `
			INSERT INTO salon_loyalty_members(id, tenant_id, name, phone, membership_no)
			VALUES ($1, $2, $3, $4, $5)`,
			id, scope.TenantID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Phone), membershipNo)
		return err
	})
	if err != nil {
		s.logger.Error("create salon loyalty member", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusBadRequest, "loyalty.member_exists", "Member could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "name": input.Name, "phone": input.Phone, "membership_no": membershipNo, "tier": "Bronze", "points": 0})
}

func (s *Server) redeemSalonLoyalty(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
		return
	}
	var input struct{ Points int `json:"points"` }
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4*1024)).Decode(&input) != nil || input.Points <= 0 {
		writeError(w, http.StatusBadRequest, "loyalty.invalid_points", "Points must be greater than zero")
		return
	}
	tag, err := s.db.Pool().Exec(r.Context(), `
		UPDATE salon_loyalty_members
		SET points = points - $1
		WHERE tenant_id = $2 AND id = $3::uuid AND points >= $1`,
		input.Points, scope.TenantID, r.PathValue("memberID"))
	if err != nil {
		s.logger.Error("redeem salon loyalty", "tenant_id", scope.TenantID, "error", err)
		writeError(w, http.StatusBadRequest, "loyalty.invalid_member", "Reward could not be applied")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusConflict, "loyalty.insufficient_points", "Member not found or has insufficient points")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": r.PathValue("memberID"), "redeemed_points": input.Points})
}
