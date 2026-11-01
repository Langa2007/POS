package secrets

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	encrypted, err := Encrypt(key, []byte("mpesa-secret"))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Decrypt(key, encrypted)
	if err != nil || string(plain) != "mpesa-secret" {
		t.Fatalf("plain=%q err=%v", plain, err)
	}
}
