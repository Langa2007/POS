package payments

import "testing"

func TestPaymentStateMachine(t *testing.T) {
	if err := Transition(Pending, Captured); err != nil {
		t.Fatal(err)
	}
	if err := Transition(Failed, Captured); err != ErrInvalidTransition {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestSplitTenderAllocation(t *testing.T) {
	err := ValidateAllocation(1000, []Tender{{Method: "cash", AmountMinor: 600}, {Method: "mpesa_stk", AmountMinor: 400}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAllocation(1000, []Tender{{Method: "cash", AmountMinor: 900}}); err == nil {
		t.Fatal("expected incomplete allocation")
	}
	amount, err := ValidatePartialAllocation(1000, []Tender{{Method: "cash", AmountMinor: 600}})
	if err != nil || amount != 600 {
		t.Fatalf("partial amount=%d err=%v", amount, err)
	}
}

func TestLedgerBalance(t *testing.T) {
	if err := ValidateBalanced([]Posting{{Account: "cash", Debit: 1000}, {Account: "sales", Credit: 1000}}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBalanced([]Posting{{Account: "cash", Debit: 1000}, {Account: "sales", Credit: 900}}); err == nil {
		t.Fatal("expected imbalance")
	}
}
