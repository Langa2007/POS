package fuelstation

import "math"

func ReconcileMeter(open, close, deliveries, sales float64, tolerance float64) bool {
	if open < 0 || close < open || deliveries < 0 || sales < 0 || tolerance < 0 {
		return false
	}
	return math.Abs((close-open)+deliveries-sales) <= tolerance
}
