package devices

import (
	"errors"
	"strconv"
	"strings"
)

type Barcode struct {
	Raw        string
	SKU        string
	Quantity   int64
	PriceMinor int64
	Kind       string
}

func ParseBarcode(raw string) (Barcode, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return Barcode{}, errors.New("barcode is required")
	}
	if len(value) == 13 && isDigits(value) && validEAN13(value) {
		return Barcode{Raw: value, SKU: value, Quantity: 1, Kind: "ean13"}, nil
	}
	if strings.HasPrefix(value, "21") && len(value) == 13 && isDigits(value) {
		weight, err := strconv.ParseInt(value[7:12], 10, 64)
		if err != nil || weight <= 0 {
			return Barcode{}, errors.New("invalid weight barcode")
		}
		return Barcode{Raw: value, SKU: value[2:7], Quantity: weight, Kind: "weight_embedded"}, nil
	}
	if strings.HasPrefix(value, "02") && len(value) == 13 && isDigits(value) {
		price, err := strconv.ParseInt(value[7:12], 10, 64)
		if err != nil || price <= 0 {
			return Barcode{}, errors.New("invalid price barcode")
		}
		return Barcode{Raw: value, SKU: value[2:7], Quantity: 1, PriceMinor: price, Kind: "price_embedded"}, nil
	}
	if len(value) == 10 || len(value) == 13 {
		return Barcode{Raw: value, SKU: value, Quantity: 1, Kind: "isbn"}, nil
	}
	return Barcode{Raw: value, SKU: value, Quantity: 1, Kind: "sku"}, nil
}

func isDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func validEAN13(value string) bool {
	sum := 0
	for index, char := range value[:12] {
		digit := int(char - '0')
		if index%2 == 1 {
			sum += digit * 3
		} else {
			sum += digit
		}
	}
	check := (10 - sum%10) % 10
	return check == int(value[12]-'0')
}
