package provider

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultRegistryHasProviders(t *testing.T) {
	r := DefaultRegistry()
	for _, name := range []string{"kubernetes", "cloudstack", "edge"} {
		if _, err := r.Get(name); err != nil {
			t.Errorf("provider %s missing: %v", name, err)
		}
	}
	if _, err := r.Get("nonexistent"); !errors.Is(err, ErrUnknownProvider) {
		t.Errorf("want ErrUnknownProvider, got %v", err)
	}
}

func TestStubProvisionRequiresTenant(t *testing.T) {
	p := NewCloudStack()
	if _, err := p.Provision(context.Background(), "", Resource{Kind: "compute"}); err == nil {
		t.Error("expected error without tenant")
	}
	inst, err := p.Provision(context.Background(), "t1", Resource{Kind: "compute"})
	if err != nil {
		t.Fatal(err)
	}
	if inst.Provider != "cloudstack" || inst.ID == "" {
		t.Errorf("unexpected instance: %+v", inst)
	}
}
