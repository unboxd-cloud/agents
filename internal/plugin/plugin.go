// Package plugin is the extension seam. It lets adapters, protocol handlers, and
// integrations be added without modifying core code.
//
// Native-runtime path (recommended): a plugin package registers itself in its
// init() via Register; a binary blank-imports that package to compile the
// extension natively into the build (the database/sql driver model). This is
// type-safe, cross-platform, and needs no dynamic loader. (A Go `plugin` .so
// path is possible too, but compile-time registration is the supported default.)
//
// Each extension is a Factory for one Kind of seam (provider, meter source,
// authz gate, tax/compliance source, ...). The caller asserts the built value to
// the seam's interface (e.g. provider.Provider), keeping this package decoupled
// from every other package.
package plugin

import (
	"fmt"
	"sort"
	"sync"
)

// Kind classifies which seam an extension plugs into.
type Kind string

const (
	KindProvider         Kind = "provider"          // infrastructure: provider.Provider
	KindMeterSource      Kind = "meter_source"      // usage ingestion: metering.Source / StreamSource
	KindAuthzGate        Kind = "authz_gate"        // policy gate: authz.Gate
	KindAuthzRelation    Kind = "authz_relation"    // ReBAC: authz.RelationChecker
	KindTaxProvider      Kind = "tax_provider"      // tax rules source: []billing.TaxRule
	KindComplianceSource Kind = "compliance_source" // compliance specs source
	KindProtocol         Kind = "protocol"          // protocol adapter (HTTP, NATS, OTLP, Kafka, ...)
	KindPublishRoute     Kind = "publish_route"     // marketplace publishing route (AWS/GCP/Azure/OpenStack/...)
)

// Factory builds a seam implementation from string config. The returned value is
// asserted to the seam's interface by the caller.
type Factory func(config map[string]string) (any, error)

// Extension is a registered plugin for one seam, identified by Kind+Name.
type Extension struct {
	Kind    Kind
	Name    string
	Factory Factory
}

var (
	mu  sync.RWMutex
	reg = map[Kind]map[string]Extension{}
)

// Register adds an extension. Intended to be called from a plugin package's
// init() so blank-importing the package compiles it into the native runtime.
func Register(e Extension) {
	if e.Name == "" || e.Kind == "" || e.Factory == nil {
		panic("plugin: invalid extension registration")
	}
	mu.Lock()
	defer mu.Unlock()
	if reg[e.Kind] == nil {
		reg[e.Kind] = map[string]Extension{}
	}
	if _, dup := reg[e.Kind][e.Name]; dup {
		panic(fmt.Sprintf("plugin: duplicate %s extension %q", e.Kind, e.Name))
	}
	reg[e.Kind][e.Name] = e
}

// New builds an instance of the named extension for a kind.
func New(kind Kind, name string, config map[string]string) (any, error) {
	mu.RLock()
	e, ok := reg[kind][name]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin: no %s extension named %q", kind, name)
	}
	return e.Factory(config)
}

// List returns the registered extension names for a kind, sorted.
func List(kind Kind) []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(reg[kind]))
	for n := range reg[kind] {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
