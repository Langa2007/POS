package fuelstation

import (
	"context"
	_ "embed"
)

//go:embed ui/manifest.json
var manifest []byte

type Module struct{}

func (m *Module) ID() string      { return "fuelstation" }
func (m *Module) Version() string { return "0.1.0" }
func (m *Module) Capabilities() []string {
	return []string{"pump_integration", "tank_dips", "meter_shifts", "pump_reconciliation"}
}
func (m *Module) DependsOn() []string               { return []string{} }
func (m *Module) Manifest() []byte                  { return manifest }
func (m *Module) Migrate(ctx context.Context) error { return nil }
