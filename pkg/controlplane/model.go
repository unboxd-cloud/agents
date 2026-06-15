package controlplane

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Provider  string    `json:"provider"`
	AccountID string    `json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Offering struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Provider    string            `json:"provider"`
	ProviderRef string            `json:"provider_ref"`
	Unit        string            `json:"unit"`
	Attributes  map[string]string `json:"attributes"`
}

type UsageRecord struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	OfferingID  string    `json:"offering_id"`
	Provider    string    `json:"provider"`
	ProviderRef string    `json:"provider_ref"`
	Quantity    float64   `json:"quantity"`
	Unit        string    `json:"unit"`
	ObservedAt  time.Time `json:"observed_at"`
}

type BillLine struct {
	TenantID   string  `json:"tenant_id"`
	OfferingID string  `json:"offering_id"`
	Quantity   float64 `json:"quantity"`
	Unit       string  `json:"unit"`
	UnitPrice  float64 `json:"unit_price"`
	Amount     float64 `json:"amount"`
}

type ComplianceEvidence struct {
	TenantID string    `json:"tenant_id"`
	Control  string    `json:"control"`
	Status   string    `json:"status"`
	Source   string    `json:"source"`
	Evidence string    `json:"evidence"`
	Time     time.Time `json:"time"`
}
