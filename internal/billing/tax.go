package billing

// Tax follows the same "policy as data" tenet as rating: jurisdictions and
// rates are data, applied by one engine. Tax is computed on the customer-facing
// amount (after partner settlement), so it composes with Rate and Settle
// without duplicating any of them.

// TaxKind classifies a tax for reporting and remittance.
type TaxKind string

const (
	TaxNone     TaxKind = "none"
	TaxVAT      TaxKind = "vat"
	TaxGST      TaxKind = "gst"
	TaxSalesTax TaxKind = "sales_tax"
)

// TaxRule describes a tax levied within a jurisdiction (law of the land).
type TaxRule struct {
	Jurisdiction  string   `json:"jurisdiction"` // e.g. "EU-DE", "GB", "US-CA", "IN", "AE"
	Name          string   `json:"name"`         // "VAT", "GST", "CA Sales Tax"
	Kind          TaxKind  `json:"kind"`
	Rate          float64  `json:"rate"`                 // 0.20 = 20%
	ReverseCharge bool     `json:"reverseCharge"`        // B2B: customer self-accounts; supplier collects 0
	CompoundOn    []string `json:"compoundOn,omitempty"` // names of taxes this is levied on top of
}

// TaxLine is one applied tax on an invoice.
type TaxLine struct {
	Name          string  `json:"name"`
	Jurisdiction  string  `json:"jurisdiction"`
	Kind          TaxKind `json:"kind"`
	Rate          float64 `json:"rate"`
	Taxable       float64 `json:"taxable"`
	Amount        float64 `json:"amount"` // notional amount (collected unless reverse-charge)
	ReverseCharge bool    `json:"reverseCharge"`
}

// TaxResult is the taxed view of a customer-facing amount.
type TaxResult struct {
	Net      float64   `json:"net"`
	TaxLines []TaxLine `json:"taxLines"`
	TaxTotal float64   `json:"taxTotal"` // collected tax (excludes reverse-charge)
	Gross    float64   `json:"gross"`
	Currency string    `json:"currency"`
}

// ApplyTax applies tax rules to a net amount. Compounding taxes are levied on
// the net plus the named prior taxes (e.g. QST on top of GST). Reverse-charge
// taxes are reported (rate + notional amount) but not collected, so they do not
// add to the gross.
func ApplyTax(net float64, currency string, rules []TaxRule) TaxResult {
	res := TaxResult{Net: round2(net), Currency: currency, Gross: round2(net)}
	applied := map[string]float64{} // tax name -> notional amount
	for _, r := range rules {
		taxable := net
		for _, base := range r.CompoundOn {
			taxable += applied[base]
		}
		amount := round2(taxable * r.Rate)
		res.TaxLines = append(res.TaxLines, TaxLine{
			Name:          r.Name,
			Jurisdiction:  r.Jurisdiction,
			Kind:          r.Kind,
			Rate:          r.Rate,
			Taxable:       round2(taxable),
			Amount:        amount,
			ReverseCharge: r.ReverseCharge,
		})
		applied[r.Name] = amount
		if !r.ReverseCharge {
			res.TaxTotal += amount
			res.Gross += amount
		}
	}
	res.TaxTotal = round2(res.TaxTotal)
	res.Gross = round2(res.Gross)
	return res
}

// TaxRulesFor returns representative tax rules for a jurisdiction. Real
// deployments source these from a tax engine (e.g. via a provider adapter);
// this sample table keeps rates as swappable data.
func TaxRulesFor(jurisdiction string) []TaxRule {
	switch jurisdiction {
	case "EU-DE":
		return []TaxRule{{Jurisdiction: "EU-DE", Name: "VAT", Kind: TaxVAT, Rate: 0.19}}
	case "EU-FR":
		return []TaxRule{{Jurisdiction: "EU-FR", Name: "VAT", Kind: TaxVAT, Rate: 0.20}}
	case "GB":
		return []TaxRule{{Jurisdiction: "GB", Name: "VAT", Kind: TaxVAT, Rate: 0.20}}
	case "IN":
		return []TaxRule{{Jurisdiction: "IN", Name: "GST", Kind: TaxGST, Rate: 0.18}}
	case "AE":
		return []TaxRule{{Jurisdiction: "AE", Name: "VAT", Kind: TaxVAT, Rate: 0.05}}
	case "US-CA":
		return []TaxRule{{Jurisdiction: "US-CA", Name: "CA Sales Tax", Kind: TaxSalesTax, Rate: 0.0725}}
	case "CA-QC":
		return []TaxRule{
			{Jurisdiction: "CA", Name: "GST", Kind: TaxGST, Rate: 0.05},
			{Jurisdiction: "CA-QC", Name: "QST", Kind: TaxSalesTax, Rate: 0.09975, CompoundOn: []string{"GST"}},
		}
	default:
		return nil
	}
}
