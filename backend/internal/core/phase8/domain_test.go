package phase8

import (
	"testing"
	"time"
)

func TestWebhookSignature(t *testing.T) {
	payload := []byte(`{"id":"evt-1"}`)
	signature := WebhookSignature("secret", payload)
	if !VerifyWebhook("secret", payload, signature) || VerifyWebhook("wrong", payload, signature) {
		t.Fatal("webhook signature verification failed")
	}
}

func TestAutoMatchLeavesAmbiguousLinesAsExceptions(t *testing.T) {
	matches, exceptions := AutoMatch(
		[]StatementLine{{ID: "line-1", AmountMinor: 1000}},
		[]Candidate{{TargetID: "sale-1", AmountMinor: 1000}, {TargetID: "sale-2", AmountMinor: 1000}},
	)
	if len(matches) != 0 || len(exceptions) != 1 || exceptions[0] != "line-1" {
		t.Fatalf("matches=%v exceptions=%v", matches, exceptions)
	}
}

func TestPayoutMakerChecker(t *testing.T) {
	payout := Payout{RequestedBy: "maker", Status: "pending"}
	if payout.Approve("maker") == nil {
		t.Fatal("maker approved own payout")
	}
	if err := payout.Approve("checker"); err != nil || payout.Status != "approved" {
		t.Fatalf("approval failed: %v", err)
	}
}

func TestProratedCharge(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	got, err := ProratedCharge(3100, start, end, start.AddDate(0, 0, 15), end)
	if err != nil || got != 1600 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
