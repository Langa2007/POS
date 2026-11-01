package devices

import "testing"

func TestParseBarcode(t *testing.T) {
	barcode, err := ParseBarcode("4006381333931")
	if err != nil {
		t.Fatal(err)
	}
	if barcode.Kind != "ean13" || barcode.Quantity != 1 {
		t.Fatalf("barcode=%+v", barcode)
	}
}

func TestParseWeightBarcode(t *testing.T) {
	barcode, err := ParseBarcode("2112345005008")
	if err != nil {
		t.Fatal(err)
	}
	if barcode.Kind != "weight_embedded" || barcode.Quantity != 500 {
		t.Fatalf("barcode=%+v", barcode)
	}
}
