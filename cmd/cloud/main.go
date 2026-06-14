// Command cloud is the platform's cloud control plane. It exposes an Apache
// CloudStack-compatible API (the contract) and reconciles the virtual machines
// clients request onto Kubernetes pods (the reconciler + pod manager), so the
// platform speaks CloudStack while running CNCF-natively. See ADR-0007.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/unboxd-cloud/platform/internal/agent"
	"github.com/unboxd-cloud/platform/internal/controlplane"
	"github.com/unboxd-cloud/platform/internal/kube"
	"github.com/unboxd-cloud/platform/internal/server"
)

func main() {
	interval := envDuration("RECONCILE_INTERVAL", 5*time.Second)
	cp := controlplane.New(kube.NewManager())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Kubernetes is the reconciler: drive the control plane's reconcile loop in
	// the background so desired VMs converge to actual pods (level-triggered).
	go func() { _ = agent.Run(ctx, cp, interval) }()

	mux := http.NewServeMux()
	mux.Handle("/", controlplane.Handler(cp))

	addr := envOr("CLOUD_ADDR", ":8086")
	srv := server.New(addr, mux)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("cloud control plane listening on %s (CloudStack contract, k8s reconciler every %s)", addr, interval)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cloud: %v", err)
	}
	log.Printf("cloud: shutting down")
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
