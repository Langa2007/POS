package rules

import (
	"testing"
)

func TestSimulateAppliesDiscount(t *testing.T) {
	// basic test without DB
	matched := []map[string]any{{"rule_id": "r1", "actions": []any{map[string]any{"type": "set_discount_percent", "value": map[string]any{"percent": 10}}}}}
	out := ApplyDiscounts(1000, matched)
	if out != 900 {
		t.Fatalf("expected 900 got %d", out)
	}
}

func TestApplyDiscountsNeverNegative(t *testing.T) {
	matched := []map[string]any{{"rule_id": "r1", "actions": []any{map[string]any{"type": "set_discount_fixed", "value": map[string]any{"amount_minor": 2000}}}}}
	out := ApplyDiscounts(1000, matched)
	if out != 0 {
		t.Fatalf("expected 0 got %d", out)
	}
}
