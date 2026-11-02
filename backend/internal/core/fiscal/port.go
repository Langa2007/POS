// Package fiscal defines the provider boundary for jurisdictional fiscal devices and services.
package fiscal

import "context"

type Document struct {
	TenantID       string
	Type           string
	ID             string
	IdempotencyKey string
	Payload        []byte
}

type Result struct {
	ControlNumber string `json:"control_number"`
	QRCode        string `json:"qr_code"`
	Signature     string `json:"signature"`
}

type Submitter interface {
	Submit(context.Context, Document) (Result, error)
}
