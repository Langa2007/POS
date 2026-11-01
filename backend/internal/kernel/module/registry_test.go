package module

import (
	"errors"
	"reflect"
	"testing"
)

func TestOrderDependenciesDeterministically(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Definition{ID: "checkout", Dependencies: []string{"catalog", "users"}})
	_ = r.Register(Definition{ID: "users"})
	_ = r.Register(Definition{ID: "catalog"})
	got, err := r.Order()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"catalog", "users", "checkout"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Order()=%v, want %v", got, want)
	}
}
func TestOrderRejectsMissingAndCycle(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(Definition{ID: "a", Dependencies: []string{"missing"}})
	if _, err := r.Order(); !errors.Is(err, ErrMissingDependency) {
		t.Fatalf("got %v", err)
	}
	r = NewRegistry()
	_ = r.Register(Definition{ID: "a", Dependencies: []string{"b"}})
	_ = r.Register(Definition{ID: "b", Dependencies: []string{"a"}})
	if _, err := r.Order(); !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("got %v", err)
	}
}
