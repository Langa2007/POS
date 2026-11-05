package dealership

import "errors"

var ErrInvalidInstallment = errors.New("invalid installment plan")

type Installment struct {
	DueMinor int64
	DueOn    string
}

func Schedule(principalMinor int64, count int, feeRate float64) ([]Installment, error) {
	if principalMinor < 0 || count <= 0 || feeRate < 0 {
		return nil, ErrInvalidInstallment
	}
	total := principalMinor + int64(float64(principalMinor)*feeRate)
	base, remainder := total/int64(count), total%int64(count)
	out := make([]Installment, count)
	for i := range out {
		out[i].DueMinor = base
		if int64(i) < remainder {
			out[i].DueMinor++
		}
	}
	return out, nil
}
