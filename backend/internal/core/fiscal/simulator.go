package fiscal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// Simulator returns stable provider references for contract and workflow tests. It is not a compliant fiscal adapter.
type Simulator struct{ Fail bool }

func (s Simulator) Submit(ctx context.Context, d Document) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if s.Fail {
		return Result{}, errors.New("simulated fiscal provider outage")
	}
	if d.TenantID == "" || d.ID == "" || d.IdempotencyKey == "" || !(d.Type == "receipt" || d.Type == "invoice" || d.Type == "credit_note") {
		return Result{}, errors.New("invalid fiscal document")
	}
	h := sha256.Sum256([]byte(strings.Join([]string{d.TenantID, d.Type, d.ID, d.IdempotencyKey}, ":")))
	ref := strings.ToUpper(hex.EncodeToString(h[:10]))
	return Result{ControlNumber: "SIM-" + ref, QRCode: "SIM:" + ref, Signature: hex.EncodeToString(h[:])}, nil
}
