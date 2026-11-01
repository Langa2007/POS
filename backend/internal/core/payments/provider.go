package payments

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type Method string

const (
	MethodSTK       Method = "mpesa_stk"
	MethodC2B       Method = "mpesa_c2b"
	MethodRefund    Method = "refund"
	MethodStatus    Method = "status"
	MethodReversal  Method = "reversal"
	MethodB2C       Method = "b2c"
	MethodQR        Method = "qr"
)

type MethodDefinition struct {
	Method Method
	Enabled bool
}

type Intent struct {
	ID             string
	AmountMinor    int64
	Currency       string
	Phone          string
	AccountReference string
	Description    string
}

type IntentResult struct {
	ProviderReference string
	MerchantReference string
	Status            Status
	Raw               []byte
}

type RefundRequest struct {
	IntentReference string
	AmountMinor     int64
	Reason          string
}

type RefundResult struct {
	ProviderReference string
	Status            Status
}

type CallbackEvent struct {
	ProviderReference string
	Status            Status
	AmountMinor       int64
	Receipt           string
	Raw               []byte
}

type Provider interface {
	Key() string
	Methods() []MethodDefinition
	Initiate(context.Context, Intent) (IntentResult, error)
	Status(context.Context, string) (Status, error)
	Refund(context.Context, RefundRequest) (RefundResult, error)
	HandleCallback(context.Context, []byte, http.Header) (CallbackEvent, error)
	Verify(context.Context, []byte, http.Header) error
}

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(provider Provider) error {
	if provider == nil || provider.Key() == "" {
		return errors.New("provider key is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[provider.Key()]; exists {
		return errors.New("payment provider is already registered")
	}
	r.providers[provider.Key()] = provider
	return nil
}

func (r *Registry) Get(key string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[key]
	return provider, ok
}
