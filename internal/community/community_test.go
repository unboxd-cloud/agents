package community

import (
	"errors"
	"testing"
)

// seed a store with two trusted reviewers so a single assess can cross the bar.
func seeded(t *testing.T) *Store {
	t.Helper()
	s := NewStore()
	for _, c := range []Contributor{
		{Subject: "alice", Expertise: "sre", Trust: 1.5},
		{Subject: "bob", Expertise: "ml", Trust: 1.5},
		{Subject: "carol", Trust: 1.0},
	} {
		if _, err := s.AddContributor(c); err != nil {
			t.Fatalf("add contributor: %v", err)
		}
	}
	return s
}

func TestProposeRequiresExploration(t *testing.T) {
	s := seeded(t)
	n, err := s.RaiseNeed("alice", "model cost drifts daily", "ctx", "high", "")
	if err != nil {
		t.Fatalf("raise need: %v", err)
	}
	if n.Stage != "identify" {
		t.Fatalf("new need stage = %q, want identify", n.Stage)
	}
	// Proposing before exploring prior art must be blocked.
	if _, err := s.Propose(n.ID, "bob", "new router", []string{"bandit"}, true); err != ErrNotExplored {
		t.Fatalf("propose before explore = %v, want ErrNotExplored", err)
	}
	if _, err := s.Explore(n.ID, "none fit our dataset",
		Solution{Name: "LiteLLM", Source: "oss", Fit: "partial", Reusable: true}); err != nil {
		t.Fatalf("explore: %v", err)
	}
	if _, err := s.Propose(n.ID, "bob", "new router", []string{"LiteLLM", "bandit"}, true); err != nil {
		t.Fatalf("propose after explore: %v", err)
	}
}

func TestProposeRequiresComposable(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "p", "", "", "")
	_, _ = s.Explore(n.ID, "gap")
	// A non-composable claim is not innovation and must be refused.
	if _, err := s.Propose(n.ID, "bob", "magic", nil, false); err != ErrNotComposable {
		t.Fatalf("non-composable propose = %v, want ErrNotComposable", err)
	}
}

func TestTrustWeightedValidation(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "p", "", "", "")
	_, _ = s.Explore(n.ID, "gap")
	idea, _ := s.Propose(n.ID, "bob", "new composition", []string{"a", "b"}, true)

	// One low-trust agree (weight 1.0) is below the 2.0 bar — stays proposed.
	got, err := s.Assess(idea.ID, "carol", true, 1.0, "")
	if err != nil {
		t.Fatalf("assess: %v", err)
	}
	if got.State != Proposed {
		t.Fatalf("after one weak agree state = %q, want proposed (score %.2f)", got.State, got.Score)
	}
	// A trusted agree (1.5) pushes net weight to 2.5 ≥ 2.0 — validated.
	got, _ = s.Assess(idea.ID, "alice", true, 1.0, "matches benchmark")
	if got.State != Validated {
		t.Fatalf("state = %q, want validated (score %.2f)", got.State, got.Score)
	}
}

func TestRejectionOnNegativeConsensus(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "p", "", "", "")
	_, _ = s.Explore(n.ID, "gap")
	idea, _ := s.Propose(n.ID, "bob", "weak idea", []string{"x"}, true)
	_, _ = s.Assess(idea.ID, "alice", false, 1.0, "no evidence")
	got, _ := s.Assess(idea.ID, "bob", false, 1.0, "duplicate")
	if got.State != Rejected {
		t.Fatalf("state = %q, want rejected (score %.2f)", got.State, got.Score)
	}
}

