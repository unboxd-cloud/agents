// Package echoprovider is an example extension showing the native-runtime plugin
// path: it implements the provider.Provider seam and self-registers in init().
//
// To include it in a binary, blank-import it once:
//
//	import _ "github.com/unboxd-cloud/platform/plugins/echoprovider"
//
// Then build it via the registry:
//
//	v, _ := plugin.New(plugin.KindProvider, "echo", map[string]string{"name": "edge-1"})
//	p := v.(provider.Provider)
//
// Copy this package as the template for any Go plugin (provider, meter source,
// protocol adapter, tax/compliance source).
package echoprovider

import (
	"context"

	"github.com/unboxd-cloud/platform/internal/plugin"
	"github.com/unboxd-cloud/platform/internal/provider"
)

type echo struct{ name string }

func (e echo) Name() string { return e.name }

func (e echo) Provision(_ context.Context, tenantID string, r provider.Resource) (provider.Instance, error) {
	return provider.Instance{ID: "echo-" + r.Kind, Provider: e.name, Kind: r.Kind, Status: "ready"}, nil
}

func (e echo) Deprovision(_ context.Context, _, _ string) error { return nil }

func init() {
	plugin.Register(plugin.Extension{
		Kind: plugin.KindProvider,
		Name: "echo",
		Factory: func(cfg map[string]string) (any, error) {
			name := cfg["name"]
			if name == "" {
				name = "echo"
			}
			return echo{name: name}, nil
		},
	})
}
