package sales

import "testing"

func TestTotalAndChange(t *testing.T) {
	total, err := Total([]Line{{Quantity: 2, UnitPriceMinor: 1250}, {Quantity: 1, UnitPriceMinor: 500, DiscountMinor: 50}})
	if err != nil || total != 2950 {
		t.Fatalf("total=%d err=%v", total, err)
	}
	change, err := ChangeDue(total, 3000)
	if err != nil || change != 50 {
		t.Fatalf("change=%d err=%v", change, err)
	}
}

func TestRejectsInvalidLine(t *testing.T) {
	if _, err := Total([]Line{{Quantity: 0, UnitPriceMinor: 1}}); err == nil {
		t.Fatal("expected invalid quantity")
	}
}

func TestReturnAndShiftVariance(t *testing.T) {
	refund, err := ReturnAmount(1250, 2, 1)
	if err != nil || refund != 1250 {
		t.Fatalf("refund=%d err=%v", refund, err)
	}
	variance, err := ShiftVariance(1000, 5000, 200, 100, 5900)
	if err != nil || variance != -200 {
		t.Fatalf("variance=%d err=%v", variance, err)
	}
}
