package membership

import "time"

type BillingRun struct {
	Key       string
	Processed map[string]time.Time
}

func (r *BillingRun) Process(memberID string, at time.Time) bool {
	if r.Processed == nil {
		r.Processed = make(map[string]time.Time)
	}
	if _, exists := r.Processed[memberID]; exists {
		return false
	}
	r.Processed[memberID] = at
	return true
}
