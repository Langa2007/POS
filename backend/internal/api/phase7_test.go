package api

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/Langa2007/POS/backend/internal/kernel/module"
	"github.com/Langa2007/POS/backend/internal/verticals/bookshop"
	"github.com/Langa2007/POS/backend/internal/verticals/dealership"
	"github.com/Langa2007/POS/backend/internal/verticals/fuelstation"
	"github.com/Langa2007/POS/backend/internal/verticals/garage"
	"github.com/Langa2007/POS/backend/internal/verticals/hardware"
	"github.com/Langa2007/POS/backend/internal/verticals/hospitality"
	"github.com/Langa2007/POS/backend/internal/verticals/membership"
	"github.com/Langa2007/POS/backend/internal/verticals/salon"
)

func TestPhase7ModuleContracts(t *testing.T) {
	modules := []module.Module{
		&hardware.Module{}, &bookshop.Module{}, &hospitality.Module{},
		&dealership.Module{}, &salon.Module{}, &garage.Module{},
		&fuelstation.Module{}, &membership.Module{},
	}
	for _, m := range modules {
		if m.ID() == "" || m.Version() == "" {
			t.Fatalf("module has incomplete identity: %#v", m)
		}
		var manifest struct {
			ID           string   `json:"id"`
			Version      string   `json:"version"`
			Capabilities []string `json:"capabilities"`
			Permissions  []string `json:"permissions"`
		}
		if err := json.Unmarshal(m.Manifest(), &manifest); err != nil {
			t.Fatalf("%s manifest is invalid: %v", m.ID(), err)
		}
		if manifest.ID != m.ID() || manifest.Version != m.Version() {
			t.Fatalf("%s manifest identity does not match module", m.ID())
		}
		if !reflect.DeepEqual(manifest.Capabilities, m.Capabilities()) {
			t.Fatalf("%s manifest capabilities do not match module", m.ID())
		}
		if len(manifest.Permissions) == 0 {
			t.Fatalf("%s has no permissions", m.ID())
		}
	}
}

func TestPhase7ModulesCanRegister(t *testing.T) {
	modules := []module.Module{
		&hardware.Module{}, &bookshop.Module{}, &hospitality.Module{},
		&dealership.Module{}, &salon.Module{}, &garage.Module{},
		&fuelstation.Module{}, &membership.Module{},
	}
	registry := module.NewRegistry()
	for _, m := range modules {
		if err := registry.Register(module.Definition{ID: m.ID(), Dependencies: m.DependsOn()}); err != nil {
			t.Fatal(err)
		}
	}
	order, err := registry.Order()
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != len(modules) {
		t.Fatalf("registered %d modules, got order of %d", len(modules), len(order))
	}
}

func TestPhase7PostgresGateConfigured(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
}
