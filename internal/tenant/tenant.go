// Package tenant models the single multi-tenancy axis (ADR-0002) and the
// persona profiles a member can act under.
package tenant

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Profile is a persona that scopes what a member can see and do.
type Profile string

const (
	ProfileDeveloper      Profile = "developer"
	ProfileProductManager Profile = "product_manager"
	ProfileSRE            Profile = "sre"
	ProfileBillingAdmin   Profile = "billing_admin"
)

// ValidProfiles lists every supported persona.
func ValidProfiles() []Profile {
	return []Profile{ProfileDeveloper, ProfileProductManager, ProfileSRE, ProfileBillingAdmin}
}

// Valid reports whether p is a known profile.
func (p Profile) Valid() bool {
	for _, v := range ValidProfiles() {
		if v == p {
			return true
		}
	}
	return false
}

// Member is a user within a tenant acting under one persona profile.
type Member struct {
	Subject string  `json:"subject"` // identity from Dex/SPIFFE
	Profile Profile `json:"profile"`
}

// Tenant is the join key across isolation, identity, usage, and billing.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Members   []Member  `json:"members,omitempty"`
	PriceBook string    `json:"priceBook,omitempty"` // assigned price book version
	CreatedAt time.Time `json:"createdAt"`
}

// Errors returned by a Store.
var (
	ErrNotFound = errors.New("tenant not found")
	ErrExists   = errors.New("tenant already exists")
	ErrInvalid  = errors.New("invalid tenant")
)

// Store is the tenant persistence seam (in-memory now, Postgres in Phase 1).
type Store interface {
	Create(t Tenant) (Tenant, error)
	Get(id string) (Tenant, error)
	List() []Tenant
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Tenant
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Tenant{}} }

// Create stores a new tenant, defaulting CreatedAt and validating profiles.
func (s *MemStore) Create(t Tenant) (Tenant, error) {
	if t.ID == "" || t.Name == "" {
		return Tenant{}, ErrInvalid
	}
	for _, m := range t.Members {
		if !m.Profile.Valid() {
			return Tenant{}, ErrInvalid
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[t.ID]; ok {
		return Tenant{}, ErrExists
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	s.m[t.ID] = t
	return t, nil
}

// Get returns a tenant by ID.
func (s *MemStore) Get(id string) (Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.m[id]
	if !ok {
		return Tenant{}, ErrNotFound
	}
	return t, nil
}

// List returns all tenants sorted by ID.
func (s *MemStore) List() []Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Tenant, 0, len(s.m))
	for _, t := range s.m {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
