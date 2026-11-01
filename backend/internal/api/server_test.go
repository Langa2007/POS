package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/Langa2007/POS/backend/internal/platform/auth"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

func TestHealthAndProtectedManifest(t *testing.T) {
	tokens, err := auth.NewTokenManager([]byte(strings.Repeat("k", 32)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	srv := New(&db.Store{}, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	live := httptest.NewRecorder()
	srv.Handler().ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("health status=%d", live.Code)
	}
	unauthorized := httptest.NewRecorder()
	srv.Handler().ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/me/manifest", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized manifest status=%d", unauthorized.Code)
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to exercise authenticated manifest against PostgreSQL")
	}
	ctx := context.Background()
	store, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	tenantID, _ := ids.NewUUIDv7()
	userID, _ := ids.NewUUIDv7()
	roleID, _ := ids.NewUUIDv7()
	passwordHash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,'manifest test')`, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO users(id,tenant_id,email,password_hash) VALUES($1,$2,'owner@example.test',$3)`, userID, tenantID, passwordHash); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO roles(id,tenant_id,name) VALUES($1,$2,'owner')`, roleID, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO permissions(id,description) VALUES('core.*','Core permissions') ON CONFLICT DO NOTHING`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO role_permissions(tenant_id,role_id,permission_id) VALUES($1,$2,'core.*')`, tenantID, roleID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles(tenant_id,user_id,role_id) VALUES($1,$2,$3)`, tenantID, userID, roleID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tenant_modules(tenant_id,module_id) VALUES($1,'core.test')`, tenantID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO tenant_capabilities(tenant_id,capability_id) VALUES($1,'test.capability')`, tenantID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	srv = New(store, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	token, _ := tokens.Issue(tenantID, userID)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/manifest", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("manifest status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["tenant_id"] != tenantID {
		t.Fatalf("manifest tenant=%v", payload["tenant_id"])
	}
	deniedToken, _ := tokens.Issue(tenantID, "00000000-0000-7000-8000-000000000001")
	deniedReq := httptest.NewRequest(http.MethodGet, "/api/v1/me/manifest", nil)
	deniedReq.Header.Set("Authorization", "Bearer "+deniedToken)
	deniedResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(deniedResponse, deniedReq)
	if deniedResponse.Code != http.StatusForbidden {
		t.Fatalf("user without role permission got status=%d", deniedResponse.Code)
	}
	if len(payload["modules"].([]any)) != 1 || payload["modules"].([]any)[0] != "core.test" {
		t.Fatalf("unexpected modules %v", payload["modules"])
	}
	if len(payload["capabilities"].([]any)) != 1 || payload["capabilities"].([]any)[0] != "test.capability" {
		t.Fatalf("unexpected capabilities %v", payload["capabilities"])
	}

	loginBody, _ := json.Marshal(map[string]string{"tenant_id": tenantID, "email": "owner@example.test", "password": "correct horse battery staple"})
	loginResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(loginResponse, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(string(loginBody))))
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	var pair map[string]any
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	oldRefresh := pair["refresh_token"].(string)
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": oldRefresh})
	refreshResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(refreshResponse, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(string(refreshBody))))
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refreshResponse.Code, refreshResponse.Body.String())
	}
	replayResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(replayResponse, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(string(refreshBody))))
	if replayResponse.Code != http.StatusUnauthorized {
		t.Fatalf("refresh replay status=%d", replayResponse.Code)
	}
}
func TestLoginRejectsMalformedRequest(t *testing.T) {
	tokens, _ := auth.NewTokenManager([]byte(strings.Repeat("k", 32)), time.Minute)
	srv := New(&db.Store{}, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"tenant_id":"bad","email":"x","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	srv.Handler().ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("login status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTerminalRegistrationPINLoginAndLockout(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run terminal auth integration test")
	}
	ctx := context.Background()
	store, err := db.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	tenantID, _ := ids.NewUUIDv7()
	businessID, _ := ids.NewUUIDv7()
	branchID, _ := ids.NewUUIDv7()
	userID, _ := ids.NewUUIDv7()
	roleID, _ := ids.NewUUIDv7()
	passwordHash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO tenants(id,name) VALUES($1,'terminal auth test')`, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO businesses(id,tenant_id,name) VALUES($1,$2,'Test business')`, businessID, tenantID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO branches(id,tenant_id,business_id,name) VALUES($1,$2,$3,'Test branch')`, branchID, tenantID, businessID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO users(id,tenant_id,email,password_hash) VALUES($1,$2,'owner@terminal.test',$3)`, userID, tenantID, passwordHash); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO user_branches(tenant_id,user_id,branch_id) VALUES($1,$2,$3)`, tenantID, userID, branchID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO roles(id,tenant_id,name) VALUES($1,$2,'owner')`, roleID, tenantID); err != nil {
			return err
		}
		for _, permission := range []string{"terminals.register", "users.manage"} {
			if _, err := tx.Exec(ctx, `INSERT INTO permissions(id,description) VALUES($1,'Test permission') ON CONFLICT DO NOTHING`, permission); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO role_permissions(tenant_id,role_id,permission_id) VALUES($1,$2,$3)`, tenantID, roleID, permission); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `INSERT INTO user_roles(tenant_id,user_id,role_id) VALUES($1,$2,$3)`, tenantID, userID, roleID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewTokenManager([]byte(strings.Repeat("t", 32)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	srv := New(store, tokens, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	ownerToken, _ := tokens.Issue(tenantID, userID)
	registerBody, _ := json.Marshal(map[string]string{"branch_id": branchID, "name": "Till one"})
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/terminals/register", strings.NewReader(string(registerBody)))
	registerReq.Header.Set("Authorization", "Bearer "+ownerToken)
	registerRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(registerRes, registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", registerRes.Code, registerRes.Body.String())
	}
	var terminal map[string]any
	if err := json.Unmarshal(registerRes.Body.Bytes(), &terminal); err != nil {
		t.Fatal(err)
	}
	setBody, _ := json.Marshal(map[string]string{"pin": "654321"})
	setReq := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+userID+"/pin", strings.NewReader(string(setBody)))
	setReq.Header.Set("Authorization", "Bearer "+ownerToken)
	setRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(setRes, setReq)
	if setRes.Code != http.StatusNoContent {
		t.Fatalf("set PIN status=%d body=%s", setRes.Code, setRes.Body.String())
	}
	loginBody, _ := json.Marshal(map[string]string{"tenant_id": tenantID, "terminal_id": terminal["terminal_id"].(string), "terminal_token": terminal["terminal_token"].(string), "user_id": userID, "pin": "654321"})
	loginRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(loginRes, httptest.NewRequest(http.MethodPost, "/api/v1/auth/pin-login", strings.NewReader(string(loginBody))))
	if loginRes.Code != http.StatusOK {
		t.Fatalf("PIN login status=%d body=%s", loginRes.Code, loginRes.Body.String())
	}
	var pair map[string]any
	if err := json.Unmarshal(loginRes.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.Verify(pair["access_token"].(string))
	if err != nil || claims.BranchID != branchID || claims.TerminalID != terminal["terminal_id"].(string) {
		t.Fatalf("PIN claims=(%+v,%v)", claims, err)
	}
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": pair["refresh_token"].(string)})
	refreshRes := httptest.NewRecorder()
	srv.Handler().ServeHTTP(refreshRes, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(string(refreshBody))))
	if refreshRes.Code != http.StatusOK {
		t.Fatalf("PIN refresh status=%d body=%s", refreshRes.Code, refreshRes.Body.String())
	}
	var refreshed map[string]any
	if err := json.Unmarshal(refreshRes.Body.Bytes(), &refreshed); err != nil {
		t.Fatal(err)
	}
	refreshedClaims, err := tokens.Verify(refreshed["access_token"].(string))
	if err != nil || refreshedClaims.BranchID != branchID || refreshedClaims.TerminalID != terminal["terminal_id"].(string) {
		t.Fatalf("refreshed claims=(%+v,%v)", refreshedClaims, err)
	}
	wrongPINBody, _ := json.Marshal(map[string]string{"tenant_id": tenantID, "terminal_id": terminal["terminal_id"].(string), "terminal_token": terminal["terminal_token"].(string), "user_id": userID, "pin": "111111"})
	for i := 0; i < 5; i++ {
		wrong := httptest.NewRecorder()
		srv.Handler().ServeHTTP(wrong, httptest.NewRequest(http.MethodPost, "/api/v1/auth/pin-login", strings.NewReader(string(wrongPINBody))))
		if wrong.Code != http.StatusUnauthorized {
			t.Fatalf("wrong PIN attempt %d status=%d", i+1, wrong.Code)
		}
	}
	locked := httptest.NewRecorder()
	srv.Handler().ServeHTTP(locked, httptest.NewRequest(http.MethodPost, "/api/v1/auth/pin-login", strings.NewReader(string(loginBody))))
	if locked.Code != http.StatusUnauthorized {
		t.Fatalf("locked user PIN status=%d", locked.Code)
	}
}
