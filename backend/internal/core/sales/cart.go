package sales

import "errors"

var ErrInvalidQuantity = errors.New("quantity must be positive")

type Line struct {
	Quantity       int64
	UnitPriceMinor int64
	DiscountMinor  int64
}

func (l Line) TotalMinor() (int64, error) {
	if l.Quantity <= 0 || l.UnitPriceMinor < 0 || l.DiscountMinor < 0 {
		return 0, ErrInvalidQuantity
	}
	total := l.Quantity*l.UnitPriceMinor - l.DiscountMinor
	if total < 0 {
		return 0, errors.New("discount cannot exceed line total")
	}
	return total, nil
}

func Total(lines []Line) (int64, error) {
	var total int64
	for _, line := range lines {
		value, err := line.TotalMinor()
		if err != nil {
			return 0, err
		}
		total += value
	}
	return total, nil
}

func ChangeDue(total, received int64) (int64, error) {
	if total < 0 || received < total {
		return 0, errors.New("received amount is less than total")
	}
	return received - total, nil
}

func ReturnAmount(unitPrice, quantity, alreadyReturned int64) (int64, error) {
	if unitPrice < 0 || quantity <= 0 || alreadyReturned < 0 || alreadyReturned > quantity {
		return 0, errors.New("invalid return quantity")
	}
	return (quantity - alreadyReturned) * unitPrice, nil
}

func ShiftVariance(openingFloat, cashSales, paidIn, paidOut, closingCash int64) (int64, error) {
	if openingFloat < 0 || cashSales < 0 || paidIn < 0 || paidOut < 0 || closingCash < 0 {
		return 0, errors.New("cash values must be non-negative")
	}
	return closingCash - (openingFloat + cashSales + paidIn - paidOut), nil
}
