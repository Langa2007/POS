package tax

import (
	"testing"
)

func TestExclusiveTax(t *testing.T) {
	got, err := (Config{
		Rates: map[string]Rate{"vat": {Code: "vat", BasisPoint: 1600}},
	}).Calculate([]Line{{NetMinor: 1000, RateCode: "vat"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.NetMinor != 1000 || got.TaxMinor != 160 || got.GrossMinor != 1160 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestInclusiveTaxUsesGrossAsInput(t *testing.T) {
	got, err := (Config{
		Rates:     map[string]Rate{"vat": {Code: "vat", BasisPoint: 1600}},
		Inclusion: Inclusive,
	}).Calculate([]Line{{NetMinor: 1160, RateCode: "vat"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.NetMinor != 1000 || got.TaxMinor != 160 || got.GrossMinor != 1160 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestExemptAndHalfEven(t *testing.T) {
	got, err := (Config{
		Rates:    map[string]Rate{"zero": {Code: "zero", Exempt: true}, "vat": {Code: "vat", BasisPoint: 5000}},
		Rounding: HalfEven,
	}).Calculate([]Line{{NetMinor: 1, RateCode: "zero"}, {NetMinor: 1, RateCode: "vat"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TaxMinor != 0 || got.GrossMinor != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestTaxRequiresKnownRateAndRejectsNegativeLines(t *testing.T) {
	if _, err := (Config{Rates: map[string]Rate{}}).Calculate([]Line{{NetMinor: 1, RateCode: "missing"}}); err != ErrInvalidRateConfig {
		t.Fatalf("expected missing rate, got %v", err)
	}
	if _, err := (Config{Rates: map[string]Rate{"vat": {BasisPoint: 1600}}}).Calculate([]Line{{NetMinor: -1, RateCode: "vat"}}); err != ErrInvalidLine {
		t.Fatalf("expected invalid line, got %v", err)
	}
}

func TestTaxRejectsOverflowInsteadOfWrapping(t *testing.T) {
	max := int64(^uint64(0) >> 1)
	_, err := (Config{Rates: map[string]Rate{"full": {Code: "full", BasisPoint: 10000}}}).Calculate([]Line{{NetMinor: max, RateCode: "full"}})
	if err != ErrOverflow {
		t.Fatalf("expected gross overflow, got %v", err)
	}
	_, err = (Config{Rates: map[string]Rate{"zero": {Code: "zero"}}}).Calculate([]Line{{NetMinor: max, RateCode: "zero"}, {NetMinor: 1, RateCode: "zero"}})
	if err != ErrOverflow {
		t.Fatalf("expected basket overflow, got %v", err)
	}
}

func TestDocumentRoundingPreservesTotals(t *testing.T) {
	cfg := Config{Rates: map[string]Rate{"vat": {Code: "vat", BasisPoint: 5000}}, RoundPerDocument: true, Rounding: HalfUp}
	got, err := cfg.Calculate([]Line{{NetMinor: 1, RateCode: "vat"}, {NetMinor: 1, RateCode: "vat"}, {NetMinor: 1, RateCode: "vat"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.NetMinor+got.TaxMinor != got.GrossMinor || got.TaxMinor != 2 {
		t.Fatalf("document totals do not reconcile: %+v", got)
	}
}
