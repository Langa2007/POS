package tenancy

import (
	"context"
	"errors"
	"testing"
)

func TestRequireScope(t *testing.T) {
	if _, err := Require(context.Background()); !errors.Is(err, ErrMissingTenant) {
		t.Fatalf("got %v", err)
	}
	want := Scope{TenantID: "tenant", UserID: "user"}
	got, err := Require(WithScope(context.Background(), want))
	if err != nil || got != want {
		t.Fatalf("Require()=(%+v,%v)", got, err)
	}
}
