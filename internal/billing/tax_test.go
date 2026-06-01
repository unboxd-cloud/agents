package billing

import "testing"

func TestApplyTax_VAT(t *testing.T) {
	res := ApplyTax(100, "EUR", TaxRulesFor("EU-DE"))
	if res.TaxTotal != 19.00 {
		t.Errorf("tax: want 19.00, got %v", res.TaxTotal)
	}
	if res.Gross != 119.00 {
		t.Errorf("gross: want 119.00, got %v", res.Gross)
	}
}

func TestApplyTax_ReverseCharge(t *testing.T) {
	rules := []TaxRule{{Jurisdiction: "EU-DE", Name: "VAT", Kind: TaxVAT, Rate: 0.19, ReverseCharge: true}}
	res := ApplyTax(100, "EUR", rules)
	if res.TaxTotal != 0 {
		t.Errorf("reverse-charge collects 0, got %v", res.TaxTotal)
	}
	if res.Gross != 100 {
		t.Errorf("reverse-charge gross stays net, got %v", res.Gross)
	}
	if len(res.TaxLines) != 1 || res.TaxLines[0].Amount != 19.00 {
		t.Errorf("reverse-charge should report notional 19.00: %+v", res.TaxLines)
	}
}

func TestApplyTax_Compound(t *testing.T) {
	// CA-QC: GST 5% on 100 = 5.00; QST 9.975% on (100+5) = 10.47; total 15.47.
	res := ApplyTax(100, "CAD", TaxRulesFor("CA-QC"))
	if res.TaxTotal != 15.47 {
		t.Errorf("compound tax: want 15.47, got %v", res.TaxTotal)
	}
	if res.Gross != 115.47 {
		t.Errorf("compound gross: want 115.47, got %v", res.Gross)
	}
}

func TestApplyTax_NoRules(t *testing.T) {
	res := ApplyTax(100, "USD", TaxRulesFor("ZZ"))
	if res.Gross != 100 || res.TaxTotal != 0 {
		t.Errorf("no rules should be tax-free: %+v", res)
	}
}
