package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid or expired access token")

type Claims struct {
	TenantID   string `json:"tenant_id"`
	UserID     string `json:"user_id"`
	BranchID   string `json:"branch_id,omitempty"`
	TerminalID string `json:"terminal_id,omitempty"`
	ExpiresAt  int64  `json:"exp"`
	IssuedAt   int64  `json:"iat"`
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret []byte, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("token lifetime must be positive")
	}
	return &TokenManager{secret: append([]byte(nil), secret...), ttl: ttl, now: time.Now}, nil
}

func (m *TokenManager) Issue(tenantID, userID string) (string, error) {
	return m.IssueForScope(tenantID, userID, "", "")
}

func (m *TokenManager) IssueForScope(tenantID, userID, branchID, terminalID string) (string, error) {
	now := m.now()
	head, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	claims, err := json.Marshal(Claims{TenantID: tenantID, UserID: userID, BranchID: branchID, TerminalID: terminalID, IssuedAt: now.Unix(), ExpiresAt: now.Add(m.ttl).Unix()})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(head) + "." + base64.RawURLEncoding.EncodeToString(claims)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (m *TokenManager) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var h map[string]string
	if json.Unmarshal(header, &h) != nil || h["alg"] != "HS256" || h["typ"] != "JWT" {
		return Claims{}, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if json.Unmarshal(payload, &claims) != nil || claims.TenantID == "" || claims.UserID == "" || claims.ExpiresAt <= m.now().Unix() || claims.IssuedAt > m.now().Add(time.Minute).Unix() {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func Bearer(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("%w", ErrInvalidToken)
	}
	return parts[1], nil
}

// NewRefreshToken returns a high-entropy opaque refresh token. Persist only its hash.
func NewRefreshToken(tenantID string) (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return tenantID + "." + base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
