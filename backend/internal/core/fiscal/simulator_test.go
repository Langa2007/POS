package fiscal

import (
	"context"
	"testing"
)

func TestSimulatorIsStableAndValidatesDocuments(t *testing.T) {
	s := Simulator{}
	d := Document{TenantID: "t", Type: "receipt", ID: "sale", IdempotencyKey: "key"}
	a, e := s.Submit(context.Background(), d)
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.Submit(context.Background(), d)
	if e != nil || a != b {
		t.Fatalf("not idempotent: %+v %+v %v", a, b, e)
	}
	if _, e = s.Submit(context.Background(), Document{}); e == nil {
		t.Fatal("invalid fiscal document accepted")
	}
}
