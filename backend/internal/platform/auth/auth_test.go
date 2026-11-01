package auth

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordHashVerification(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash format %q", hash)
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Fatal("password did not verify")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Fatal("wrong password verified")
	}
	if VerifyPassword("$argon2id$v=19$m=999999999,t=2,p=2$bad$bad", "pw") {
		t.Fatal("malformed hash verified")
	}
}

func TestPINHashValidation(t *testing.T) {
	hash, err := HashPIN("004207")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "004207") {
		t.Fatal("valid PIN did not verify")
	}
	if VerifyPassword(hash, "004208") {
		t.Fatal("wrong PIN verified")
	}
	for _, pin := range []string{"12345", "1234567", "12a456"} {
		if _, err := HashPIN(pin); err == nil {
			t.Errorf("accepted invalid PIN %q", pin)
		}
	}
}
func TestTokenIssueVerifyAndExpiry(t *testing.T) {
	m, err := NewTokenManager([]byte(strings.Repeat("s", 32)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return time.Unix(1000, 0) }
	token, err := m.Issue("tenant", "user")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.Verify(token)
	if err != nil || claims.TenantID != "tenant" || claims.UserID != "user" {
		t.Fatalf("Verify()=(%+v,%v)", claims, err)
	}
	m.now = func() time.Time { return time.Unix(1061, 0) }
	if _, err = m.Verify(token); err == nil {
		t.Fatal("expired token accepted")
	}
}
