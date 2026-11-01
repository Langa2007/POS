package payments

import (
	"context"
	"net/http"
	"testing"
)

type testProvider struct{}

func (testProvider) Key() string { return "test" }
func (testProvider) Methods() []MethodDefinition { return []MethodDefinition{{Method: MethodSTK, Enabled: true}} }
func (testProvider) Initiate(context.Context, Intent) (IntentResult, error) { return IntentResult{}, nil }
func (testProvider) Status(context.Context, string) (Status, error) { return Pending, nil }
func (testProvider) Refund(context.Context, RefundRequest) (RefundResult, error) { return RefundResult{}, nil }
func (testProvider) HandleCallback(context.Context, []byte, http.Header) (CallbackEvent, error) { return CallbackEvent{}, nil }
func (testProvider) Verify(context.Context, []byte, http.Header) error { return nil }

func TestProviderRegistry(t *testing.T) {
	registry := NewRegistry()
	provider := testProvider{}
	if err := registry.Register(provider); err != nil { t.Fatal(err) }
	if _, ok := registry.Get("test"); !ok { t.Fatal("provider was not registered") }
	if err := registry.Register(provider); err == nil { t.Fatal("expected duplicate provider error") }
}
