package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/unboxd-cloud/platform/pkg/sdk"
)

func TestNewDefaultClient(t *testing.T) {
	p := New()
	if p.Control() == nil {
		t.Fatal("expected a default control-plane client")
	}
}

func TestWithClientAndTenant(t *testing.T) {
	c := sdk.New()
	p := New(WithClient(c), WithTenant("acme"))
	if p.Control() != c {
		t.Error("WithClient should set the client")
	}
	if p.Control().Tenant != "acme" {
		t.Errorf("WithTenant = %q, want acme", p.Control().Tenant)
	}
}

func TestCompileAgent(t *testing.T) {
	p := New()
	if res := p.CompileAgent(`entity A { name: string required }`); res.HasErrors() {
		t.Errorf("valid agent should compile clean: %+v", res.Diagnostics)
	}
	if res := p.CompileAgent(`entity A { x: Missing }`); !res.HasErrors() {
		t.Error("unresolved reference should be an error")
	}
}

func TestLoadAgent(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.agent")
	if err := os.WriteFile(good, []byte("entity A { name: string required }\nrelation R A -> A {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := New()
	agent, _, err := p.LoadAgent(good)
	if err != nil {
		t.Fatalf("LoadAgent(good): %v", err)
	}
	if len(agent.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(agent.Entities))
	}

	bad := filepath.Join(dir, "bad.agent")
	if err := os.WriteFile(bad, []byte("entity A { x: Missing }"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, diags, err := p.LoadAgent(bad); err == nil {
		t.Errorf("LoadAgent(bad) should error; diags=%+v", diags)
	}

	if _, _, err := p.LoadAgent(filepath.Join(dir, "nope.agent")); err == nil {
		t.Error("LoadAgent of a missing file should error")
	}
}
