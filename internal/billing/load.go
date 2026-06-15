package billing

import (
	"encoding/json"
	"fmt"
	"io"
)

// LoadPriceBook reads a PriceBook from JSON. Price books are versioned data
// loaded at deployment time (e.g. from a mounted ConfigMap), not baked in.
func LoadPriceBook(r io.Reader) (PriceBook, error) {
	var pb PriceBook
	if err := json.NewDecoder(r).Decode(&pb); err != nil {
		return PriceBook{}, fmt.Errorf("decode price book: %w", err)
	}
	if pb.Version == "" {
		return PriceBook{}, fmt.Errorf("price book missing version")
	}
	if pb.Prices == nil {
		pb.Prices = map[string]MeterPrice{}
	}
	return pb, nil
}

// LoadTaxTable reads a jurisdiction->rules tax table from JSON, loaded at
// deployment time. Use TaxTable.For to resolve rules per jurisdiction.
func LoadTaxTable(r io.Reader) (TaxTable, error) {
	var t TaxTable
	if err := json.NewDecoder(r).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode tax table: %w", err)
	}
	return t, nil
}

// TaxTable maps a jurisdiction code to its tax rules.
type TaxTable map[string][]TaxRule

// For returns the rules for a jurisdiction (nil if none), falling back to the
// built-in sample table when the table itself is nil.
func (t TaxTable) For(jurisdiction string) []TaxRule {
	if t == nil {
		return TaxRulesFor(jurisdiction)
	}
	return t[jurisdiction]
}
