// Package compliance adds support for law-of-the-land, industry-specific, and
// security frameworks (e.g. GDPR, HIPAA, PCI-DSS, SOC2, ISO-27001, FedRAMP).
//
// It follows the platform tenets: the engine is data-free (frameworks are
// loaded as datasets at deployment, not baked in), persistence is behind a
// Store interface (database-agnostic), and enforcement is delegated to the
// existing OPA policy gate + OpenFGA relationship checks. Compliance is keyed by
// the same TenantID join axis as the rest of the platform (ADR-0002).
package compliance

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
)

// Category groups frameworks by domain.
type Category string

const (
	CategoryPrivacy    Category = "privacy"
	CategoryHealthcare Category = "healthcare"
	CategoryFinance    Category = "finance"
	CategorySecurity   Category = "security"
	CategoryGovernment Category = "government"
)

// Spec is the loadable definition of a framework. Specs are data, supplied at
// deployment via a dataset (see Registry.Load); nothing is hard-coded here.
type Spec struct {
	Framework          string   `json:"framework"` // stable key, e.g. "GDPR", "HIPAA"
	Name               string   `json:"name"`
	Category           Category `json:"category"`
	Authority          string   `json:"authority"` // regulator / standards body
	Regions            []string `json:"regions"`   // jurisdictions where it applies
	Controls           []string `json:"controls"`  // representative control domains
	RequiresEncryption bool     `json:"requiresEncryption"`
	Description        string   `json:"description"`
}

// Registry holds the framework specs loaded for a deployment.
type Registry struct {
	mu sync.RWMutex
	m  map[string]Spec
}

// NewRegistry returns an empty Registry. Populate it with Load at startup.
func NewRegistry() *Registry { return &Registry{m: map[string]Spec{}} }

// Register adds or replaces a spec.
func (r *Registry) Register(s Spec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[s.Framework] = s
}

// Load reads a JSON array of Specs from r and registers them. Datasets are
// loaded at deployment time (e.g. from a mounted ConfigMap).
func (r *Registry) Load(reader io.Reader) (int, error) {
	var specs []Spec
	if err := json.NewDecoder(reader).Decode(&specs); err != nil {
		return 0, fmt.Errorf("decode compliance specs: %w", err)
	}
	for _, s := range specs {
		if s.Framework == "" {
			return 0, fmt.Errorf("spec missing framework key")
		}
		r.Register(s)
	}
	return len(specs), nil
}

// Get returns a spec by framework key.
func (r *Registry) Get(framework string) (Spec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.m[framework]
	return s, ok
}

// List returns all specs sorted by framework key.
func (r *Registry) List() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Spec, 0, len(r.m))
	for _, s := range r.m {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Framework < out[j].Framework })
	return out
}

// Filter returns specs in a category.
func (r *Registry) Filter(c Category) []Spec {
	var out []Spec
	for _, s := range r.List() {
		if s.Category == c {
			out = append(out, s)
		}
	}
	return out
}

// Profile is a tenant's compliance posture (assigned per TenantID).
type Profile struct {
	TenantID      string   `json:"tenantId"`
	Jurisdiction  string   `json:"jurisdiction"`  // legal jurisdiction (also drives tax)
	Frameworks    []string `json:"frameworks"`    // frameworks the tenant must meet
	DataResidency []string `json:"dataResidency"` // allowed regions for data at rest/processing
}

// Placement is a proposed resource placement to check against a Profile.
type Placement struct {
	Provider       string   `json:"provider"`
	Region         string   `json:"region"`
	OfferingID     string   `json:"offeringId"`
	Certifications []string `json:"certifications"` // frameworks the offering is certified for
	Encrypted      bool     `json:"encrypted"`
}

// Finding is one control check result.
type Finding struct {
	Framework string `json:"framework,omitempty"`
	Control   string `json:"control"`
	OK        bool   `json:"ok"`
	Detail    string `json:"detail"`
}

// Report is the outcome of evaluating a placement against a profile.
type Report struct {
	TenantID  string    `json:"tenantId"`
	Compliant bool      `json:"compliant"`
	Findings  []Finding `json:"findings"`
}

// Evaluate checks a placement against a profile. The Registry (loaded data)
// informs framework-specific requirements such as encryption-at-rest; the
// engine itself carries no framework data. A nil Registry simply skips
// data-dependent controls. OPA/OpenFGA enforce the resulting decision.
func Evaluate(p Profile, pl Placement, reg *Registry) Report {
	rep := Report{TenantID: p.TenantID, Compliant: true}
	add := func(f Finding) {
		if !f.OK {
			rep.Compliant = false
		}
		rep.Findings = append(rep.Findings, f)
	}

	// Law of the land: data residency must be within the allowed regions.
	if len(p.DataResidency) > 0 {
		ok := contains(p.DataResidency, pl.Region)
		add(Finding{Control: "data-residency", OK: ok,
			Detail: fmt.Sprintf("region %q allowed=%v (allowed: %v)", pl.Region, ok, p.DataResidency)})
	}

	// Industry/framework: the chosen offering must be certified, and encryption
	// is required when any required framework demands it.
	encryptionRequired := false
	for _, fw := range p.Frameworks {
		ok := contains(pl.Certifications, fw)
		add(Finding{Framework: fw, Control: "offering-certification", OK: ok,
			Detail: fmt.Sprintf("offering %q certified for %s=%v", pl.OfferingID, fw, ok)})
		if reg != nil {
			if s, found := reg.Get(fw); found && s.RequiresEncryption {
				encryptionRequired = true
			}
		}
	}
	if encryptionRequired {
		add(Finding{Control: "encryption-at-rest", OK: pl.Encrypted,
			Detail: fmt.Sprintf("encryption-at-rest required, enabled=%v", pl.Encrypted)})
	}
	return rep
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// Store persists tenant compliance profiles (database-agnostic seam).
type Store interface {
	Set(p Profile) error
	Get(tenantID string) (Profile, bool)
	List() []Profile
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Profile
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Profile{}} }

// Set stores a profile keyed by tenant.
func (s *MemStore) Set(p Profile) error {
	if p.TenantID == "" {
		return fmt.Errorf("profile missing tenantId")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[p.TenantID] = p
	return nil
}

// Get returns a tenant's profile.
func (s *MemStore) Get(tenantID string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.m[tenantID]
	return p, ok
}

// List returns all profiles sorted by tenant.
func (s *MemStore) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Profile, 0, len(s.m))
	for _, p := range s.m {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TenantID < out[j].TenantID })
	return out
}
