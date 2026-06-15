// Package asset is the IT-admin asset discovery and catalog: a
// database-agnostic inventory of platform assets (services, providers,
// deployments, nodes, agents, connectors, images, datasets) with discovery and
// tagging, behind the same Store seam used elsewhere.
package asset

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Asset is a discovered, cataloged platform asset.
type Asset struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	Source       string    `json:"source,omitempty"`
	Status       string    `json:"status,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	DiscoveredAt time.Time `json:"discoveredAt"`
}

// ErrInvalid is returned for an asset missing id/kind/name.
var ErrInvalid = errors.New("invalid asset")

// Store is the asset catalog persistence seam (in-memory now).
type Store interface {
	Upsert(a Asset) (Asset, error)
	Get(id string) (Asset, bool)
	List(kind string) []Asset
	Kinds() map[string]int
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Asset
}

// NewMemStore returns an empty in-memory catalog.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Asset{}} }

// Upsert inserts or updates an asset, preserving DiscoveredAt and existing tags
// when the incoming asset omits them.
func (s *MemStore) Upsert(a Asset) (Asset, error) {
	if a.ID == "" || a.Kind == "" || a.Name == "" {
		return Asset{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ex, ok := s.m[a.ID]; ok {
		if a.DiscoveredAt.IsZero() {
			a.DiscoveredAt = ex.DiscoveredAt
		}
		if a.Tags == nil {
			a.Tags = ex.Tags
		}
	}
	if a.DiscoveredAt.IsZero() {
		a.DiscoveredAt = time.Now().UTC()
	}
	if a.Status == "" {
		a.Status = "active"
	}
	s.m[a.ID] = a
	return a, nil
}

// Get returns the asset with the given id.
func (s *MemStore) Get(id string) (Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.m[id]
	return a, ok
}

// List returns assets of the given kind ("" for all), ordered by kind then name.
func (s *MemStore) List(kind string) []Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Asset
	for _, a := range s.m {
		if kind == "" || a.Kind == kind {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Kinds returns the count of assets per kind.
func (s *MemStore) Kinds() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]int{}
	for _, a := range s.m {
		out[a.Kind]++
	}
	return out
}

// Source discovers assets from some part of the platform.
type Source func() []Asset

// Discover runs each source and upserts the results into the store, returning
// the number of assets cataloged.
func Discover(store Store, sources ...Source) int {
	n := 0
	for _, src := range sources {
		for _, a := range src() {
			if _, err := store.Upsert(a); err == nil {
				n++
			}
		}
	}
	return n
}
