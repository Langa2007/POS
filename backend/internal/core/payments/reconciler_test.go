package payments

import "testing"

func TestResolveStuckIntent(t *testing.T) {
	decision, err := ResolveStuckIntent("checkout-1", Pending, Captured)
	if err != nil || decision.To != Captured {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	if _, err := ResolveStuckIntent("checkout-1", Pending, Pending); err == nil {
		t.Fatal("expected non-terminal status error")
	}
}
