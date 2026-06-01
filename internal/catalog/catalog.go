// Package catalog exposes CNCF (and AI-native) projects as provisionable,
// metered services. Each offering binds an upstream project, a Crossplane
// composition, the meters it is billed on, and the personas that may see it.
package catalog

import (
	"encoding/json"
	"fmt"
	"io"
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

	// Certifications lists the compliance frameworks this offering is certified
	// for (e.g. "SOC2", "ISO-27001", "GDPR"). The compliance engine checks these
	// against a tenant's required frameworks at placement time.
	Certifications []string `json:"certifications,omitempty"`
}

// Load reads a JSON array of Offerings from r into a new MemStore. Catalog data
// is loaded at deployment time (e.g. from a mounted ConfigMap), not baked in.
func Load(r io.Reader) (*MemStore, error) {
	var offerings []Offering
	if err := json.NewDecoder(r).Decode(&offerings); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}
	s := NewMemStore()
	for _, o := range offerings {
		if o.ID == "" {
			return nil, fmt.Errorf("offering missing id")
		}
		s.Add(o)
	}
	return s, nil
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

// ForCategory returns offerings in a CNCF-landscape category (the category-wise
// registry view), enabling a composable full-stack picker.
func (s *MemStore) ForCategory(category string) []Offering {
	var out []Offering
	for _, o := range s.List() {
		if o.Category == category {
			out = append(out, o)
		}
	}
	return out
}

// Categories returns the distinct categories present, sorted (the registry index).
func (s *MemStore) Categories() []string {
	seen := map[string]bool{}
	for _, o := range s.List() {
		if o.Category != "" {
			seen[o.Category] = true
		}
	}
	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// Seeded returns a MemStore preloaded with representative CNCF and AI-native
// offerings. Adding an offering is data, not code — keeping the system
// composable.
func Seeded() *MemStore {
	s := NewMemStore()
	all := []string{"developer", "product_manager", "sre", "billing_admin"}
	tech := []string{"developer", "sre"}
	// Focused MVP: the open-source, AWS-compatible service set. Each maps an AWS
	// service to an open-source backend, exposed with an AWS-compatible API so
	// existing AWS SDKs/tools interoperate (see docs/aws-interop.md).
	for _, o := range []Offering{
		// compute (EC2-compatible)
		{ID: "compute", Name: "Compute (EC2-compatible)", Project: "Kubernetes + KubeVirt", Category: "compute",
			Composition: "xcompute.platform.unboxd/v1", Meters: []string{"compute.vcpu.hour", "compute.mem.gb.hour"}, Profiles: all, Certifications: []string{"SOC2", "ISO-27001"}},
		// lambda (Lambda-compatible functions)
		{ID: "lambda", Name: "Functions (Lambda-compatible)", Project: "Knative / OpenFaaS", Category: "serverless",
			Composition: "xfunctions.platform.unboxd/v1", Meters: []string{"function.invocation.million", "function.gb.second"}, Profiles: tech, Certifications: []string{"SOC2"}},
		// sts (Security Token Service)
		{ID: "sts", Name: "Security Token Service (STS-compatible)", Project: "Dex + SPIFFE/SPIRE", Category: "security",
			Composition: "xsts.platform.unboxd/v1", Meters: []string{"token.issued.million"}, Profiles: tech, Certifications: []string{"SOC2", "ISO-27001"}},
		// sns (notifications / pub-sub)
		{ID: "sns", Name: "Notifications (SNS-compatible)", Project: "NATS", Category: "messaging",
			Composition: "xsns.platform.unboxd/v1", Meters: []string{"messaging.msg.million"}, Profiles: tech, Certifications: []string{"SOC2"}},
		// ses (email)
		{ID: "ses", Name: "Email (SES-compatible)", Project: "Postal / Haraka SMTP", Category: "messaging",
			Composition: "xses.platform.unboxd/v1", Meters: []string{"email.sent.thousand"}, Profiles: tech, Certifications: []string{"SOC2", "GDPR"}},
		// s3 (S3-compatible object storage)
		{ID: "s3", Name: "Object Storage (S3-compatible)", Project: "Rook/Ceph RGW", Category: "storage",
			Composition: "xs3.platform.unboxd/v1", Meters: []string{"storage.gb.month", "network.egress.gb", "s3.request.million"}, Profiles: all, Certifications: []string{"SOC2", "ISO-27001", "GDPR"}},
		// bedrock (model inference on open-source CPU LLMs)
		{ID: "bedrock", Name: "Model Inference (Bedrock-compatible)", Project: "KServe + llama.cpp/Ollama", Category: "ai",
			Composition: "xbedrock.platform.unboxd/v1", Meters: []string{"ai.tokens.million", "ai.cpu.hour", "ai.gpu.hour"}, Profiles: tech, Certifications: []string{"SOC2"}},
		// agentcore (agent runtime)
		{ID: "agentcore", Name: "Agent Runtime (AgentCore-compatible)", Project: "Dapr Agents (open-source)", Category: "ai",
			Composition: "xagentcore.platform.unboxd/v1", Meters: []string{"agent.run.hour", "ai.tokens.million", "compute.vcpu.hour"}, Profiles: tech, Certifications: []string{"SOC2"}},
	} {
		s.Add(o)
	}
	return s
}
