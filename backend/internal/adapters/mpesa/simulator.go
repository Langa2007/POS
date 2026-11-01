package mpesa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Simulator struct {
	mu            sync.Mutex
	Mode          string
	CallbackDelay time.Duration
	Requests      []STKRequest
	Callbacks     []Callback
}

func (s *Simulator) InitiateSTK(_ context.Context, request STKRequest) (STKResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Requests = append(s.Requests, request)
	if s.Mode == "provider_error" {
		return STKResponse{}, ErrProviderUnavailable
	}
	if s.Mode == "slow" {
		// A canceled/deadline context is modeled by callers; avoid blocking tests here.
		return STKResponse{}, ErrProviderUnavailable
	}
	return STKResponse{CheckoutRequestID: fmt.Sprintf("ws_CO_%d", len(s.Requests)), MerchantRequestID: "simulator", ResponseCode: "0", ResponseDescription: "Success"}, nil
}

// DeliverCallback models callback delay, loss, and duplicates for resilience tests.
// The returned sequence intentionally preserves duplicate and out-of-order delivery.
func (s *Simulator) DeliverCallback(callback Callback) []Callback {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Mode == "lost_callback" {
		return nil
	}
	s.Callbacks = append(s.Callbacks, callback)
	result := []Callback{callback}
	if s.Mode == "duplicate_callback" {
		result = append(result, callback)
	}
	return result
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

func (s *Simulator) B2C(_ context.Context, _ B2CRequest) (B2CResponse, error) {
	if s.Mode == "provider_error" {
		return B2CResponse{}, ErrProviderUnavailable
	}
	return B2CResponse{ResponseCode: "0", ResponseDescription: "Accepted"}, nil
}

func (s *Simulator) GenerateQR(_ context.Context, payload string) (QRResponse, error) {
	if payload == "" {
		return QRResponse{}, errors.New("QR payload is required")
	}
	if s.Mode == "provider_error" {
		return QRResponse{}, ErrProviderUnavailable
	}
	return QRResponse{QRCode: "SIMULATOR:" + payload}, nil
}

func (s *Simulator) ParseCallback(body []byte) (Callback, error) {
	var callback Callback
	if err := json.Unmarshal(body, &callback); err != nil {
		return Callback{}, err
	}
	return callback, nil
}
