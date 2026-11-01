package ids

import (
	"regexp"
	"testing"
)

func TestNewUUIDv7(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id, err := NewUUIDv7()
		if err != nil {
			t.Fatal(err)
		}
		if !pattern.MatchString(id) {
			t.Fatalf("invalid UUIDv7 %q", id)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID %q", id)
		}
		seen[id] = true
	}
}
