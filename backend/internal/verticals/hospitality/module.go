package hospitality

import (
	"context"
	_ "embed"
)

//go:embed ui/manifest.json
var manifest []byte

type Module struct{}

func (m *Module) ID() string      { return "hospitality" }
func (m *Module) Version() string { return "0.1.0" }
func (m *Module) Capabilities() []string {
	return []string{"room_booking", "rate_plans", "guest_folio", "housekeeping", "night_audit", "group_billing", "deposits"}
}
func (m *Module) DependsOn() []string               { return []string{} }
func (m *Module) Manifest() []byte                  { return manifest }
func (m *Module) Migrate(ctx context.Context) error { return nil }
