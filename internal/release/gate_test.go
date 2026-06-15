package release

import (
	"testing"

	"github.com/unboxd-cloud/platform/internal/depgraph"
	"github.com/unboxd-cloud/platform/internal/workflow"
)

func trustedReg() *workflow.TrustedRegistry {
	t := workflow.NewTrustedRegistry()
	t.Trust(workflow.Ref{Kind: workflow.RefTool, Name: "trivy"})
	return t
}

func approvedWorkflow(t *testing.T) *workflow.Workflow {
	t.Helper()
	wf, err := workflow.New("w", "acme", workflow.KindPublishing, "offering:s3",
		[]workflow.Stage{{Name: "signoff", Approver: workflow.Approver{Type: workflow.ApproverHuman, ID: "lead@acme"},
			Allowed: []workflow.Ref{{Kind: workflow.RefTool, Name: "trivy"}}}}, trustedReg())
	if err != nil {
		t.Fatal(err)
	}
	if err := wf.Decide("lead@acme", true, []workflow.Ref{{Kind: workflow.RefTool, Name: "trivy"}}, "approved"); err != nil {
		t.Fatal(err)
	}
	return wf
}

func resolvableGraph() *depgraph.Graph {
	g := depgraph.New()
	g.DependsOn("s3", "compute")
	return g
}

func TestGate_PassesWhenAllGreen(t *testing.T) {
	g := Gate{Deps: resolvableGraph(), Known: map[string]bool{"s3": true, "compute": true}}
	res := g.Check("s3", approvedWorkflow(t))
	if !res.OK {
		t.Fatalf("expected pass, got reasons: %v", res.Reasons)
	}
}

func TestGate_FailsOnUnmetDependency(t *testing.T) {
	g := Gate{Deps: resolvableGraph(), Known: map[string]bool{"s3": true}} // compute unmet
	res := g.Check("s3", approvedWorkflow(t))
	if res.OK {
		t.Fatal("expected failure for unmet dependency")
	}
}

func TestGate_FailsWithoutHumanSignoff(t *testing.T) {
	// LLM-only approval should not satisfy the human accountability gate.
	reg := workflow.NewTrustedRegistry()
	wf, _ := workflow.New("w", "acme", workflow.KindPublishing, "x",
		[]workflow.Stage{{Name: "llm", Approver: workflow.Approver{Type: workflow.ApproverLLM, ID: "llama"}}}, reg)
	_ = wf.AutoDecideLLM(autoApprove{})
	g := Gate{Deps: depgraph.New(), Known: map[string]bool{}}
	res := g.Check("x", wf)
	if res.OK {
		t.Fatal("expected failure without human sign-off")
	}
}

func TestGate_FailsOnCycle(t *testing.T) {
	g := depgraph.New()
	g.DependsOn("a", "b")
	g.DependsOn("b", "a")
	res := Gate{Deps: g, Known: map[string]bool{"a": true, "b": true}}.Check("a", approvedWorkflow(t))
	if res.OK {
		t.Fatal("expected failure on dependency cycle")
	}
}

type autoApprove struct{}

func (autoApprove) Review(*workflow.Workflow, workflow.Stage) (bool, string, error) {
	return true, "ok", nil
}
