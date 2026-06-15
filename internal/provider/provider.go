// Package provider is the single vendor-neutral seam to infrastructure.
//
// Kubernetes, Apache CloudStack, and public clouds are interchangeable
// implementations behind one interface (ADR-0001). Control-plane services never
// import a specific vendor; they take a Provider (or a Registry) so swapping a
// vendor is configuration, not code edits.
package provider

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// Resource is a vendor-neutral description of something to provision.
type Resource struct {
	Kind   string            `json:"kind"` // e.g. "kubernetes", "compute", "objectstore"
	Name   string            `json:"name"`
	Params map[string]string `json:"params,omitempty"`
}

// Instance is a handle to a provisioned resource.
type Instance struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
}

// Provider is the one seam any infrastructure vendor implements.
type Provider interface {
	Name() string
	Provision(ctx context.Context, tenantID string, r Resource) (Instance, error)
	Deprovision(ctx context.Context, tenantID, instanceID string) error
}

// ErrUnknownProvider is returned when a name is not registered.
var ErrUnknownProvider = errors.New("unknown provider")

// Registry holds available providers by name. Composable: register any
// implementation; selection is configuration.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{providers: map[string]Provider{}}
}

// Register adds (or replaces) a provider.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, ErrUnknownProvider
	}
	return p, nil
}

// Names lists registered providers, sorted.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
