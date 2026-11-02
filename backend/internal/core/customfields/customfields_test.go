package customfields

import "testing"

func TestAllowed(t *testing.T) {
	if !allowed("text") || !allowed("select") {
		t.Fatal("bad")
	}
	if allowed("script") {
		t.Fatal("bad")
	}
}
