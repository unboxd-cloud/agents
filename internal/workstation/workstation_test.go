package workstation

import (
	"context"
	"errors"
	"testing"

	"github.com/unboxd-cloud/platform/internal/kube"
)

func TestManager_LaunchMultiPortSingleDesk(t *testing.T) {
	ctx := context.Background()
	pods := kube.NewManager()
	m := NewManager(pods, "localhost")

	ws, err := m.Launch(ctx, LaunchRequest{Account: "t1", Name: "dev", Ports: []int{8080, 3000, 5173}})
	if err != nil {
		t.Fatal(err)
	}
	if ws.State != StateRunning {
		t.Fatalf("want Running, got %s", ws.State)
	}
	if len(ws.Endpoints) != 3 { // multi port
		t.Fatalf("want 3 endpoints, got %v", ws.Endpoints)
	}

	// Single desk: exactly one pod backs the workstation, exposing all ports.
	ps, err := pods.List(ctx, "tenant-t1")
	if err != nil || len(ps) != 1 {
		t.Fatalf("want 1 pod (single desk), got %d (%v)", len(ps), err)
	}
	if len(ps[0].Ports) != 3 {
		t.Fatalf("pod should expose 3 ports, got %v", ps[0].Ports)
	}

	if got, err := m.Get(ctx, ws.ID); err != nil || got.ID != ws.ID {
		t.Fatalf("get: %+v / %v", got, err)
	}

	if err := m.Stop(ctx, ws.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, ws.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after stop, got %v", err)
	}
	if ps, _ := pods.List(ctx, "tenant-t1"); len(ps) != 0 {
		t.Fatalf("pod not removed after stop: %d", len(ps))
	}
}

func TestManager_LaunchDefaults(t *testing.T) {
	ctx := context.Background()
	m := NewManager(kube.NewManager(), "")

	ws, err := m.Launch(ctx, LaunchRequest{Account: "t1", Name: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ws.Ports) != 1 || ws.Ports[0] != defaultPort {
		t.Fatalf("want default port %d, got %v", defaultPort, ws.Ports)
	}
	if ws.Image != defaultImage {
		t.Fatalf("want default image, got %s", ws.Image)
	}
	if len(ws.Endpoints) != 1 || ws.Endpoints[0] != "localhost:8080" {
		t.Fatalf("want localhost:8080 endpoint, got %v", ws.Endpoints)
	}
}

func TestManager_LaunchValidates(t *testing.T) {
	m := NewManager(kube.NewManager(), "")
	if _, err := m.Launch(context.Background(), LaunchRequest{Name: "dev"}); err == nil {
		t.Error("expected error when account is missing")
	}
	if _, err := m.Launch(context.Background(), LaunchRequest{Account: "t1"}); err == nil {
		t.Error("expected error when name is missing")
	}
}