func TestPromoteOnlyValidated(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "p", "", "", "")
	_, _ = s.Explore(n.ID, "gap")
	idea, _ := s.Propose(n.ID, "bob", "idea", []string{"a"}, true)
	if _, err := s.Promote(idea.ID, "npd", "owner"); err != ErrNotValidated {
		t.Fatalf("promote unvalidated = %v, want ErrNotValidated", err)
	}
	_, _ = s.Assess(idea.ID, "alice", true, 1.0, "")
	_, _ = s.Assess(idea.ID, "bob", true, 1.0, "")
	if _, err := s.Promote(idea.ID, "no-such-track", "owner"); !errors.Is(err, ErrUnknownTrack) {
		t.Fatalf("unknown track = %v, want ErrUnknownTrack", err)
	}
	got, err := s.Promote(idea.ID, "npd", "owner")
	if err != nil {
		t.Fatalf("promote: %v", err)
	}
	if got.State != Maturing || got.Stage != "prototype" {
		t.Fatalf("promoted idea = %s/%s, want maturing/prototype", got.State, got.Stage)
	}
}

func TestNPDMaturationToRelease(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "p", "", "", "")
	_, _ = s.Explore(n.ID, "gap")
	idea, _ := s.Propose(n.ID, "bob", "idea", []string{"a"}, true)
	_, _ = s.Assess(idea.ID, "alice", true, 1.0, "")
	_, _ = s.Assess(idea.ID, "bob", true, 1.0, "")
	cur, _ := s.Promote(idea.ID, "npd", "owner")

	want := []string{"mvp", "alpha", "beta", "production"}
	for _, stage := range want {
		var err error
		cur, err = s.Advance(idea.ID, "owner")
		if err != nil {
			t.Fatalf("advance to %s: %v", stage, err)
		}
		if cur.Stage != stage {
			t.Fatalf("stage = %q, want %q", cur.Stage, stage)
		}
	}
	if cur.State != Released {
		t.Fatalf("final state = %q, want released", cur.State)
	}
	// No advancing past the terminal stage.
	if _, err := s.Advance(idea.ID, "owner"); err != ErrTerminal {
		t.Fatalf("advance past terminal = %v, want ErrTerminal", err)
	}
	if got := len(cur.History); got != 5 {
		t.Fatalf("history steps = %d, want 5", got)
	}
}

func TestResearchTrack(t *testing.T) {
	s := seeded(t)
	n, _ := s.RaiseNeed("alice", "novel eval", "", "", "")
	_, _ = s.Explore(n.ID, "no prior benchmark")
	idea, _ := s.Propose(n.ID, "bob", "CLEAR-X", []string{"CLEAR"}, true)
	_, _ = s.Assess(idea.ID, "alice", true, 1.0, "")
	_, _ = s.Assess(idea.ID, "bob", true, 1.0, "")
	cur, _ := s.Promote(idea.ID, "research", "chair")
	if cur.Stage != "preprint" {
		t.Fatalf("research start = %q, want preprint", cur.Stage)
	}
	for cur.State != Released {
		var err error
		cur, err = s.Advance(idea.ID, "chair")
		if err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	if cur.Stage != "published" {
		t.Fatalf("research terminal = %q, want published", cur.Stage)
	}
}

func TestGestureSeeding(t *testing.T) {
	s := seeded(t)
	first, err := s.Contribute(Contribution{Kind: Suggestion, By: "alice", Target: "router", Body: "try a bandit"})
	if err != nil {
		t.Fatalf("contribute: %v", err)
	}
	// One person's suggestion seeds another's — the social edge.
	second, err := s.Contribute(Contribution{Kind: Suggestion, By: "bob", Body: "make it cost-aware", SeedOf: first.ID})
	if err != nil {
		t.Fatalf("seeded contribute: %v", err)
	}
	if second.SeedOf != first.ID {
		t.Fatalf("seedOf = %q, want %q", second.SeedOf, first.ID)
	}
	// A significant gesture escalates into a governed need.
	if _, err := s.RaiseNeed("bob", "cost-aware routing", "", "high", first.ID); err != nil {
		t.Fatalf("escalate: %v", err)
	}
	if _, err := s.Contribute(Contribution{Kind: Suggestion, By: "x", Body: "b", SeedOf: "nope"}); err == nil {
		t.Fatalf("seed of missing contribution should fail")
	}
}

func TestUnknownKindRejected(t *testing.T) {
	s := seeded(t)
	if _, err := s.Contribute(Contribution{Kind: "rant", By: "a", Body: "b"}); err == nil {
		t.Fatalf("unknown gesture kind should be rejected")
	}
}
