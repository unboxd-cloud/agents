package controlplane

import (
	"context"
	"testing"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
	"github.com/unboxd-cloud/platform/internal/kube"
)

func TestReaper_DeletesOrphanPods(t *testing.T) {
	ctx := context.Background()
	pods := kube.NewManager()
	store := NewMemStore()
	cp := NewWithStore(pods, store)

	vm, err := cp.DeployVirtualMachine(ctx, cloudstack.DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cp.Reconcile(ctx); err != nil { // creates the pod
		t.Fatal(err)
	}

	reaper := NewReaper(pods, store)

	// A live VM's pod is left alone.
	if err := reaper.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	if n := podCount(t, ctx, pods, "tenant-t1"); n != 1 {
		t.Fatalf("reaper removed a live pod: %d", n)
	}

	// Orphan the pod by dropping the VM straight from the store (as a crash
	// might), then reap.
	if err := store.Delete(ctx, vm.ID); err != nil {
		t.Fatal(err)
	}
	if err := reaper.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	if n := podCount(t, ctx, pods, "tenant-t1"); n != 0 {
		t.Fatalf("orphan pod not reaped: %d", n)
	}
}
