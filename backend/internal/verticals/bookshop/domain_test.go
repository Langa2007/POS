package bookshop

import "testing"

func TestValidateISBN(t *testing.T) {
	for _, isbn := range []string{"0-306-40615-2", "9780306406157"} {
		if err := ValidateISBN(isbn); err != nil {
			t.Errorf("%s: %v", isbn, err)
		}
	}
	if ValidateISBN("9780306406158") == nil {
		t.Fatal("invalid ISBN accepted")
	}
}
