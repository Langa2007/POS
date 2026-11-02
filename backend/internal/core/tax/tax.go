// Package tax calculates taxes using integer minor units.
package tax

import (
	"errors"
	"math/big"
)

var (
	ErrInvalidRate       = errors.New("tax rate must be between zero and 10000 basis points")
	ErrInvalidLine       = errors.New("tax line amount must be non-negative")
	ErrInvalidRateConfig = errors.New("tax rate is missing")
	ErrOverflow          = errors.New("tax result exceeds int64")
)

type Inclusion uint8

const (
	Exclusive Inclusion = iota
	Inclusive
)

type Rounding uint8

const (
	HalfUp Rounding = iota
	HalfEven
	Down
)

// Rate is expressed in basis points: 1600 is 16%.
type Rate struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	BasisPoint int64  `json:"basis_points"`
	Exempt     bool   `json:"exempt"`
}

type Line struct {
	NetMinor int64  `json:"net_minor"`
	RateCode string `json:"rate_code"`
}

type Config struct {
	Rates            map[string]Rate `json:"rates"`
	Inclusion        Inclusion       `json:"inclusion"`
	Rounding         Rounding        `json:"rounding"`
	RoundPerDocument bool            `json:"round_per_document"`
}

type LineResult struct {
	NetMinor   int64 `json:"net_minor"`
	TaxMinor   int64 `json:"tax_minor"`
	GrossMinor int64 `json:"gross_minor"`
}

type Result struct {
	NetMinor   int64        `json:"net_minor"`
	TaxMinor   int64        `json:"tax_minor"`
	GrossMinor int64        `json:"gross_minor"`
	Lines      []LineResult `json:"lines"`
}

func (c Config) Calculate(lines []Line) (Result, error) {
	result := Result{Lines: make([]LineResult, len(lines))}
	if c.Inclusion != Exclusive && c.Inclusion != Inclusive {
		return Result{}, errors.New("invalid tax inclusion")
	}
	if c.Rounding != HalfUp && c.Rounding != HalfEven && c.Rounding != Down {
		return Result{}, errors.New("invalid tax rounding")
	}

	for i, line := range lines {
		if line.NetMinor < 0 {
			return Result{}, ErrInvalidLine
		}
		rate, ok := c.Rates[line.RateCode]
		if !ok {
			return Result{}, ErrInvalidRateConfig
		}
		if rate.BasisPoint < 0 || rate.BasisPoint > 10000 {
			return Result{}, ErrInvalidRate
		}
		net, tax := line.NetMinor, int64(0)
		var err error
		if !rate.Exempt && rate.BasisPoint != 0 {
			if c.Inclusion == Inclusive {
				tax, net, err = inclusiveTax(line.NetMinor, rate.BasisPoint, c.Rounding)
				if err != nil {
					return Result{}, err
				}
			} else {
				tax, err = ratio(line.NetMinor, rate.BasisPoint, 10000, c.Rounding)
				if err != nil {
					return Result{}, err
				}
			}
		}
		gross, ok := add(net, tax)
		if !ok {
			return Result{}, ErrOverflow
		}
		result.Lines[i] = LineResult{NetMinor: net, TaxMinor: tax, GrossMinor: gross}
	}
	if c.RoundPerDocument {
		var err error
		result, err = roundDocument(result, lines, c)
		if err != nil {
			return Result{}, err
		}
	}
	for _, line := range result.Lines {
		var ok bool
		if result.NetMinor, ok = add(result.NetMinor, line.NetMinor); !ok {
			return Result{}, ErrOverflow
		}
		if result.TaxMinor, ok = add(result.TaxMinor, line.TaxMinor); !ok {
			return Result{}, ErrOverflow
		}
		if result.GrossMinor, ok = add(result.GrossMinor, line.GrossMinor); !ok {
			return Result{}, ErrOverflow
		}
	}
	return result, nil
}

func inclusiveTax(gross, rate int64, mode Rounding) (int64, int64, error) {
	tax, err := ratio(gross, rate, 10000+rate, mode)
	if err != nil {
		return 0, 0, err
	}
	return tax, gross - tax, nil
}

func ratio(value, numerator, denominator int64, mode Rounding) (int64, error) {
	n := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	d := big.NewInt(denominator)
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(n, d, r)
	if mode == Down || r.Sign() == 0 {
		if !q.IsInt64() {
			return 0, ErrOverflow
		}
		return q.Int64(), nil
	}
	twice := new(big.Int).Lsh(new(big.Int).Abs(r), 1)
	absD := new(big.Int).Abs(d)
	up := twice.Cmp(absD) > 0 || (twice.Cmp(absD) == 0 && (mode == HalfUp || q.Bit(0) == 1))
	if up {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsInt64() {
		return 0, ErrOverflow
	}
	return q.Int64(), nil
}

func roundDocument(result Result, lines []Line, config Config) (Result, error) {
	lastByRate := make(map[string]int)
	for i, line := range lines {
		lastByRate[line.RateCode] = i
	}
	taxByRate := make(map[string]int64)
	for _, line := range lines {
		rate := config.Rates[line.RateCode]
		if rate.Exempt || rate.BasisPoint == 0 {
			continue
		}
		base := line.NetMinor
		if config.Inclusion == Inclusive {
			base = line.NetMinor
		}
		var ok bool
		if taxByRate[line.RateCode], ok = add(taxByRate[line.RateCode], base); !ok {
			return Result{}, ErrOverflow
		}
	}
	for code, base := range taxByRate {
		rate := config.Rates[code]
		documentTax, err := ratio(base, rate.BasisPoint, 10000, config.Rounding)
		if err != nil {
			return Result{}, err
		}
		if config.Inclusion == Inclusive {
			documentTax, err = ratio(base, rate.BasisPoint, 10000+rate.BasisPoint, config.Rounding)
			if err != nil {
				return Result{}, err
			}
		}
		var lineTax int64
		for i, line := range lines {
			if line.RateCode == code {
				var ok bool
				if lineTax, ok = add(lineTax, result.Lines[i].TaxMinor); !ok {
					return Result{}, ErrOverflow
				}
			}
		}
		delta, ok := subtract(documentTax, lineTax)
		if !ok {
			return Result{}, ErrOverflow
		}
		i := lastByRate[code]
		result.Lines[i].TaxMinor, ok = add(result.Lines[i].TaxMinor, delta)
		if !ok || result.Lines[i].TaxMinor < 0 {
			return Result{}, ErrOverflow
		}
		if config.Inclusion == Inclusive {
			result.Lines[i].NetMinor, ok = subtract(result.Lines[i].NetMinor, delta)
			if !ok || result.Lines[i].NetMinor < 0 {
				return Result{}, ErrOverflow
			}
		}
		result.Lines[i].GrossMinor, ok = add(result.Lines[i].NetMinor, result.Lines[i].TaxMinor)
		if !ok {
			return Result{}, ErrOverflow
		}
	}
	return result, nil
}

func add(a, b int64) (int64, bool) {
	if b > 0 && a > int64(^uint64(0)>>1)-b || b < 0 && a < -int64(^uint64(0)>>1)-1-b {
		return 0, false
	}
	return a + b, true
}
func subtract(a, b int64) (int64, bool) {
	if b == -int64(^uint64(0)>>1)-1 {
		if a >= 0 {
			return 0, false
		}
		return a - b, true
	}
	return add(a, -b)
}
