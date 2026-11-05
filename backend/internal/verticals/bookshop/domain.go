package bookshop

import "errors"

var ErrInvalidISBN = errors.New("invalid ISBN")

func NormalizeISBN(value string) string {
	out := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			out = append(out, value[i])
		} else if value[i] == 'X' || value[i] == 'x' {
			out = append(out, 'X')
		}
	}
	return string(out)
}

func ValidateISBN(value string) error {
	isbn := NormalizeISBN(value)
	switch len(isbn) {
	case 10:
		sum := 0
		for i := 0; i < 10; i++ {
			digit := 0
			if isbn[i] == 'X' {
				if i != 9 {
					return ErrInvalidISBN
				}
				digit = 10
			} else if isbn[i] >= '0' && isbn[i] <= '9' {
				digit = int(isbn[i] - '0')
			} else {
				return ErrInvalidISBN
			}
			sum += (10 - i) * digit
		}
		if sum%11 == 0 {
			return nil
		}
	case 13:
		sum := 0
		for i := 0; i < 13; i++ {
			if isbn[i] < '0' || isbn[i] > '9' {
				return ErrInvalidISBN
			}
			weight := 1
			if i%2 == 1 {
				weight = 3
			}
			sum += int(isbn[i]-'0') * weight
		}
		if sum%10 == 0 {
			return nil
		}
	}
	return ErrInvalidISBN
}

func ConsignmentPayable(salesMinor int64, supplierRate float64) (int64, error) {
	if salesMinor < 0 || supplierRate < 0 || supplierRate > 1 {
		return 0, ErrInvalidISBN
	}
	return int64(float64(salesMinor) * supplierRate), nil
}
