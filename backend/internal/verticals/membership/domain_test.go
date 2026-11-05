package membership

import (
	"testing"
	"time"
)

func TestBillingIsIdempotent(t *testing.T) {
	run := BillingRun{Key: "2026-01"}
	if !run.Process("member-1", time.Now()) {
		t.Fatal("first billing skipped")
	}
	if run.Process("member-1", time.Now()) {
		t.Fatal("duplicate billing processed")
	}
}
