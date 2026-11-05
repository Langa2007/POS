package garage

import "errors"

var ErrInvalidJob = errors.New("job total cannot be negative")

type JobLine struct {
	Kind        string
	AmountMinor int64
}

func Total(lines []JobLine) (int64, error) {
	var total int64
	for _, line := range lines {
		if line.AmountMinor < 0 {
			return 0, ErrInvalidJob
		}
		total += line.AmountMinor
	}
	return total, nil
}
