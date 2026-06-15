package depgraph

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestResolveOrder(t *testing.T) {
	g := New()
	// codespace depends on notebooks + coding-assistant; both depend on compute.
	g.DependsOn("codespace", "notebooks")
	g.DependsOn("codespace", "coding-assistant")
	g.DependsOn("notebooks", "compute")
	g.DependsOn("coding-assistant", "compute")

	order, err := g.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if pos["compute"] > pos["notebooks"] || pos["notebooks"] > pos["codespace"] {
		t.Fatalf("bad topological order: %v", order)
	}
}

func TestResolveDeterministic(t *testing.T) {
	g := New()
	g.Add("b")
	g.Add("a")
	g.Add("c")
	order, _ := g.Resolve()
	if !reflect.DeepEqual(order, []string{"a", "b", "c"}) {
		t.Fatalf("want deterministic [a b c], got %v", order)
	}
}

func TestCycleDetected(t *testing.T) {
	g := New()
	g.DependsOn("a", "b")
	g.DependsOn("b", "a")
	if _, err := g.Resolve(); !errors.Is(err, ErrCycle) {
		t.Fatalf("want ErrCycle, got %v", err)
	}
}

func TestMissingDeps(t *testing.T) {
	g := New()
	g.DependsOn("bedrock", "compute")
	known := map[string]bool{"bedrock": true} // compute not installable
	miss := g.Missing(known)
	if len(miss) != 1 || miss[0] != "compute" {
		t.Fatalf("want [compute], got %v", miss)
	}
}

func TestResolverAgent(t *testing.T) {
	g := New()
	g.DependsOn("codespace", "compute")
	var got []string
	r := &Resolver{Graph: g, Known: map[string]bool{"codespace": true, "compute": true},
		OnOrder: func(o []string) { got = o }}
	if err := r.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "compute" {
		t.Fatalf("resolver did not produce order: %v", got)
	}
}
