package mpesa

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Simulator struct {
	mu            sync.Mutex
	Mode          string
	CallbackDelay time.Duration
	Requests      []STKRequest
}

func (s *Simulator) InitiateSTK(_ context.Context, request STKRequest) (STKResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Requests = append(s.Requests, request)
	if s.Mode == "provider_error" {
		return STKResponse{}, ErrProviderUnavailable
	}
	return STKResponse{CheckoutRequestID: fmt.Sprintf("ws_CO_%d", len(s.Requests)), MerchantRequestID: "simulator", ResponseCode: "0", ResponseDescription: "Success"}, nil
}

func (s *Simulator) QueryStatus(_ context.Context, _ string) (string, error) {
	if s.Mode == "timeout" {
		return "", ErrProviderUnavailable
	}
	if s.Mode == "failed" {
		return "Failed", nil
	}
	return "Completed", nil
}

func (s *Simulator) Reverse(_ context.Context, _ string, _ int64) error {
	if s.Mode == "provider_error" {
		return ErrProviderUnavailable
	}
	return nil
}

func (s *Simulator) ParseCallback(body []byte) (Callback, error) {
	var callback Callback
	if err := json.Unmarshal(body, &callback); err != nil {
		return Callback{}, err
	}
	return callback, nil
}
