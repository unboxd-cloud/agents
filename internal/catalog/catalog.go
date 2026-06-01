// Package catalog exposes CNCF (and AI-native) projects as provisionable,
// metered services. Each offering binds an upstream project, a Crossplane
// composition, the meters it is billed on, and the personas that may see it.
package catalog

import (
	"sort"
	"sync"
)

// Offering is one provisionable, metered service in the catalog.
type Offering struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Project     string   `json:"project"`     // upstream CNCF / AI project
	Category    string   `json:"category"`    // compute, data, observability, ai, ...
	Composition string   `json:"composition"` // Crossplane composition ref
	Meters      []string `json:"meters"`      // meter keys billed pay-as-you-go
	Profiles    []string `json:"profiles"`    // personas allowed to see/order it

	// Marketplace publishing model. Publisher is the listing owner ("platform"
	// for first-party). RevShare is the publisher's share of rated revenue
	// (0.0-1.0); the platform keeps the remainder. This lets third parties
	// publish offerings, settled via the same billing engine.
	Publisher string  `json:"publisher,omitempty"`
	RevShare  float64 `json:"revShare,omitempty"`
}

// Store is the catalog persistence seam.
type Store interface {
	List() []Offering
	ForProfile(profile string) []Offering
	Get(id string) (Offering, bool)
}

// MemStore is an in-memory catalog.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Offering
}

// NewMemStore returns an in-memory catalog with no entries.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Offering{}} }

// Add inserts or replaces an offering.
func (s *MemStore) Add(o Offering) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[o.ID] = o
}

// Get returns an offering by ID.
func (s *MemStore) Get(id string) (Offering, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.m[id]
	return o, ok
}

// List returns all offerings sorted by ID.
func (s *MemStore) List() []Offering {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Offering, 0, len(s.m))
	for _, o := range s.m {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ForProfile returns offerings visible to the given persona profile.
func (s *MemStore) ForProfile(profile string) []Offering {
	var out []Offering
	for _, o := range s.List() {
		for _, p := range o.Profiles {
			if p == profile {
				out = append(out, o)
				break
			}
		}
	}
	return out
}

// Seeded returns a MemStore preloaded with representative CNCF and AI-native
// offerings. Adding an offering is data, not code — keeping the system
// composable.
func Seeded() *MemStore {
	s := NewMemStore()
	all := []string{"developer", "product_manager", "sre", "billing_admin"}
	tech := []string{"developer", "sre"}
	for _, o := range []Offering{
		{ID: "managed-kubernetes", Name: "Managed Kubernetes", Project: "vCluster", Category: "compute",
			Composition: "xcluster.platform.unboxd/v1", Meters: []string{"compute.vcpu.hour", "compute.mem.gb.hour"}, Profiles: all},
		{ID: "managed-prometheus", Name: "Managed Prometheus", Project: "Prometheus", Category: "observability",
			Composition: "xmonitoring.platform.unboxd/v1", Meters: []string{"metrics.series.hour"}, Profiles: tech},
		{ID: "managed-nats", Name: "Managed NATS", Project: "NATS", Category: "messaging",
			Composition: "xmessaging.platform.unboxd/v1", Meters: []string{"messaging.msg.million"}, Profiles: tech},
		{ID: "object-storage", Name: "Object Storage", Project: "Rook", Category: "data",
			Composition: "xobjectstore.platform.unboxd/v1", Meters: []string{"storage.gb.month", "network.egress.gb"}, Profiles: all},
		{ID: "managed-inference", Name: "Managed Model Inference", Project: "KServe", Category: "ai",
			Composition: "xinference.platform.unboxd/v1", Meters: []string{"ai.gpu.hour", "ai.tokens.million"}, Profiles: tech},
		{ID: "ml-pipelines", Name: "ML Pipelines", Project: "Kubeflow", Category: "ai",
			Composition: "xpipelines.platform.unboxd/v1", Meters: []string{"ai.gpu.hour", "compute.vcpu.hour"}, Profiles: tech},
		// Example third-party marketplace listing with a publisher revenue share.
		{ID: "vector-db", Name: "Vector Database", Project: "Milvus", Category: "ai",
			Composition: "xvectordb.partner.example/v1", Meters: []string{"storage.gb.month", "ai.tokens.million"},
			Profiles: tech, Publisher: "partner-acme", RevShare: 0.80},
	} {
		s.Add(o)
	}
	return s
}
