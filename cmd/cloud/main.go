// Command cloud is the platform's cloud control plane. It exposes an Apache
// CloudStack-compatible API (the contract) and reconciles the virtual machines
// clients request onto Kubernetes pods (the reconciler + pod manager), so the
// platform speaks CloudStack while running CNCF-natively. See ADR-0007.
//
// One agent core (agent.Operator) hosts multiple operators — the control-plane
// reconciler and a pod reaper — over a configurable store (single core, multi
// store, multi operator). CLOUD_MODE selects which planes run (multi mode):
//
//	all       (default) serve the API and run the operators
//	api       serve the API only (write desired state to the store)
//	operator  run the operators only (reconcile the store; no API)
//
// In api/operator mode the planes share state through a persistent store
// (CLOUD_STORE=file), the apiserver/controller-manager split.
package main

import (
	"context"
	"fmt"
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
	"github.com/unboxd-cloud/platform/internal/workstation"
)

func main() {
	mode := envOr("CLOUD_MODE", "all")
	runAPI, runOperator, err := planes(mode)
	if err != nil {
		log.Fatalf("cloud: %v", err)
	}
	interval := envDuration("RECONCILE_INTERVAL", 5*time.Second)

	store, err := selectStore()
	if err != nil {
		log.Fatalf("cloud: %v", err)
	}
	pods := kube.NewManager()
	cp := controlplane.NewWithStore(pods, store)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Single core, multiple operators: one agent runtime drives the control-plane
	// reconciler and the pod reaper, both over the same store.
	if runOperator {
		go func() {
			_ = agent.Operator(ctx,
				agent.Scheduled{Agent: cp, Interval: interval},
				agent.Scheduled{Agent: controlplane.NewReaper(pods, store), Interval: interval},
			)
		}()
	}

	log.Printf("cloud control plane: mode=%s store=%s interval=%s", mode, envOr("CLOUD_STORE", "mem"), interval)
	if !runAPI {
		<-ctx.Done() // operator-only: run until signalled
		log.Printf("cloud: shutting down")
		return
	}

	// Cloud workstations share the same pod substrate as the control plane.
	wsh := workstation.Handler(workstation.NewManager(pods, envOr("CLOUD_HOST", "localhost")))
	mux := http.NewServeMux()
	mux.Handle("/v1/workstations", wsh)
	mux.Handle("/v1/workstations/", wsh)
	mux.Handle("/", controlplane.Handler(cp))
	addr := envOr("CLOUD_ADDR", ":8086")
	srv := server.New(addr, mux)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("cloud control plane listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cloud: %v", err)
	}
	log.Printf("cloud: shutting down")
}

// planes resolves CLOUD_MODE into which planes to run.
func planes(mode string) (api, operator bool, err error) {
	switch mode {
	case "all":
		return true, true, nil
	case "api":
		return true, false, nil
	case "operator":
		return false, true, nil
	default:
		return false, false, fmt.Errorf("unknown CLOUD_MODE %q (want all|api|operator)", mode)
	}
}

// selectStore picks the desired-state backend (single core, multi store).
func selectStore() (controlplane.Store, error) {
	switch envOr("CLOUD_STORE", "mem") {
	case "mem":
		return controlplane.NewMemStore(), nil
	case "file":
		return controlplane.NewFileStore(envOr("CLOUD_STORE_PATH", "data/cloud-state.json"))
	default:
		return nil, fmt.Errorf("unknown CLOUD_STORE %q (want mem|file)", os.Getenv("CLOUD_STORE"))
	}
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
