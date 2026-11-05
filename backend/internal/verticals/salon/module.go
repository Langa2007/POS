package salon

import (
	"context"
	_ "embed"
)

//go:embed ui/manifest.json
var manifest []byte

type Module struct{}

func (m *Module) ID() string      { return "salon" }
func (m *Module) Version() string { return "0.1.0" }
func (m *Module) Capabilities() []string {
	return []string{"appointments", "commissions", "packages", "loyalty"}
}
func (m *Module) DependsOn() []string               { return []string{} }
func (m *Module) Manifest() []byte                  { return manifest }
func (m *Module) Migrate(ctx context.Context) error { return nil }
