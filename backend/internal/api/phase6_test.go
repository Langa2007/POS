package api

import (
	"encoding/json"
	"testing"

	"github.com/Langa2007/POS/backend/internal/verticals/pharmacy"
	"github.com/Langa2007/POS/backend/internal/verticals/restaurant"
	"github.com/Langa2007/POS/backend/internal/verticals/retail"
)

type phase6Module interface {
	ID() string
	Capabilities() []string
	Manifest() []byte
}

func TestPhase6ModulesAreIsolated(t *testing.T) {
	modules := []phase6Module{&retail.Module{}, &pharmacy.Module{}, &restaurant.Module{}}
	seenIDs := make(map[string]bool, len(modules))
	seenCapabilities := make(map[string]string)
	for _, module := range modules {
		if seenIDs[module.ID()] {
			t.Fatalf("duplicate module id %q", module.ID())
		}
		seenIDs[module.ID()] = true

		var manifest struct {
			ID           string   `json:"id"`
			Capabilities []string `json:"capabilities"`
		}
		if err := json.Unmarshal(module.Manifest(), &manifest); err != nil {
			t.Fatalf("%s manifest is invalid: %v", module.ID(), err)
		}
		if manifest.ID != module.ID() {
			t.Fatalf("%s manifest id=%q", module.ID(), manifest.ID)
		}
		for _, capability := range module.Capabilities() {
			if owner, exists := seenCapabilities[capability]; exists && owner != module.ID() {
				t.Fatalf("capability %q shared by %s and %s", capability, owner, module.ID())
			}
			seenCapabilities[capability] = module.ID()
		}
	}
}
