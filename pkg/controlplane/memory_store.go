package controlplane

type MemoryStore struct {
	tenants   map[string]Tenant
	offerings map[string]Offering
	usage     []UsageRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tenants:   map[string]Tenant{},
		offerings: map[string]Offering{},
		usage:     []UsageRecord{},
	}
}

func (m *MemoryStore) PutTenant(t Tenant) error {
	m.tenants[t.ID] = t
	return nil
}

func (m *MemoryStore) ListTenants() ([]Tenant, error) {
	out := make([]Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		out = append(out, t)
	}
	return out, nil
}

func (m *MemoryStore) PutOffering(o Offering) error {
	m.offerings[o.ID] = o
	return nil
}

func (m *MemoryStore) ListOfferings() ([]Offering, error) {
	out := make([]Offering, 0, len(m.offerings))
	for _, o := range m.offerings {
		out = append(out, o)
	}
	return out, nil
}

func (m *MemoryStore) PutUsage(u UsageRecord) error {
	m.usage = append(m.usage, u)
	return nil
}

func (m *MemoryStore) ListUsage(tenantID string) ([]UsageRecord, error) {
	out := []UsageRecord{}
	for _, u := range m.usage {
		if tenantID == "" || u.TenantID == tenantID {
			out = append(out, u)
		}
	}
	return out, nil
}
