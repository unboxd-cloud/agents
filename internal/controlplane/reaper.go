package controlplane

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/unboxd-cloud/platform/internal/kube"
)

// Reaper is a second operator that runs on the same core as the control-plane
// reconciler: it garbage-collects pods whose backing VM no longer exists (e.g.
// left behind by a crash), keeping actual state converged to desired. Sharing
// the pod manager and store with the control plane, it demonstrates multiple
// operators on one core over one store.
type Reaper struct {
	pods  kube.PodManager
	store Store
}

// NewReaper returns a Reaper over the given pod manager and store.
func NewReaper(pods kube.PodManager, store Store) *Reaper {
	return &Reaper{pods: pods, store: store}
}

// Name implements agent.Agent.
func (r *Reaper) Name() string { return "pod-reaper" }

// Reconcile deletes control-plane-managed pods that have no backing VM.
func (r *Reaper) Reconcile(ctx context.Context) error {
	pods, err := r.pods.List(ctx, "") // all namespaces
	if err != nil {
		return fmt.Errorf("reaper: list pods: %w", err)
	}
	for _, p := range pods {
		id := p.Labels["unboxd.cloud/vm-id"]
		if id == "" {
			continue // not managed by the control plane
		}
		_, ok, err := r.store.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("reaper: store get %s: %w", id, err)
		}
		if ok {
			continue // VM still desired
		}
		if err := r.pods.Delete(ctx, p.Namespace, p.Name); err != nil && !errors.Is(err, kube.ErrNotFound) {
			return fmt.Errorf("reaper: delete orphan pod %s/%s: %w", p.Namespace, p.Name, err)
		}
		log.Printf("reaper: reaped orphan pod %s/%s (vm %s gone)", p.Namespace, p.Name, id)
	}
	return nil
}
