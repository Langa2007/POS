package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/jackc/pgx/v5"
)

type registerTerminalRequest struct {
	BranchID string `json:"branch_id"`
	Name     string `json:"name"`
}
type pinLoginRequest struct {
	TenantID      string `json:"tenant_id"`
	TerminalID    string `json:"terminal_id"`
	TerminalToken string `json:"terminal_token"`
	UserID        string `json:"user_id"`
	PIN           string `json:"pin"`
}
type setPINRequest struct {
	PIN string `json:"pin"`
}

func (s *Server) registerTerminal(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	var input registerTerminalRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil || !uuidPattern.MatchString(input.BranchID) || strings.TrimSpace(input.Name) == "" || len(input.Name) > 120 {
		writeError(w, 400, "request.invalid", "Provide a valid branch_id and terminal name")
		return
	}
	terminalID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	auditID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	credential, err := auth.NewRefreshToken(scope.TenantID)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(), `INSERT INTO terminals(id,tenant_id,branch_id,name,token_hash) VALUES($1,$2,$3,$4,$5)`, terminalID, scope.TenantID, input.BranchID, strings.TrimSpace(input.Name), auth.HashRefreshToken(credential)); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'terminal.registered','terminal',$4,jsonb_build_object('branch_id',$5::text,'name',$6::text))`, auditID, scope.TenantID, scope.UserID, terminalID, input.BranchID, strings.TrimSpace(input.Name))
		return err
	})
	if err != nil {
		s.logger.Error("register terminal", "tenant_id", scope.TenantID, "branch_id", input.BranchID, "error", err)
		writeError(w, 400, "terminal.registration_failed", "Terminal could not be registered for that branch")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"terminal_id": terminalID, "terminal_token": credential, "branch_id": input.BranchID})
}

func (s *Server) setUserPIN(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	userID := r.PathValue("userID")
	if !uuidPattern.MatchString(userID) {
		writeError(w, 400, "request.invalid", "A valid user_id is required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var input setPINRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "request.invalid", "Provide a six-digit PIN")
		return
	}
	pinHash, err := auth.HashPIN(input.PIN)
	if err != nil {
		writeError(w, 400, "request.invalid", "PIN must contain exactly six digits")
		return
	}
	auditID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(), `UPDATE users SET pin_hash=$1,pin_failed_count=0,pin_locked_until=NULL WHERE tenant_id=$2 AND id=$3`, pinHash, scope.TenantID, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return pgx.ErrNoRows
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO audit_log(id,tenant_id,actor_id,action,entity_type,entity_id) VALUES($1,$2,$3,'user.pin_changed','user',$4)`, auditID, scope.TenantID, scope.UserID, userID)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 404, "user.not_found", "User was not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal.error", "PIN could not be updated")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) pinLogin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	var input pinLoginRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil || !uuidPattern.MatchString(input.TenantID) || !uuidPattern.MatchString(input.TerminalID) || !uuidPattern.MatchString(input.UserID) || len(input.PIN) != 6 {
		writeError(w, 400, "request.invalid", "Provide tenant, terminal, user, and six-digit PIN details")
		return
	}
	parts := strings.SplitN(input.TerminalToken, ".", 2)
	if len(parts) != 2 || parts[0] != input.TenantID || len(parts[1]) < 40 {
		writeError(w, 401, "auth.invalid_terminal", "Terminal credentials are invalid")
		return
	}
	refreshToken, err := auth.NewRefreshToken(input.TenantID)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	refreshID, err := ids.NewUUIDv7()
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	var branchID, userID string
	var authFailure bool
	err = s.db.TenantTx(r.Context(), input.TenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT branch_id::text FROM terminals WHERE tenant_id=$1 AND id=$2 AND token_hash=$3 AND revoked_at IS NULL`, input.TenantID, input.TerminalID, auth.HashRefreshToken(input.TerminalToken)).Scan(&branchID); err != nil {
			return err
		}
		var pinHash sql.NullString
		var lockedUntil sql.NullTime
		var branchAllowed bool
		err := tx.QueryRow(r.Context(), `SELECT pin_hash,pin_locked_until,EXISTS(SELECT 1 FROM user_branches WHERE tenant_id=u.tenant_id AND user_id=u.id AND branch_id=$3) FROM users u WHERE tenant_id=$1 AND id=$2 AND active FOR UPDATE OF u`, input.TenantID, input.UserID, branchID).Scan(&pinHash, &lockedUntil, &branchAllowed)
		if err != nil {
			return err
		}
		userID = input.UserID
		if !branchAllowed || !pinHash.Valid || (lockedUntil.Valid && lockedUntil.Time.After(time.Now())) {
			authFailure = true
			return nil
		}
		if !auth.VerifyPassword(pinHash.String, input.PIN) {
			authFailure = true
			_, err = tx.Exec(r.Context(), `UPDATE users SET pin_failed_count=pin_failed_count+1,pin_locked_until=CASE WHEN pin_failed_count+1>=5 THEN now()+interval '15 minutes' ELSE pin_locked_until END WHERE tenant_id=$1 AND id=$2`, input.TenantID, input.UserID)
			return err
		}
		if _, err := tx.Exec(r.Context(), `UPDATE users SET pin_failed_count=0,pin_locked_until=NULL WHERE tenant_id=$1 AND id=$2`, input.TenantID, input.UserID); err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO refresh_tokens(id,tenant_id,user_id,branch_id,terminal_id,token_hash,expires_at) VALUES($1,$2,$3,$4,$5,$6,now()+interval '30 days')`, refreshID, input.TenantID, input.UserID, branchID, input.TerminalID, auth.HashRefreshToken(refreshToken))
		return err
	})
	if err != nil || authFailure {
		writeError(w, 401, "auth.invalid_credentials", "Terminal, user, or PIN is invalid")
		return
	}
	access, err := s.tokens.IssueForScope(input.TenantID, userID, branchID, input.TerminalID)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	writeJSON(w, 200, map[string]any{"access_token": access, "refresh_token": refreshToken, "token_type": "Bearer", "expires_in": 900, "user_id": userID, "tenant_id": input.TenantID, "branch_id": branchID, "terminal_id": input.TerminalID})
}
