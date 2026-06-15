// Package orchestrator provides the Kubernetes orchestrator agent: it reconciles
// each tenant's desired set of resources toward actual provisioned instances via
// the vendor-neutral provider seam. This realizes "Kubernetes orchestrator as an
// agent" on the shared agent runtime (level-triggered, like a controller).
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/unboxd-cloud/platform/internal/provider"
)

// Desired is a tenant's desired set of resources.
type Desired struct {
	TenantID  string
	Resources []provider.Resource
}

func key(tenantID, kind, name string) string {
	return tenantID + "/" + kind + "/" + name
}

// Orchestrator reconciles Desired specs to actual instances through a Provider.
type Orchestrator struct {
	provider provider.Provider

	mu      sync.Mutex
	desired map[string]Desired           // by tenant
	actual  map[string]provider.Instance // by resource key
}

// New returns an Orchestrator backed by the given provider.
func New(p provider.Provider) *Orchestrator {
	return &Orchestrator{
		provider: p,
		desired:  map[string]Desired{},
		actual:   map[string]provider.Instance{},
	}
}

// Name implements agent.Agent.
func (o *Orchestrator) Name() string { return "k8s-orchestrator" }

// SetDesired records (or replaces) a tenant's desired resources.
func (o *Orchestrator) SetDesired(d Desired) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.desired[d.TenantID] = d
}

// Reconcile provisions any desired resource that has no actual instance yet
// (create-only in Phase 0; deletes/drift-repair are a later phase).
func (o *Orchestrator) Reconcile(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, d := range o.desired {
		for _, r := range d.Resources {
			k := key(d.TenantID, r.Kind, r.Name)
			if _, ok := o.actual[k]; ok {
				continue // already reconciled
			}
			inst, err := o.provider.Provision(ctx, d.TenantID, r)
			if err != nil {
				return fmt.Errorf("provision %s: %w", k, err)
			}
			o.actual[k] = inst
			log.Printf("orchestrator: provisioned %s -> %s (%s)", k, inst.ID, inst.Status)
		}
	}
	return nil
}

// Instances returns the currently provisioned instances, sorted by ID.
func (o *Orchestrator) Instances() []provider.Instance {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]provider.Instance, 0, len(o.actual))
	for _, i := range o.actual {
		out = append(out, i)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
