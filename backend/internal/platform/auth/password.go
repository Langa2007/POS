package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
)

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", errors.New("password must be at least 12 characters")
	}
	return hashSecret(password)
}

func HashPIN(pin string) (string, error) {
	if len(pin) != 6 {
		return "", errors.New("PIN must contain exactly six digits")
	}
	for _, r := range pin {
		if r < '0' || r > '9' {
			return "", errors.New("PIN must contain exactly six digits")
		}
	}
	return hashSecret(pin)
}

func hashSecret(secret string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(secret), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, enc.EncodeToString(salt), enc.EncodeToString(hash)), nil
}

func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false
	}
	memory, e1 := param(params[0], "m")
	time, e2 := param(params[1], "t")
	threads, e3 := param(params[2], "p")
	enc := base64.RawStdEncoding
	salt, e4 := enc.DecodeString(parts[4])
	want, e5 := enc.DecodeString(parts[5])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || memory < 8*1024 || memory > 1024*1024 || time < 1 || time > 10 || threads < 1 || threads > 16 || len(salt) < 8 || len(salt) > 64 || len(want) < 16 || len(want) > 64 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func param(value, key string) (uint32, error) {
	pair := strings.SplitN(value, "=", 2)
	if len(pair) != 2 || pair[0] != key {
		return 0, errors.New("invalid argon2 parameter")
	}
	n, err := strconv.ParseUint(pair[1], 10, 32)
	return uint32(n), err
}
