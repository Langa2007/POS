package phase8

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"sort"
	"time"
)

var (
	ErrAmbiguousMatch = errors.New("statement line has multiple equally strong matches")
	ErrPayoutApproval = errors.New("payout requires a different approver")
	ErrInvalidBilling = errors.New("invalid billing period")
)

type PaymentRequest struct {
	AmountMinor    int64
	Currency       string
	IdempotencyKey string
	ReturnURL      string
}

type PaymentResult struct {
	ProviderReference string
	Status            string
}

type CardBankProvider interface {
	Name() string
	CreatePayment(PaymentRequest) (PaymentResult, error)
	Refund(providerReference string, amountMinor int64) error
	VerifyWebhook(payload []byte, signature string) bool
}

func WebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyWebhook(secret string, payload []byte, signature string) bool {
	expected := WebhookSignature(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

type StatementLine struct {
	ID                string
	ExternalReference string
	AmountMinor       int64
	Description       string
}

type Candidate struct {
	TargetID    string
	AmountMinor int64
	Reference   string
}

type Match struct {
	LineID     string
	TargetID   string
	Confidence float64
	Automatic  bool
}

func AutoMatch(lines []StatementLine, candidates []Candidate) ([]Match, []string) {
	matches := make([]Match, 0)
	exceptions := make([]string, 0)
	for _, line := range lines {
		best := make([]Candidate, 0)
		for _, candidate := range candidates {
			if candidate.AmountMinor != line.AmountMinor {
				continue
			}
			if line.ExternalReference != "" && candidate.Reference == line.ExternalReference {
				best = append([]Candidate{candidate}, best...)
			} else if line.ExternalReference == "" || candidate.Reference == "" {
				best = append(best, candidate)
			}
		}
		if len(best) == 1 {
			confidence := 0.8
			if line.ExternalReference != "" && best[0].Reference == line.ExternalReference {
				confidence = 1
			}
			matches = append(matches, Match{LineID: line.ID, TargetID: best[0].TargetID, Confidence: confidence, Automatic: confidence == 1})
		} else {
			exceptions = append(exceptions, line.ID)
		}
	}
	sort.Strings(exceptions)
	return matches, exceptions
}

type Payout struct {
	ID          string
	AmountMinor int64
	RequestedBy string
	ApprovedBy  string
	Status      string
}

func (p *Payout) Approve(userID string) error {
	if userID == "" || userID == p.RequestedBy {
		return ErrPayoutApproval
	}
	if p.Status != "pending" {
		return ErrPayoutApproval
	}
	p.ApprovedBy = userID
	p.Status = "approved"
	return nil
}

func ProratedCharge(monthlyMinor int64, periodStart, periodEnd, activeFrom, activeTo time.Time) (int64, error) {
	if monthlyMinor < 0 || !periodEnd.After(periodStart) || !activeTo.After(activeFrom) {
		return 0, ErrInvalidBilling
	}
	days := periodEnd.Sub(periodStart).Hours() / 24
	start := activeFrom
	if start.Before(periodStart) {
		start = periodStart
	}
	end := activeTo
	if end.After(periodEnd) {
		end = periodEnd
	}
	activeDays := end.Sub(start).Hours() / 24
	if activeDays < 0 {
		activeDays = 0
	}
	return int64(math.Round(float64(monthlyMinor) * activeDays / days)), nil
}
