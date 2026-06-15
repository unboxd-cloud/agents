package controlplane

import (
	"errors"
	"time"
)

type Store interface {
	PutTenant(Tenant) error
	ListTenants() ([]Tenant, error)
	PutOffering(Offering) error
	ListOfferings() ([]Offering, error)
	PutUsage(UsageRecord) error
	ListUsage(tenantID string) ([]UsageRecord, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) Service {
	return Service{store: store, now: time.Now}
}

func (s Service) RegisterTenant(t Tenant) (Tenant, error) {
	if t.ID == "" {
		return Tenant{}, errors.New("tenant id is required")
	}
	if t.Provider == "" {
		t.Provider = "cloudstack"
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = s.now().UTC()
	}
	return t, s.store.PutTenant(t)
}

func (s Service) RegisterOffering(o Offering) (Offering, error) {
	if o.ID == "" {
		return Offering{}, errors.New("offering id is required")
	}
	if o.Provider == "" {
		o.Provider = "cloudstack"
	}
	return o, s.store.PutOffering(o)
}

func (s Service) IngestUsage(u UsageRecord) (UsageRecord, error) {
	if u.ID == "" {
		return UsageRecord{}, errors.New("usage id is required")
	}
	if u.TenantID == "" {
		return UsageRecord{}, errors.New("tenant id is required")
	}
	if u.Provider == "" {
		u.Provider = "cloudstack"
	}
	if u.ObservedAt.IsZero() {
		u.ObservedAt = s.now().UTC()
	}
	return u, s.store.PutUsage(u)
}

func (s Service) RateUsage(u UsageRecord, unitPrice float64) BillLine {
	return BillLine{
		TenantID:   u.TenantID,
		OfferingID: u.OfferingID,
		Quantity:   u.Quantity,
		Unit:       u.Unit,
		UnitPrice:  unitPrice,
		Amount:     u.Quantity * unitPrice,
	}
}
