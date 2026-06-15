package orchestrator

import (
	"context"
	"testing"

	"github.com/unboxd-cloud/platform/internal/provider"
)

func TestReconcile_ProvisionsDesiredOnce(t *testing.T) {
	o := New(provider.NewKubernetes())
	o.SetDesired(Desired{
		TenantID: "t1",
		Resources: []provider.Resource{
			{Kind: "kubernetes", Name: "cluster-a"},
			{Kind: "objectstore", Name: "bucket-a"},
		},
	})

	if err := o.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := len(o.Instances()); got != 2 {
		t.Fatalf("want 2 instances, got %d", got)
	}

	// Second reconcile is a no-op (idempotent / level-triggered).
	if err := o.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := len(o.Instances()); got != 2 {
		t.Fatalf("reconcile not idempotent: got %d instances", got)
	}
}
