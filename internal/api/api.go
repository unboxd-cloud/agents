// Package api holds the request/response DTOs shared by the control-plane
// services and the SDK, so the wire contract is defined once (no duplication).
package api

import (
	"time"

	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/compliance"
	"github.com/unboxd-cloud/platform/internal/metering"
)

// RateRequest is the stateless billing rating input.
type RateRequest struct {
	TenantID     string                `json:"tenantId"`
	From         time.Time             `json:"from"`
	To           time.Time             `json:"to"`
	Events       []metering.UsageEvent `json:"events"`
	PriceBook    *billing.PriceBook    `json:"priceBook,omitempty"`
	Partner      *billing.Partner      `json:"partner,omitempty"`
	Jurisdiction string                `json:"jurisdiction,omitempty"`
}

// RateResponse composes the rated invoice with optional settlement and tax.
type RateResponse struct {
	Invoice    billing.Invoice     `json:"invoice"`
	Settlement *billing.Settlement `json:"settlement,omitempty"`
	Tax        *billing.TaxResult  `json:"tax,omitempty"`
}

// EvalRequest pairs a tenant compliance profile with a proposed placement.
type EvalRequest struct {
	Profile   compliance.Profile   `json:"profile"`
	Placement compliance.Placement `json:"placement"`
}
