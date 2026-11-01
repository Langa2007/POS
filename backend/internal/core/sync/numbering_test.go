package sync

import "testing"

func TestOfflineReceiptNumber(t *testing.T) {
	got, err := OfflineReceiptNumber("till-01", 12)
	if err != nil {
		t.Fatal(err)
	}
	if got != "TILL-01-000012" {
		t.Fatalf("receipt=%q", got)
	}
}
