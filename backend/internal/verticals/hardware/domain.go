package hardware

import (
	"errors"
	"math"
)

var ErrInvalidConversion = errors.New("uom conversion must be finite and positive")

func Convert(quantity, multiplier float64) (float64, error) {
	if quantity < 0 || multiplier <= 0 || math.IsNaN(quantity) || math.IsInf(quantity, 0) || math.IsNaN(multiplier) || math.IsInf(multiplier, 0) {
		return 0, ErrInvalidConversion
	}
	return quantity * multiplier, nil
}

func RoundTrip(quantity, forward, reverse float64) (float64, error) {
	converted, err := Convert(quantity, forward)
	if err != nil {
		return 0, err
	}
	return Convert(converted, reverse)
}
