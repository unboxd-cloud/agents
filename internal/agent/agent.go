// Package agent is the minimal reconcile-loop runtime shared by the platform's
// agents (GitOps, Kubernetes orchestrator, ...). An agent is anything that can
// reconcile desired state toward actual state, repeatedly.
package agent

import (
	"context"
	"log"
	"sync"
	"time"
)

// Agent reconciles desired state toward actual state.
type Agent interface {
	Name() string
	Reconcile(ctx context.Context) error
}

// Scheduled pairs an Agent with the interval it should reconcile on.
type Scheduled struct {
	Agent    Agent
	Interval time.Duration
}

// Operator is the generic platform operator: it supervises any set of agents,
// running each on its own interval until the context is cancelled. Composing
// agents (GitOps, orchestrator, ...) under one operator is the standard way to
// run the platform's control loops.
func Operator(ctx context.Context, agents ...Scheduled) error {
	var wg sync.WaitGroup
	for _, s := range agents {
		wg.Add(1)
		go func(s Scheduled) {
			defer wg.Done()
			log.Printf("operator: starting agent %s (every %s)", s.Agent.Name(), s.Interval)
			_ = Run(ctx, s.Agent, s.Interval)
		}(s)
	}
	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
}

// Run drives an Agent: once immediately, then every interval, until the context
// is cancelled. Reconcile errors are logged, not fatal — the loop keeps trying
// (level-triggered reconciliation, GitOps-style).
func Run(ctx context.Context, a Agent, interval time.Duration) error {
	reconcile := func() {
		if err := a.Reconcile(ctx); err != nil {
			log.Printf("agent %s: reconcile error: %v", a.Name(), err)
		}
	}
	reconcile()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			reconcile()
		}
	}
}
