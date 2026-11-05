package garage

import "testing"

func TestTotal(t *testing.T) {
	got, err := Total([]JobLine{{Kind: "parts", AmountMinor: 100}, {Kind: "labor", AmountMinor: 250}})
	if err != nil || got != 350 {
		t.Fatalf("got %d, %v", got, err)
	}
}
