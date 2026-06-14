package kube

import (
	"context"
	"errors"
	"testing"
)

func TestManager_CreateGetListDelete(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	spec := PodSpec{Namespace: "tenant-t1", Name: "web-1", Image: "nginx", CPUNumber: 1, MemoryMB: 256}
	pod, err := m.Create(ctx, spec)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if pod.Phase != PodRunning || pod.UID == "" || pod.Node == "" {
		t.Fatalf("unexpected pod: %+v", pod)
	}

	// Create is idempotent: same namespace/name returns the existing pod.
	again, err := m.Create(ctx, spec)
	if err != nil || again.UID != pod.UID {
		t.Fatalf("create not idempotent: %+v / %v", again, err)
	}

	got, err := m.Get(ctx, "tenant-t1", "web-1")
	if err != nil || got.UID != pod.UID {
		t.Fatalf("get: %+v / %v", got, err)
	}

	pods, err := m.List(ctx, "tenant-t1")
	if err != nil || len(pods) != 1 {
		t.Fatalf("list: %d pods, %v", len(pods), err)
	}

	if err := m.Delete(ctx, "tenant-t1", "web-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := m.Get(ctx, "tenant-t1", "web-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
	if err := m.Delete(ctx, "tenant-t1", "web-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound deleting missing pod, got %v", err)
	}
}

func TestManager_CreateValidates(t *testing.T) {
	m := NewManager()
	if _, err := m.Create(context.Background(), PodSpec{Namespace: "tenant-t1", Name: "x"}); err == nil {
		t.Error("expected error when image is missing")
	}
	if _, err := m.Create(context.Background(), PodSpec{Image: "nginx"}); err == nil {
		t.Error("expected error when namespace/name are missing")
	}
}
