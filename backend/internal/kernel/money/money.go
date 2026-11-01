// Package money provides currency-safe arithmetic in integer minor units.
package money

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrOverflow         = errors.New("money arithmetic overflow")
	ErrInvalidPrecision = errors.New("money precision must be between zero and nine decimal places")
)

// Money stores an amount in the currency's minor unit (for example, cents).
type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type RoundingMode uint8

const (
	RoundHalfUp RoundingMode = iota
	RoundHalfEven
	RoundDown // toward zero
)

func New(amountMinor int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency is required")
	}
	return Money{AmountMinor: amountMinor, Currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	if (other.AmountMinor > 0 && m.AmountMinor > math.MaxInt64-other.AmountMinor) ||
		(other.AmountMinor < 0 && m.AmountMinor < math.MinInt64-other.AmountMinor) {
		return Money{}, ErrOverflow
	}
	m.AmountMinor += other.AmountMinor
	return m, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.sameCurrency(other); err != nil {
		return Money{}, err
	}
	if other.AmountMinor == math.MinInt64 {
		return Money{}, ErrOverflow
	}
	return m.Add(Money{AmountMinor: -other.AmountMinor, Currency: other.Currency})
}

// Format renders an amount using an explicit minor-unit precision. Callers provide
// the currency's scale so the kernel does not guess currency-specific rules.
func (m Money) Format(decimalPlaces uint8) (string, error) {
	if decimalPlaces > 9 {
		return "", ErrInvalidPrecision
	}
	scale := uint64(1)
	for i := uint8(0); i < decimalPlaces; i++ {
		scale *= 10
	}
	amount := absUint(m.AmountMinor)
	whole, fraction := amount/scale, amount%scale
	sign := ""
	if m.AmountMinor < 0 {
		sign = "-"
	}
	if decimalPlaces == 0 {
		return fmt.Sprintf("%s %s%d", m.Currency, sign, whole), nil
	}
	return fmt.Sprintf("%s %s%d.%0*d", m.Currency, sign, whole, int(decimalPlaces), fraction), nil
}

// Divide returns m divided by divisor, rounded to the nearest minor unit.
func (m Money) Divide(divisor int64, mode RoundingMode) (Money, error) {
	if divisor == 0 {
		return Money{}, ErrDivisionByZero
	}
	if divisor == math.MinInt64 {
		return Money{}, ErrOverflow
	}
	q, r := m.AmountMinor/divisor, m.AmountMinor%divisor
	if r == 0 || mode == RoundDown {
		m.AmountMinor = q
		return m, nil
	}
	// Compare |r| with |divisor|/2 without overflowing.
	ar, ad := absUint(r), absUint(divisor)
	twiceR := ar * 2
	up := twiceR > ad || (twiceR == ad && (mode == RoundHalfUp || (mode == RoundHalfEven && q%2 != 0)))
	if up {
		if (m.AmountMinor < 0) != (divisor < 0) {
			q--
		} else {
			q++
		}
	}
	m.AmountMinor = q
	return m, nil
}

func (m Money) sameCurrency(other Money) error {
	if m.Currency != other.Currency {
		return fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	return nil
}

func absUint(v int64) uint64 {
	if v < 0 {
		return uint64(-(v + 1)) + 1
	}
	return uint64(v)
}
