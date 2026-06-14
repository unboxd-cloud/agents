package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
	"github.com/unboxd-cloud/platform/internal/kube"
)

func TestControlPlane_VMLifecycle(t *testing.T) {
	ctx := context.Background()
	pods := kube.NewManager()
	cp := New(pods)

	vm, err := cp.DeployVirtualMachine(ctx, cloudstack.DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	})
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if vm.State != cloudstack.StateStarting {
		t.Fatalf("want Starting after deploy, got %s", vm.State)
	}
	// No pod until the reconcile loop runs (desired state recorded only).
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 0 {
		t.Fatalf("want 0 pods before reconcile, got %d", got)
	}

	// Reconcile creates the pod and the VM converges to Running.
	mustReconcile(t, ctx, cp)
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 1 {
		t.Fatalf("want 1 pod after reconcile, got %d", got)
	}
	if got, _ := cp.GetVirtualMachine(ctx, vm.ID); got.State != cloudstack.StateRunning {
		t.Fatalf("want Running after reconcile, got %s", got.State)
	}

	// Reconcile is idempotent.
	mustReconcile(t, ctx, cp)
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 1 {
		t.Fatalf("reconcile not idempotent: %d pods", got)
	}

	// Stop removes the pod; Start brings it back.
	if _, err := cp.StopVirtualMachine(ctx, vm.ID); err != nil {
		t.Fatal(err)
	}
	mustReconcile(t, ctx, cp)
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 0 {
		t.Fatalf("want 0 pods after stop, got %d", got)
	}
	if got, _ := cp.GetVirtualMachine(ctx, vm.ID); got.State != cloudstack.StateStopped {
		t.Fatalf("want Stopped, got %s", got.State)
	}
	if _, err := cp.StartVirtualMachine(ctx, vm.ID); err != nil {
		t.Fatal(err)
	}
	mustReconcile(t, ctx, cp)
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 1 {
		t.Fatalf("want 1 pod after start, got %d", got)
	}

	// Destroy tears down the VM and its pod.
	if err := cp.DestroyVirtualMachine(ctx, vm.ID); err != nil {
		t.Fatal(err)
	}
	mustReconcile(t, ctx, cp)
	if vms, _ := cp.ListVirtualMachines(ctx, "t1"); len(vms) != 0 {
		t.Fatalf("want 0 vms after destroy, got %d", len(vms))
	}
	if got := podCount(t, ctx, pods, "tenant-t1"); got != 0 {
		t.Fatalf("want 0 pods after destroy, got %d", got)
	}
}

func TestControlPlane_DeployValidation(t *testing.T) {
	ctx := context.Background()
	cp := New(kube.NewManager())

	if _, err := cp.DeployVirtualMachine(ctx, cloudstack.DeployVMRequest{Account: "t1", Name: "x"}); err == nil {
		t.Error("expected validation error for incomplete request")
	}
	_, err := cp.DeployVirtualMachine(ctx, cloudstack.DeployVMRequest{
		Account: "t1", Name: "x", ZoneID: "zone-1", TemplateID: "nope", ServiceOfferingID: "so-small",
	})
	if !errors.Is(err, cloudstack.ErrNotFound) {
		t.Errorf("want ErrNotFound for unknown template, got %v", err)
	}
}

func TestControlPlane_UnknownVM(t *testing.T) {
	ctx := context.Background()
	cp := New(kube.NewManager())
	if _, err := cp.StartVirtualMachine(ctx, "nope"); !errors.Is(err, cloudstack.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
	if err := cp.DestroyVirtualMachine(ctx, "nope"); !errors.Is(err, cloudstack.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestControlPlane_DirectDelivery(t *testing.T) {
	ctx := context.Background()
	pods := kube.NewManager()
	cp := NewWithStore(pods, NewMemStore())

	// One call: deploy + reconcile inline -> running VM with its pod.
	vm, err := cp.DeliverVirtualMachine(ctx, cloudstack.DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	})
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if vm.State != cloudstack.StateRunning {
		t.Fatalf("direct delivery should return Running, got %s", vm.State)
	}
	if n := podCount(t, ctx, pods, "tenant-t1"); n != 1 {
		t.Fatalf("want 1 pod after direct delivery, got %d", n)
	}
}

func mustReconcile(t *testing.T, ctx context.Context, cp *ControlPlane) {
	t.Helper()
	if err := cp.Reconcile(ctx); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
}

func podCount(t *testing.T, ctx context.Context, pods kube.PodManager, ns string) int {
	t.Helper()
	ps, err := pods.List(ctx, ns)
	if err != nil {
		t.Fatalf("list pods: %v", err)
	}
	return len(ps)
}
