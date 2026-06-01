package provider

import (
	"context"
	"fmt"
	"sync/atomic"
)

// stub is a minimal in-memory Provider used until real compositions land.
// Both Kubernetes and CloudStack ship as stubs in Phase 0; they prove the seam
// without duplicating logic — only Name() differs.
type stub struct {
	name string
	seq  atomic.Int64
}

func (s *stub) Name() string { return s.name }

func (s *stub) Provision(_ context.Context, tenantID string, r Resource) (Instance, error) {
	if tenantID == "" {
		return Instance{}, fmt.Errorf("%s: tenantID required", s.name)
	}
	id := fmt.Sprintf("%s-%s-%d", s.name, r.Kind, s.seq.Add(1))
	return Instance{ID: id, Provider: s.name, Kind: r.Kind, Status: "provisioning"}, nil
}

func (s *stub) Deprovision(_ context.Context, tenantID, instanceID string) error {
	if tenantID == "" || instanceID == "" {
		return fmt.Errorf("%s: tenantID and instanceID required", s.name)
	}
	return nil
}

// NewKubernetes returns the Kubernetes provider (stub in Phase 0).
func NewKubernetes() Provider { return &stub{name: "kubernetes"} }

// NewCloudStack returns the Apache CloudStack provider (stub in Phase 0).
func NewCloudStack() Provider { return &stub{name: "cloudstack"} }

// NewEdge returns the edge provider (KubeEdge/K3s, stub in Phase 0) so the same
// control plane can place workloads at the edge.
func NewEdge() Provider { return &stub{name: "edge"} }

// DefaultRegistry returns a Registry preloaded with the built-in providers,
// demonstrating vendor neutrality and multi-cloud + edge reach: CloudStack is
// one option among several behind the same seam.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewKubernetes()) // any conformant cluster: on-prem or any cloud
	r.Register(NewCloudStack())
	r.Register(NewEdge())
	return r
}
