package echoprovider

import (
	"context"
	"testing"

	"github.com/unboxd-cloud/platform/internal/plugin"
	"github.com/unboxd-cloud/platform/internal/provider"
)

// TestNativeRuntimePath proves a Go plugin registered at init() can be built via
// the registry and satisfies the real provider.Provider interface.
func TestNativeRuntimePath(t *testing.T) {
	v, err := plugin.New(plugin.KindProvider, "echo", map[string]string{"name": "edge-1"})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := v.(provider.Provider)
	if !ok {
		t.Fatal("extension does not satisfy provider.Provider")
	}
	if p.Name() != "edge-1" {
		t.Errorf("config not applied: %s", p.Name())
	}
	inst, err := p.Provision(context.Background(), "t1", provider.Resource{Kind: "compute"})
	if err != nil {
		t.Fatal(err)
	}
	if inst.Status != "ready" {
		t.Errorf("unexpected instance: %+v", inst)
	}
}
