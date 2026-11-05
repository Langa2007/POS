package hardware

import (
	"context"
	_ "embed"
)

//go:embed ui/manifest.json
var manifest []byte

type Module struct{}

func (m *Module) ID() string      { return "hardware" }
func (m *Module) Version() string { return "0.1.0" }
func (m *Module) Capabilities() []string {
	return []string{"multi_uom", "cut_to_length", "credit_accounts", "contract_pricing", "quotes", "delivery"}
}
func (m *Module) DependsOn() []string               { return []string{} }
func (m *Module) Manifest() []byte                  { return manifest }
func (m *Module) Migrate(ctx context.Context) error { return nil }
