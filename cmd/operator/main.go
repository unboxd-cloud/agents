// Command operator is the generic platform operator agent. It supervises the
// platform's control loops — the GitOps reconciler and the Kubernetes
// orchestrator — on the shared agent runtime.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/unboxd-cloud/platform/internal/agent"
	"github.com/unboxd-cloud/platform/internal/gitops"
	"github.com/unboxd-cloud/platform/internal/orchestrator"
	"github.com/unboxd-cloud/platform/internal/provider"
)

func main() {
	dir := envOr("GITOPS_DIR", "deploy/datasets")
	interval := envDuration("RECONCILE_INTERVAL", 30*time.Second)

	// Vendor-neutral: choose the provider by name (kubernetes|cloudstack|edge).
	prov, err := provider.DefaultRegistry().Get(envOr("PROVIDER", "kubernetes"))
	if err != nil {
		log.Fatalf("operator: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("operator: provider=%s gitops-dir=%s interval=%s", prov.Name(), dir, interval)
	_ = agent.Operator(ctx,
		agent.Scheduled{Agent: &gitops.Reconciler{Dir: dir}, Interval: interval},
		agent.Scheduled{Agent: orchestrator.New(prov), Interval: interval},
	)
	log.Printf("operator: shutting down")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
