package module

import "context"

// Module is the runtime contract implemented by every vertical.
type Module interface {
	ID() string
	Version() string
	Capabilities() []string
	DependsOn() []string
	Manifest() []byte
	Migrate(context.Context) error
}
