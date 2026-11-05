package fuelstation

import "testing"

func TestReconcileMeter(t *testing.T) {
	if !ReconcileMeter(100, 250, 10, 160, 0.01) {
		t.Fatal("expected reconciled meter")
	}
	if ReconcileMeter(100, 250, 10, 140, 0.01) {
		t.Fatal("accepted out-of-tolerance meter")
	}
}
