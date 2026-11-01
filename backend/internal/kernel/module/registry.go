// Package module defines the contracts and dependency ordering for platform modules.
package module

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrDuplicate         = errors.New("duplicate module")
	ErrMissingDependency = errors.New("missing module dependency")
	ErrDependencyCycle   = errors.New("module dependency cycle")
)

type Definition struct {
	ID           string
	Dependencies []string
}

type Registry struct{ modules map[string]Definition }

func NewRegistry() *Registry { return &Registry{modules: make(map[string]Definition)} }

func (r *Registry) Register(def Definition) error {
	if def.ID == "" {
		return errors.New("module ID is required")
	}
	if _, exists := r.modules[def.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicate, def.ID)
	}
	deps := append([]string(nil), def.Dependencies...)
	sort.Strings(deps)
	def.Dependencies = deps
	r.modules[def.ID] = def
	return nil
}

// Order returns a deterministic dependency-first ordering of registered modules.
func (r *Registry) Order() ([]string, error) {
	state := make(map[string]uint8, len(r.modules))
	ordered := make([]string, 0, len(r.modules))
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("%w at %s", ErrDependencyCycle, id)
		case 2:
			return nil
		}
		def, ok := r.modules[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrMissingDependency, id)
		}
		state[id] = 1
		for _, dep := range def.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[id] = 2
		ordered = append(ordered, id)
		return nil
	}
	ids := make([]string, 0, len(r.modules))
	for id := range r.modules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
