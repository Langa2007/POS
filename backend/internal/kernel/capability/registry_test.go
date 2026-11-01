package capability

import (
	"errors"
	"testing"
)

func TestValidateDependenciesAndConflicts(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Definition{ID: "inventory"})
	_ = r.Register(Definition{ID: "expiry", Requires: []string{"inventory"}, Conflicts: []string{"no-stock"}})
	_ = r.Register(Definition{ID: "no-stock"})
	if err := r.Validate([]string{"expiry"}); err == nil {
		t.Fatal("expected missing requirement")
	}
	if err := r.Validate([]string{"inventory", "expiry", "no-stock"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if err := r.Validate([]string{"inventory", "expiry"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Validate([]string{"unknown"}); !errors.Is(err, ErrUnknown) {
		t.Fatalf("expected unknown, got %v", err)
	}
}
