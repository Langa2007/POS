package mpesa

import (
	"context"
	"errors"
	"time"
)

var ErrProviderUnavailable = errors.New("payment provider unavailable")

type STKRequest struct {
	BusinessShortCode string
	Phone             string
	AmountMinor       int64
	AccountReference  string
	Description       string
}

type STKResponse struct {
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
}

type Callback struct {
	CheckoutRequestID string
	ResultCode        int
	ResultDescription string
	ReceiptNumber     string
	AmountMinor       int64
	Phone             string
	AccountReference  string
}

type Provider interface {
	InitiateSTK(context.Context, STKRequest) (STKResponse, error)
	QueryStatus(context.Context, string) (string, error)
	Reverse(context.Context, string, int64) error
	ParseCallback([]byte) (Callback, error)
}

type TokenCache struct {
	AccessToken string
	ExpiresAt   time.Time
}

func (c TokenCache) Valid(now time.Time) bool {
	return c.AccessToken != "" && now.Before(c.ExpiresAt.Add(-30*time.Second))
}
