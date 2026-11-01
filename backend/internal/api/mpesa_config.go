package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Langa2007/POS/backend/internal/adapters/mpesa"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/platform/secrets"
	"github.com/jackc/pgx/v5"
)

type mpesaStoredConfig struct {
	mpesa.DarajaConfig
	CallbackToken string `json:"callback_token"`
}

func (s *Server) configureMPesa(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	if len(s.credentialKey) != 32 {
		writeError(w, 503, "payments.credentials_unavailable", "POS_CREDENTIALS_KEY must be a base64 encoded or raw 32-byte key")
		return
	}
	var cfg mpesaStoredConfig
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&cfg) != nil {
		writeError(w, 400, "payments.config_invalid", "Invalid M-Pesa configuration")
		return
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("POS_PUBLIC_URL")), "/")
	u, parseErr := url.Parse(base)
	if parseErr != nil || u == nil || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) || u.Host == "" {
		writeError(w, 400, "payments.public_url_required", "Set POS_PUBLIC_URL to the publicly reachable HTTPS API base URL")
		return
	}
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	cfg.CallbackToken = base64.RawURLEncoding.EncodeToString(rawToken)
	pathToken := scope.TenantID + "." + cfg.CallbackToken
	cfg.CallbackURL = base + "/api/v1/callbacks/mpesa/stk/" + pathToken
	cfg.ConfirmationURL = base + "/api/v1/callbacks/mpesa/c2b/" + pathToken
	cfg.ValidationURL = base + "/api/v1/callbacks/mpesa/c2b/validate/" + pathToken
	provider, err := mpesa.NewDaraja(cfg.DarajaConfig, nil)
	if err != nil {
		writeError(w, 400, "payments.config_invalid", err.Error())
		return
	}
	_ = provider
	plain, err := json.Marshal(cfg)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	ciphertext, err := secrets.Encrypt(s.credentialKey, plain)
	if err != nil {
		writeError(w, 500, "internal.error", "Request could not be completed")
		return
	}
	hash := sha256.Sum256([]byte(cfg.CallbackToken))
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `INSERT INTO mpesa_configurations(tenant_id,environment,business_shortcode,encrypted_config,callback_token_hash,enabled,updated_at) VALUES($1,$2,$3,$4,$5,true,now()) ON CONFLICT(tenant_id) DO UPDATE SET environment=excluded.environment,business_shortcode=excluded.business_shortcode,encrypted_config=excluded.encrypted_config,callback_token_hash=excluded.callback_token_hash,enabled=true,c2b_registered_at=NULL,updated_at=now()`, scope.TenantID, cfg.Environment, cfg.BusinessShortCode, ciphertext, hash[:])
		return err
	})
	if err != nil {
		writeError(w, 400, "payments.config_save_failed", "M-Pesa configuration could not be saved")
		return
	}
	writeJSON(w, 200, map[string]any{"enabled": true, "environment": cfg.Environment, "business_shortcode": cfg.BusinessShortCode, "stk_callback_url": cfg.CallbackURL, "c2b_confirmation_url": cfg.ConfirmationURL, "c2b_validation_url": cfg.ValidationURL, "callback_token": cfg.CallbackToken})
}

func (s *Server) getMPesaConfig(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	var environment, shortcode string
	var enabled bool
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT environment,business_shortcode,enabled FROM mpesa_configurations WHERE tenant_id=$1`, scope.TenantID).Scan(&environment, &shortcode, &enabled)
	})
	if err != nil {
		writeError(w, 404, "payments.config_missing", "M-Pesa is not configured")
		return
	}
	writeJSON(w, 200, map[string]any{"environment": environment, "business_shortcode": shortcode, "enabled": enabled, "credentials_configured": true})
}

func (s *Server) loadMPesaConfig(r *http.Request, tenantID string) (*mpesa.Daraja, mpesaStoredConfig, error) {
	if len(s.credentialKey) != 32 {
		return nil, mpesaStoredConfig{}, errors.New("credential encryption key unavailable")
	}
	var ciphertext []byte
	err := s.db.TenantTx(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `SELECT encrypted_config FROM mpesa_configurations WHERE tenant_id=$1 AND enabled`, tenantID).Scan(&ciphertext)
	})
	if err != nil {
		return nil, mpesaStoredConfig{}, err
	}
	fingerprint := sha256.Sum256(ciphertext)
	s.mpesaMu.Lock()
	defer s.mpesaMu.Unlock()
	if cached, ok := s.mpesaProviders[tenantID]; ok && cached.fingerprint == fingerprint {
		return cached.provider, cached.config, nil
	}
	plain, err := secrets.Decrypt(s.credentialKey, ciphertext)
	if err != nil {
		return nil, mpesaStoredConfig{}, err
	}
	var cfg mpesaStoredConfig
	if err := json.Unmarshal(plain, &cfg); err != nil {
		return nil, cfg, err
	}
	provider, err := mpesa.NewDaraja(cfg.DarajaConfig, nil)
	if err == nil {
		s.mpesaProviders[tenantID] = cachedMPesaProvider{fingerprint: fingerprint, provider: provider, config: cfg}
	}
	return provider, cfg, err
}

func (s *Server) registerMPesaC2B(w http.ResponseWriter, r *http.Request) {
	scope, err := tenancy.Require(r.Context())
	if err != nil {
		writeError(w, 401, "auth.required", "A valid bearer token is required")
		return
	}
	provider, cfg, err := s.loadMPesaConfig(r, scope.TenantID)
	if err != nil {
		writeError(w, 400, "payments.provider_unavailable", err.Error())
		return
	}
	if cfg.ValidationURL == "" {
		writeError(w, 400, "payments.c2b_unavailable", "M-Pesa configuration must include a validation URL before C2B registration")
		return
	}
	if err := provider.RegisterC2B(r.Context()); err != nil {
		writeError(w, 502, "payments.provider_error", "M-Pesa C2B URL registration failed")
		return
	}
	err = s.db.TenantTx(r.Context(), scope.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `UPDATE mpesa_configurations SET c2b_registered_at=now() WHERE tenant_id=$1`, scope.TenantID)
		return err
	})
	if err != nil {
		writeError(w, 500, "payments.registration_record_failed", "Registration succeeded but status could not be recorded")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "registered"})
}

func callbackCredentials(pathToken string) (string, string, error) {
	parts := strings.SplitN(pathToken, ".", 2)
	if len(parts) != 2 || !uuidPattern.MatchString(parts[0]) || len(parts[1]) < 32 {
		return "", "", fmt.Errorf("invalid callback token")
	}
	if _, err := base64.RawURLEncoding.DecodeString(parts[1]); err != nil {
		return "", "", fmt.Errorf("invalid callback token")
	}
	return parts[0], parts[1], nil
}

func callbackTokenValid(expected []byte, token string) bool {
	sum := sha256.Sum256([]byte(token))
	return len(expected) == len(sum) && subtle.ConstantTimeCompare(expected, sum[:]) == 1
}
