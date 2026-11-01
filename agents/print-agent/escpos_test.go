package printagent

import "testing"

func TestEncodeReceipt(t *testing.T) {
	data, err := EncodeReceipt(Receipt{Header: "SHOP", Lines: []string{"Milk 100"}, Total: "100 KES"})
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || data[0] != 0x1b {
		t.Fatalf("unexpected ESC/POS output: %v", data)
	}
}
