package depgraph

import (
	"context"
	"log"
)

// Resolver is the native dependency-tracker agent. Each reconcile it resolves
// the current graph into an install order and reports unmet dependencies,
// surfacing cycles as errors. It implements agent.Agent.
type Resolver struct {
	Graph *Graph
	// Known marks which nodes are actually installable (e.g. catalog offering
	// IDs). Dependencies outside this set are reported as unmet.
	Known map[string]bool
	// OnOrder, if set, receives the resolved order each cycle.
	OnOrder func(order []string)
}

// Name implements agent.Agent.
func (r *Resolver) Name() string { return "dependency-resolver" }

// Reconcile resolves the graph and reports order + unmet deps; returns ErrCycle
// on a cycle so the agent loop logs it (level-triggered, self-correcting).
func (r *Resolver) Reconcile(_ context.Context) error {
	order, err := r.Graph.Resolve()
	if err != nil {
		return err
	}
	if unmet := r.Graph.Missing(r.Known); len(unmet) > 0 {
		log.Printf("dependency-resolver: unmet dependencies: %v", unmet)
	}
	if r.OnOrder != nil {
		r.OnOrder(order)
	}
	return nil
}
