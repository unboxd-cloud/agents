package controlplane

import "testing"

func TestRegisterTenantDefaultsProvider(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)

	tenant, err := svc.RegisterTenant(Tenant{ID: "tenant-1", Name: "Tenant One"})
	if err != nil {
		t.Fatal(err)
	}
	if tenant.Provider != "cloudstack" {
		t.Fatalf("expected cloudstack provider, got %s", tenant.Provider)
	}
}

func TestRegisterOfferingDefaultsProvider(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)

	offering, err := svc.RegisterOffering(Offering{ID: "cs-small", Name: "Small Compute"})
	if err != nil {
		t.Fatal(err)
	}
	if offering.Provider != "cloudstack" {
		t.Fatalf("expected cloudstack provider, got %s", offering.Provider)
	}
}

func TestIngestUsageAndRate(t *testing.T) {
	store := NewMemoryStore()
	svc := NewService(store)

	u, err := svc.IngestUsage(UsageRecord{
		ID:         "usage-1",
		TenantID:   "tenant-1",
		OfferingID: "cs-small",
		Quantity:   10,
		Unit:       "hour",
	})
	if err != nil {
		t.Fatal(err)
	}

	line := svc.RateUsage(u, 2.5)
	if line.Amount != 25 {
		t.Fatalf("expected amount 25, got %f", line.Amount)
	}
}
