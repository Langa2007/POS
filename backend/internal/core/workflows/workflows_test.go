package workflows

import "testing"

func TestCanTransition(t *testing.T) {
	d := Definition{States: []string{"open", "closed"}, Transitions: []Transition{{From: "open", To: "closed"}}}
	if !CanTransition(d, "open", "closed") {
		t.Fatal("expected true")
	}
	if CanTransition(d, "closed", "open") {
		t.Fatal("expected false")
	}
}
