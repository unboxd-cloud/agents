// Package billing rates metered usage against a versioned price book to produce
// per-tenant invoices (ADR-0003). One engine expresses every pricing shape:
// flat rates are a single unbounded tier, free allowances are a leading
// subtraction, and graduated pricing is an ordered list of tiers.
package billing

import (
	"math"
	"sort"
	"time"

	"github.com/unboxd-cloud/platform/internal/metering"
)

// Tier is one graduated pricing band, applied to billable quantity.
type Tier struct {
	// UpTo is the inclusive cumulative upper bound for this band.
	// Zero means unbounded (must be the last tier).
	UpTo      float64 `json:"upTo"`
	UnitPrice float64 `json:"unitPrice"`
}

// MeterPrice prices one meter: an optional free allowance plus graduated tiers.
type MeterPrice struct {
	Meter     string  `json:"meter"`
	Allowance float64 `json:"allowance"` // free units before charging
	Tiers     []Tier  `json:"tiers"`     // ascending by UpTo; unbounded tier last
	Currency  string  `json:"currency"`
}

// PriceBook is a versioned set of meter prices.
type PriceBook struct {
	Version string                `json:"version"`
	Prices  map[string]MeterPrice `json:"prices"`
}

// LineItem is a rated meter for a tenant over a period.
type LineItem struct {
	Meter    string  `json:"meter"`
	Quantity float64 `json:"quantity"`
	Billable float64 `json:"billable"` // quantity after allowance
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Invoice is the per-tenant result of rating a period.
type Invoice struct {
	TenantID    string     `json:"tenantId"`
	PriceBook   string     `json:"priceBook"`
	From        time.Time  `json:"from"`
	To          time.Time  `json:"to"`
	Lines       []LineItem `json:"lines"`
	Total       float64    `json:"total"`
	Currency    string     `json:"currency"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

// Rate aggregates usage per meter and rates it once. It is a pure function of
// (price book, tenant, period, events), so invoices are reproducible.
func Rate(pb PriceBook, tenantID string, from, to time.Time, events []metering.UsageEvent) Invoice {
	totals := map[string]float64{}
	for _, e := range events {
		if e.TenantID != tenantID {
			continue
		}
		totals[e.Meter] += e.Quantity
	}

	inv := Invoice{
		TenantID:    tenantID,
		PriceBook:   pb.Version,
		From:        from,
		To:          to,
		GeneratedAt: time.Now().UTC(),
	}

	meters := make([]string, 0, len(totals))
	for m := range totals {
		meters = append(meters, m)
	}
	sort.Strings(meters)

	for _, m := range meters {
		price, ok := pb.Prices[m]
		if !ok {
			continue // unpriced meter: skip (could route to a default later)
		}
		qty := totals[m]
		billable := math.Max(0, qty-price.Allowance)
		amount := round2(rateTiers(price.Tiers, billable))
		inv.Lines = append(inv.Lines, LineItem{
			Meter:    m,
			Quantity: qty,
			Billable: billable,
			Amount:   amount,
			Currency: price.Currency,
		})
		inv.Total += amount
		inv.Currency = price.Currency
	}
	inv.Total = round2(inv.Total)
	return inv
}

// rateTiers walks graduated tiers once over the billable quantity.
func rateTiers(tiers []Tier, qty float64) float64 {
	var amount, prev float64
	for _, t := range tiers {
		if qty <= prev {
			break
		}
		upper := qty
		if t.UpTo > 0 && t.UpTo < qty {
			upper = t.UpTo
		}
		amount += (upper - prev) * t.UnitPrice
		if t.UpTo == 0 {
			break // unbounded tier consumed the remainder
		}
		prev = t.UpTo
	}
	return amount
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }
