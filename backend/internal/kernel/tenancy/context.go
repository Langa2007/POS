// Package tenancy carries the authenticated tenant scope through request work.
package tenancy

import (
	"context"
	"errors"
)

var ErrMissingTenant = errors.New("tenant context is required")

type Scope struct {
	TenantID   string
	BusinessID string
	BranchID   string
	TerminalID string
	UserID     string
}

type scopeKey struct{}

func WithScope(ctx context.Context, scope Scope) context.Context {
	return context.WithValue(ctx, scopeKey{}, scope)
}

func FromContext(ctx context.Context) (Scope, bool) {
	scope, ok := ctx.Value(scopeKey{}).(Scope)
	return scope, ok && scope.TenantID != ""
}

func Require(ctx context.Context) (Scope, error) {
	scope, ok := FromContext(ctx)
	if !ok {
		return Scope{}, ErrMissingTenant
	}
	return scope, nil
}
