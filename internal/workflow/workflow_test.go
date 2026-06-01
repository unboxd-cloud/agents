package workflow

import (
	"errors"
	"testing"
)

func trusted() *TrustedRegistry {
	t := NewTrustedRegistry()
	t.Trust(Ref{Kind: RefTool, Name: "trivy"})
	t.Trust(Ref{Kind: RefSkill, Name: "code-review"})
	t.Trust(Ref{Kind: RefArtifact, Name: "sbom"})
	return t
}

func twoStage() []Stage {
	return []Stage{
		{Name: "llm-review", Approver: Approver{Type: ApproverLLM, ID: "llama-cpu"},
			Allowed: []Ref{{Kind: RefSkill, Name: "code-review"}, {Kind: RefArtifact, Name: "sbom"}}},
		{Name: "human-signoff", Approver: Approver{Type: ApproverHuman, ID: "lead@acme"},
			Allowed: []Ref{{Kind: RefTool, Name: "trivy"}}},
	}
}

func TestNew_RejectsUntrustedRef(t *testing.T) {
	stages := []Stage{{Name: "s", Approver: Approver{Type: ApproverHuman, ID: "a"},
		Allowed: []Ref{{Kind: RefTool, Name: "evil"}}}}
	if _, err := New("w1", "acme", KindPublishing, "offering:s3", stages, trusted()); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("want ErrUntrusted, got %v", err)
	}
}

func TestHappyPath_LLMThenHuman(t *testing.T) {
	w, err := New("w1", "acme", KindDevelopment, "deploy:billing", twoStage(), trusted())
	if err != nil {
		t.Fatal(err)
	}
	// Stage 1: LLM approver via Judge.
	if err := w.AutoDecideLLM(approveJudge{}); err != nil {
		t.Fatal(err)
	}
	if w.Decision != Pending {
		t.Fatalf("should still be pending after stage 1, got %s", w.Decision)
	}
	// Stage 2: human approves using a trusted tool.
	if err := w.Decide("lead@acme", true, []Ref{{Kind: RefTool, Name: "trivy"}}, "lgtm"); err != nil {
		t.Fatal(err)
	}
	if w.Decision != Approved {
		t.Fatalf("workflow should be approved, got %s", w.Decision)
	}
}

func TestDecide_WrongApprover(t *testing.T) {
	w, _ := New("w1", "acme", KindPublishing, "x", twoStage(), trusted())
	// current stage approver is the LLM "llama-cpu", not this human.
	if err := w.Decide("someone@acme", true, nil, ""); !errors.Is(err, ErrWrongApprover) {
		t.Fatalf("want ErrWrongApprover, got %v", err)
	}
}

func TestDecide_ToolNotAllowedInStage(t *testing.T) {
	w, _ := New("w1", "acme", KindPublishing, "x", twoStage(), trusted())
	_ = w.AutoDecideLLM(approveJudge{}) // advance to human stage
	// trivy is trusted globally and allowed in stage 2; code-review is not in stage 2's allow-list.
	if err := w.Decide("lead@acme", true, []Ref{{Kind: RefSkill, Name: "code-review"}}, ""); !errors.Is(err, ErrToolNotAllowed) {
		t.Fatalf("want ErrToolNotAllowed, got %v", err)
	}
}

func TestReject_ShortCircuits(t *testing.T) {
	w, _ := New("w1", "acme", KindPublishing, "x", twoStage(), trusted())
	if err := w.AutoDecideLLM(rejectJudge{}); err != nil {
		t.Fatal(err)
	}
	if w.Decision != Rejected {
		t.Fatalf("reject at stage 1 should reject workflow, got %s", w.Decision)
	}
	if err := w.Decide("lead@acme", true, nil, ""); !errors.Is(err, ErrDone) {
		t.Fatalf("want ErrDone after rejection, got %v", err)
	}
}

type approveJudge struct{}

func (approveJudge) Review(*Workflow, Stage) (bool, string, error) { return true, "auto-approved", nil }

type rejectJudge struct{}

func (rejectJudge) Review(*Workflow, Stage) (bool, string, error) {
	return false, "policy violation", nil
}
