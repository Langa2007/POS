package hardware

import "testing"

func TestRoundTrip(t *testing.T) {
	got, err := RoundTrip(3, 12, 1.0/12)
	if err != nil || got < 2.999999 || got > 3.000001 {
		t.Fatalf("got %v, %v", got, err)
	}
}
