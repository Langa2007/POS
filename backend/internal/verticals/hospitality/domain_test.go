package hospitality

import (
	"testing"
	"time"
)

func TestStayAndIdempotentFolio(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := Stay{CheckIn: base, CheckOut: base.Add(24 * time.Hour)}
	b := Stay{CheckIn: base.Add(12 * time.Hour), CheckOut: base.Add(48 * time.Hour)}
	if !Overlaps(a, b) {
		t.Fatal("overlap not detected")
	}
	entries := AppendOnce(nil, FolioEntry{ID: "restaurant-1", AmountMinor: 100})
	entries = AppendOnce(entries, FolioEntry{ID: "restaurant-1", AmountMinor: 100})
	if len(entries) != 1 {
		t.Fatalf("entries=%d", len(entries))
	}
}
