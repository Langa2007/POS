package api

import (
	"net/http"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Server) requirePermission(permission string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope, err := tenancy.Require(r.Context())
		if err != nil {
			writeError(w, http.StatusUnauthorized, "auth.required", "A valid bearer token is required")
			return
		}
		allowed := false
		err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
			rows, err := tx.Query(r.Context(), `SELECT rp.permission_id FROM user_roles ur JOIN role_permissions rp ON rp.tenant_id=ur.tenant_id AND rp.role_id=ur.role_id WHERE ur.tenant_id=$1 AND ur.user_id=$2`, scope.TenantID, scope.UserID)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var granted string
				if err := rows.Scan(&granted); err != nil {
					return err
				}
				if permissionMatches(granted, permission) {
					allowed = true
				}
			}
			return rows.Err()
		})
		if err != nil {
			s.logger.Error("check permission", "tenant_id", scope.TenantID, "user_id", scope.UserID, "permission", permission, "error", err)
			writeError(w, 500, "internal.error", "Request could not be completed")
			return
		}
		if !allowed {
			writeError(w, http.StatusForbidden, "auth.permission_denied", "You do not have permission to perform this action")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func permissionMatches(granted, required string) bool {
	if granted == "*" || granted == required {
		return true
	}
	if strings.HasSuffix(granted, ".*") {
		return strings.HasPrefix(required, strings.TrimSuffix(granted, "*"))
	}
	return false
}
