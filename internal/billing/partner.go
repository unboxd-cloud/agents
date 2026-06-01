package billing

// OperatingMode is how the platform is commercially operated for a tenant.
// Multiple modes can coexist across different tenants on one deployment.
type OperatingMode string

const (
	// ModeDirect bills the end customer at list price.
	ModeDirect OperatingMode = "direct"
	// ModeReseller bills a partner who resells to their own customers; a markup
	// or discount is applied to list price.
	ModeReseller OperatingMode = "reseller"
	// ModeAgency means a partner manages the account on the customer's behalf;
	// list price to the customer, a referral/commission to the agency.
	ModeAgency OperatingMode = "agency"
	// ModeMarketplace bills through a marketplace that takes a commission.
	ModeMarketplace OperatingMode = "marketplace"
	// ModeServiceProvider means an MSP wraps the platform under its own brand.
	ModeServiceProvider OperatingMode = "service_provider"
)

// Partner describes a commercial relationship layered over a tenant's invoice.
// One engine + this thin overlay expresses agency, marketplace, reselling, and
// service-provider models without duplicating the rater.
type Partner struct {
	ID   string        `json:"id"`
	Mode OperatingMode `json:"mode"`
	// Rate is mode-dependent: reseller/service-provider markup (e.g. 0.15 = +15%),
	// or marketplace/agency commission/discount (e.g. 0.10 = 10%). May be negative
	// for a reseller discount.
	Rate float64 `json:"rate"`
}

// Settlement is the result of applying a partner overlay to a base invoice.
type Settlement struct {
	Invoice    Invoice       `json:"invoice"`     // the underlying rated usage
	PartnerID  string        `json:"partnerId"`
	Mode       OperatingMode `json:"mode"`
	Adjustment float64       `json:"adjustment"`  // markup added or commission taken
	NetToPlatform float64    `json:"netToPlatform"`
	GrossToCustomer float64  `json:"grossToCustomer"`
	Currency   string        `json:"currency"`
}

// Settle applies a Partner overlay to a base invoice.
//
//   - Reseller / Service-provider: markup is added on top; customer pays gross,
//     platform receives the base (partner keeps the markup).
//   - Marketplace / Agency: a commission is taken from the base; customer pays
//     the base, platform nets base minus commission.
//   - Direct (or nil rate): pass-through.
func Settle(inv Invoice, p Partner) Settlement {
	s := Settlement{
		Invoice:   inv,
		PartnerID: p.ID,
		Mode:      p.Mode,
		Currency:  inv.Currency,
	}
	switch p.Mode {
	case ModeReseller, ModeServiceProvider:
		s.Adjustment = round2(inv.Total * p.Rate)
		s.GrossToCustomer = round2(inv.Total + s.Adjustment)
		s.NetToPlatform = inv.Total
	case ModeMarketplace, ModeAgency:
		s.Adjustment = round2(inv.Total * p.Rate)
		s.GrossToCustomer = inv.Total
		s.NetToPlatform = round2(inv.Total - s.Adjustment)
	default: // ModeDirect / unknown
		s.GrossToCustomer = inv.Total
		s.NetToPlatform = inv.Total
	}
	return s
}
