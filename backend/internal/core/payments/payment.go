package payments

import "errors"

type Status string

const (
	Pending    Status = "pending"
	Authorized Status = "authorized"
	Captured   Status = "captured"
	Failed     Status = "failed"
	Cancelled  Status = "cancelled"
	Refunded   Status = "refunded"
)

var ErrInvalidTransition = errors.New("invalid payment state transition")

func CanTransition(from, to Status) bool {
	switch from {
	case Pending:
		return to == Authorized || to == Captured || to == Failed || to == Cancelled
	case Authorized:
		return to == Captured || to == Failed || to == Cancelled
	case Captured:
		return to == Refunded
	default:
		return false
	}
}

func Transition(from, to Status) error {
	if !CanTransition(from, to) {
		return ErrInvalidTransition
	}
	return nil
}

type Tender struct {
	Method      string `json:"method"`
	AmountMinor int64  `json:"amount_minor"`
	Reference   string `json:"reference,omitempty"`
}

func ValidateAllocation(total int64, tenders []Tender) error {
	if total <= 0 || len(tenders) == 0 {
		return errors.New("a positive total and at least one tender are required")
	}
	var allocated int64
	for _, tender := range tenders {
		if tender.AmountMinor <= 0 {
			return errors.New("tender amounts must be positive")
		}
		switch tender.Method {
		case "cash", "mpesa_stk", "mpesa_c2b", "credit":
		default:
			return errors.New("unsupported tender method")
		}
		allocated += tender.AmountMinor
	}
	if allocated != total {
		return errors.New("tenders must exactly equal the sale total")
	}
	return nil
}

type Posting struct {
	Account string
	Debit   int64
	Credit  int64
}

func ValidateBalanced(postings []Posting) error {
	var debit, credit int64
	for _, posting := range postings {
		if posting.Debit < 0 || posting.Credit < 0 || (posting.Debit == 0) == (posting.Credit == 0) {
			return errors.New("each posting must have exactly one non-negative side")
		}
		debit += posting.Debit
		credit += posting.Credit
	}
	if len(postings) == 0 || debit != credit {
		return errors.New("ledger postings are not balanced")
	}
	return nil
}
