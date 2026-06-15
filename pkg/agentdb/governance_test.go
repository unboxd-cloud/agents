package agentdb

import "testing"

func proposeCreate(actor, id, kind string) Proposal {
	return Proposal{Actor: actor, Op: OpCreate, Record: &Record{ID: id, Kind: kind}}
}

func TestKernelAutoApproveAndApply(t *testing.T) {
	k := NewKernel(NewMemStore())
	p, err := k.Propose(proposeCreate("orchestrator", "agent:1", "agent"))
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if p.State != StateApproved {
		t.Fatalf("expected auto-approved with no policies, got %s", p.State)
	}
	applied, err := k.Apply(p.ID)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.State != StateApplied {
		t.Fatalf("expected applied, got %s", applied.State)
	}
	if _, ok := k.Store().GetRecord("agent:1"); !ok {
		t.Error("record should be reconciled into the store after Apply")
	}
}

func TestKernelDeny(t *testing.T) {
	deny := PolicyFunc("no-humans", func(p Proposal, _ Store) Decision {
		if p.Record != nil && p.Record.Kind == "human" {
			return Decision{Verdict: Deny, Reason: "agents may not create humans"}
		}
		return Decision{Verdict: Allow}
	})
	k := NewKernel(NewMemStore(), deny)
	p, _ := k.Propose(proposeCreate("rogue", "human:1", "human"))
	if p.State != StateRejected {
		t.Fatalf("expected rejected, got %s (%s)", p.State, p.Reason)
	}
	if _, err := k.Apply(p.ID); err == nil {
		t.Error("Apply of a rejected proposal should fail")
	}
}

func TestKernelRequireApprovalFlow(t *testing.T) {
	gate := PolicyFunc("sensitive", func(p Proposal, _ Store) Decision {
		if p.Op == OpCreate && p.Record.Kind == "policy" {
			return Decision{Verdict: RequireApproval, Reason: "policy changes need human approval"}
		}
		return Decision{Verdict: Allow}
	})
	k := NewKernel(NewMemStore(), gate)
	p, _ := k.Propose(proposeCreate("orchestrator", "policy:1", "policy"))
	if p.State != StateNeedsApproval {
		t.Fatalf("expected needs_approval, got %s", p.State)
	}
	// Cannot apply until approved.
	if _, err := k.Apply(p.ID); err == nil {
		t.Error("Apply before approval should fail")
	}
	if _, err := k.Approve(p.ID, "alice@platform"); err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if _, err := k.Apply(p.ID); err != nil {
		t.Fatalf("Apply after approval: %v", err)
	}
	if _, ok := k.Store().GetRecord("policy:1"); !ok {
		t.Error("approved+applied record should be in the store")
	}
}

func TestKernelRelate(t *testing.T) {
	k := NewKernel(NewMemStore())
	p, err := k.Propose(Proposal{Actor: "orchestrator", Op: OpRelate,
		Edge: &Edge{ID: "e1", Kind: "supervises", From: "agent:1", To: "agent:2"}})
	if err != nil || p.State != StateApproved {
		t.Fatalf("Propose relate: %v state=%s", err, p.State)
	}
	if _, err := k.Apply(p.ID); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if e := k.Store().Edges("agent:1"); len(e) != 1 {
		t.Errorf("edge should be reconciled, got %+v", e)
	}
}

func TestKernelRejectAndAudit(t *testing.T) {
	k := NewKernel(NewMemStore())
	p, _ := k.Propose(proposeCreate("orchestrator", "agent:9", "agent"))
	if _, err := k.Reject(p.ID, "bob@platform", "not now"); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	got, _ := k.Proposal(p.ID)
	if got.State != StateRejected {
		t.Errorf("expected rejected, got %s", got.State)
	}
	// Audit log: proposed -> approved -> rejected (3 transitions recorded).
	ev := k.Events()
	if len(ev) < 2 {
		t.Fatalf("expected an audit trail, got %d events", len(ev))
	}
	if ev[0].To != StateProposed {
		t.Errorf("first event should be 'proposed', got %s", ev[0].To)
	}
	if ev[len(ev)-1].To != StateRejected {
		t.Errorf("last event should be 'rejected', got %s", ev[len(ev)-1].To)
	}
}

func TestKernelValidates(t *testing.T) {
	k := NewKernel(NewMemStore())
	if _, err := k.Propose(Proposal{Actor: "", Op: OpCreate}); err == nil {
		t.Error("empty actor should be invalid")
	}
	if _, err := k.Propose(Proposal{Actor: "a", Op: OpCreate}); err == nil {
		t.Error("create without a record should be invalid")
	}
	if _, err := k.Apply("nope"); err == nil {
		t.Error("Apply of unknown proposal should error")
	}
}
