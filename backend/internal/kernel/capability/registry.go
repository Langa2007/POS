// Package capability validates named capabilities and their relationships.
package capability

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrDuplicate = errors.New("duplicate capability")
	ErrUnknown   = errors.New("unknown capability")
	ErrConflict  = errors.New("conflicting capabilities")
)

type Definition struct {
	ID        string
	Requires  []string
	Conflicts []string
}

type Registry struct{ definitions map[string]Definition }

func NewRegistry() *Registry { return &Registry{definitions: make(map[string]Definition)} }

func (r *Registry) Register(def Definition) error {
	if def.ID == "" {
		return errors.New("capability ID is required")
	}
	if _, exists := r.definitions[def.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicate, def.ID)
	}
	def.Requires = uniqueSorted(def.Requires)
	def.Conflicts = uniqueSorted(def.Conflicts)
	r.definitions[def.ID] = def
	return nil
}

// Validate checks that every enabled capability exists, its requirements are enabled,
// and no enabled pair declares a conflict.
func (r *Registry) Validate(enabled []string) error {
	set := make(map[string]bool, len(enabled))
	for _, id := range enabled {
		if set[id] {
			continue
		}
		if _, ok := r.definitions[id]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknown, id)
		}
		set[id] = true
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		def := r.definitions[id]
		for _, req := range def.Requires {
			if _, ok := r.definitions[req]; !ok {
				return fmt.Errorf("%w: %s requires undefined %s", ErrUnknown, id, req)
			}
			if !set[req] {
				return fmt.Errorf("capability %s requires %s", id, req)
			}
		}
		for _, conflict := range def.Conflicts {
			if set[conflict] {
				return fmt.Errorf("%w: %s and %s", ErrConflict, id, conflict)
			}
		}
	}
	return nil
}

func uniqueSorted(values []string) []string {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = true
		}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
