package money

import (
	"errors"
	"math"
	"testing"
)

func TestAddSubAndCurrency(t *testing.T) {
	a := Money{AmountMinor: 125, Currency: "KES"}
	b := Money{AmountMinor: 75, Currency: "KES"}
	got, err := a.Add(b)
	if err != nil || got.AmountMinor != 200 {
		t.Fatalf("Add()=(%+v,%v)", got, err)
	}
	got, err = got.Sub(b)
	if err != nil || got.AmountMinor != 125 {
		t.Fatalf("Sub()=(%+v,%v)", got, err)
	}
	if _, err = a.Add(Money{AmountMinor: 1, Currency: "USD"}); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected currency mismatch, got %v", err)
	}
}
func TestDivideRoundingModes(t *testing.T) {
	cases := []struct {
		amount, divisor, want int64
		mode                  RoundingMode
	}{{5, 2, 3, RoundHalfUp}, {5, 2, 2, RoundHalfEven}, {7, 2, 4, RoundHalfEven}, {-5, 2, -3, RoundHalfUp}, {5, 2, 2, RoundDown}, {5, -2, -3, RoundHalfUp}}
	for _, tc := range cases {
		got, err := (Money{AmountMinor: tc.amount, Currency: "KES"}).Divide(tc.divisor, tc.mode)
		if err != nil || got.AmountMinor != tc.want {
			t.Errorf("Divide(%d,%d,%d)=(%d,%v), want %d", tc.amount, tc.divisor, tc.mode, got.AmountMinor, err, tc.want)
		}
	}
	if _, err := (Money{Currency: "KES"}).Divide(0, RoundHalfUp); !errors.Is(err, ErrDivisionByZero) {
		t.Fatalf("expected div by zero, got %v", err)
	}
}
func TestAddOverflow(t *testing.T) {
	if _, err := (Money{AmountMinor: math.MaxInt64, Currency: "KES"}).Add(Money{AmountMinor: 1, Currency: "KES"}); !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected overflow, got %v", err)
	}
}

func TestFormatWithExplicitCurrencyPrecision(t *testing.T) {
	got, err := (Money{AmountMinor: 12345, Currency: "KES"}).Format(2)
	if err != nil || got != "KES 123.45" {
		t.Fatalf("Format()=(%q,%v)", got, err)
	}
	got, err = (Money{AmountMinor: -123, Currency: "JPY"}).Format(0)
	if err != nil || got != "JPY -123" {
		t.Fatalf("Format()=(%q,%v)", got, err)
	}
	if _, err = (Money{Currency: "KES"}).Format(10); !errors.Is(err, ErrInvalidPrecision) {
		t.Fatalf("expected precision error, got %v", err)
	}
}
