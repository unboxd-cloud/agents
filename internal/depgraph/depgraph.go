// Package depgraph is a native (stdlib-only) dependency tracker and resolver.
// It tracks dependencies between platform units (offerings, modules,
// requirements) and resolves a safe install/build order via topological sort
// with cycle detection. It backs the dependency-resolver agent and the
// orchestrator's provisioning order — kept in-codebase, no external resolver.
package depgraph

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// ErrCycle is returned when the graph cannot be ordered due to a cycle.
var ErrCycle = errors.New("dependency cycle detected")

// Graph is a directed dependency graph: an edge A->B means "A depends on B"
// (B must come first).
type Graph struct {
	mu    sync.RWMutex
	deps  map[string]map[string]bool // node -> set of deps
	nodes map[string]bool
}

// New returns an empty graph.
func New() *Graph {
	return &Graph{deps: map[string]map[string]bool{}, nodes: map[string]bool{}}
}

// Add registers a node (no-op if present).
func (g *Graph) Add(node string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ensure(node)
}

// DependsOn records that node depends on dep (both are created if needed).
func (g *Graph) DependsOn(node, dep string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ensure(node)
	g.ensure(dep)
	g.deps[node][dep] = true
}

func (g *Graph) ensure(n string) {
	if !g.nodes[n] {
		g.nodes[n] = true
		g.deps[n] = map[string]bool{}
	}
}

// Missing returns nodes referenced as dependencies but never added explicitly as
// real nodes — i.e. unresolved/unknown dependencies. (All deps are auto-added as
// nodes, so this reports deps that have no own entry beyond being a target.)
func (g *Graph) Missing(known map[string]bool) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	miss := map[string]bool{}
	for _, ds := range g.deps {
		for d := range ds {
			if known != nil && !known[d] {
				miss[d] = true
			}
		}
	}
	return sortedKeys(miss)
}

// Resolve returns a topological order (dependencies before dependents). It is
// deterministic (ties broken alphabetically) and returns ErrCycle on a cycle.
func (g *Graph) Resolve() ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	const (
		white = 0 // unvisited
		gray  = 1 // in progress
		black = 2 // done
	)
	color := map[string]int{}
	var order []string
	var visit func(n string) error
	visit = func(n string) error {
		switch color[n] {
		case gray:
			return fmt.Errorf("%w: at %q", ErrCycle, n)
		case black:
			return nil
		}
		color[n] = gray
		for _, d := range sortedKeys(g.deps[n]) {
			if err := visit(d); err != nil {
				return err
			}
		}
		color[n] = black
		order = append(order, n)
		return nil
	}
	for _, n := range sortedKeys(g.nodes) {
		if err := visit(n); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
