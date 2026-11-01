package sync

import (
	"errors"
	"time"
)

type OfflinePolicy struct {
	MaxDuration     time.Duration
	MaxSaleMinor    int64
	AllowCredit     bool
	AllowControlled bool
}

func (p OfflinePolicy) Validate() error {
	if p.MaxDuration <= 0 {
		return errors.New("max offline duration must be positive")
	}
	if p.MaxSaleMinor <= 0 {
		return errors.New("max offline sale amount must be positive")
	}
	return nil
}

func (p OfflinePolicy) AllowsSale(offlineSince time.Time, amountMinor int64, credit, controlled bool, now time.Time) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if offlineSince.IsZero() || now.Before(offlineSince) {
		return errors.New("offline start time is invalid")
	}
	if now.Sub(offlineSince) > p.MaxDuration {
		return errors.New("offline duration limit exceeded")
	}
	if amountMinor <= 0 || amountMinor > p.MaxSaleMinor {
		return errors.New("offline sale amount limit exceeded")
	}
	if credit && !p.AllowCredit {
		return errors.New("credit sales are unavailable offline")
	}
	if controlled && !p.AllowControlled {
		return errors.New("controlled items are unavailable offline")
	}
	return nil
}
