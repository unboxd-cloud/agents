package agentdb

import (
	"fmt"
	"sync"
	"time"
)

// Op is the kind of change a proposal requests.
type Op string

const (
	OpCreate Op = "create" // create or update a record
	OpUpdate Op = "update"
	OpRelate Op = "relate" // create or update an edge
)

// State is a proposal's lifecycle state.
type State string

const (
	StateProposed      State = "proposed"
	StateNeedsApproval State = "needs_approval"
	StateApproved      State = "approved"
	StateRejected      State = "rejected"
	StateApplied       State = "applied"
)

// Proposal is a change an agent proposes to the governed state.
type Proposal struct {
	ID     string  `json:"id"`
	Actor  string  `json:"actor"` // the agent proposing
	Op     Op      `json:"op"`
	Record *Record `json:"record,omitempty"` // for create/update
	Edge   *Edge   `json:"edge,omitempty"`   // for relate
	State  State   `json:"state"`
	Reason string  `json:"reason,omitempty"` // governance/approval reason
}

// Verdict is a policy's ruling on a proposal.
type Verdict int

const (
	Allow Verdict = iota
	Deny
	RequireApproval
)

// Decision is a policy verdict with a human-readable reason.
type Decision struct {
	Verdict Verdict
	Reason  string
}

// Policy governs proposals. Implementations must be pure reads: they may read
// the kernel's Store but must not call back into kernel methods that mutate or
// lock (Propose/Approve/...), which would deadlock.
type Policy interface {
	Name() string
	Evaluate(p Proposal, store Store) Decision
}

type policyFunc struct {
	name string
	fn   func(Proposal, Store) Decision
}

func (p policyFunc) Name() string                           { return p.name }
func (p policyFunc) Evaluate(pr Proposal, s Store) Decision { return p.fn(pr, s) }

// PolicyFunc adapts a function to a Policy.
func PolicyFunc(name string, fn func(Proposal, Store) Decision) Policy {
	return policyFunc{name: name, fn: fn}
}

// Event is an audit record of a proposal state transition. Everything the kernel
// does is traceable through these.
type Event struct {
	At         time.Time `json:"at"`
	ProposalID string    `json:"proposalId"`
	Actor      string    `json:"actor"`
	From       State     `json:"from,omitempty"`
	To         State     `json:"to"`
	Policy     string    `json:"policy,omitempty"`
	Note       string    `json:"note,omitempty"`
}

// Kernel is the governed state machine: agents propose, policies govern, humans
// approve, and approved proposals are applied (reconciled) into the Store.
//
//	Agents propose. Policies govern. Humans approve. Workers execute. The store reconciles.
type Kernel struct {
	mu        sync.Mutex
	store     Store
	policies  []Policy
	proposals map[string]Proposal
	events    []Event
	seq       int
}

// NewKernel builds a Kernel over a Store with the given policies (evaluated in
// order).
func NewKernel(store Store, policies ...Policy) *Kernel {
	return &Kernel{store: store, policies: policies, proposals: map[string]Proposal{}}
}

// Store returns the underlying store.
func (k *Kernel) Store() Store { return k.store }

// Propose submits a change. Policies are evaluated in order: any Deny rejects
// immediately; otherwise any RequireApproval parks it as needs_approval; with
// neither, it is auto-approved. The proposal is not applied until Apply.
func (k *Kernel) Propose(p Proposal) (Proposal, error) {
	if p.Actor == "" || p.Op == "" {
		return Proposal{}, ErrInvalid
	}
	switch p.Op {
	case OpCreate, OpUpdate:
		if p.Record == nil {
			return Proposal{}, ErrInvalid
		}
	case OpRelate:
		if p.Edge == nil {
			return Proposal{}, ErrInvalid
		}
	default:
		return Proposal{}, ErrInvalid
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	k.seq++
	if p.ID == "" {
		p.ID = fmt.Sprintf("prop-%d", k.seq)
	}
	p.State = StateProposed
	k.proposals[p.ID] = p
	k.record(p, "", StateProposed, "", "proposed")

	verdict, reason, policy := Allow, "", ""
	for _, pol := range k.policies {
		d := pol.Evaluate(p, k.store)
		if d.Verdict == Deny {
			verdict, reason, policy = Deny, d.Reason, pol.Name()
			break
		}
		if d.Verdict == RequireApproval && verdict == Allow {
			verdict, reason, policy = RequireApproval, d.Reason, pol.Name()
		}
	}

	switch verdict {
	case Deny:
		p.State, p.Reason = StateRejected, reason
		k.record(p, StateProposed, StateRejected, policy, reason)
	case RequireApproval:
		p.State, p.Reason = StateNeedsApproval, reason
		k.record(p, StateProposed, StateNeedsApproval, policy, reason)
	default:
		p.State = StateApproved
		k.record(p, StateProposed, StateApproved, "", "policies allowed")
	}
	k.proposals[p.ID] = p
	return p, nil
}

// Approve approves a proposal awaiting human approval.
func (k *Kernel) Approve(id, approver string) (Proposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	p, ok := k.proposals[id]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	if p.State != StateNeedsApproval {
		return p, fmt.Errorf("proposal %s is %s, not awaiting approval", id, p.State)
	}
	p.State = StateApproved
	p.Reason = "approved by " + approver
	k.proposals[id] = p
	k.record(p, StateNeedsApproval, StateApproved, "", p.Reason)
	return p, nil
}

// Reject rejects a proposal (from any non-terminal state).
func (k *Kernel) Reject(id, approver, reason string) (Proposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	p, ok := k.proposals[id]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	from := p.State
	p.State = StateRejected
	p.Reason = fmt.Sprintf("rejected by %s: %s", approver, reason)
	k.proposals[id] = p
	k.record(p, from, StateRejected, "", p.Reason)
	return p, nil
}

// Apply commits an approved proposal to the store (the worker/reconcile step).
func (k *Kernel) Apply(id string) (Proposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	p, ok := k.proposals[id]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	if p.State != StateApproved {
		return p, fmt.Errorf("proposal %s is %s, not approved", id, p.State)
	}
	var err error
	switch p.Op {
	case OpCreate, OpUpdate:
		_, err = k.store.PutRecord(*p.Record)
	case OpRelate:
		_, err = k.store.PutEdge(*p.Edge)
	}
	if err != nil {
		return p, err
	}
	p.State = StateApplied
	k.proposals[id] = p
	k.record(p, StateApproved, StateApplied, "", "applied to store")
	return p, nil
}

// Proposal returns a proposal by id.
func (k *Kernel) Proposal(id string) (Proposal, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	p, ok := k.proposals[id]
	return p, ok
}

// Events returns a copy of the audit log in order.
func (k *Kernel) Events() []Event {
	k.mu.Lock()
	defer k.mu.Unlock()
	out := make([]Event, len(k.events))
	copy(out, k.events)
	return out
}

// record appends an audit event. Caller must hold k.mu.
func (k *Kernel) record(p Proposal, from, to State, policy, note string) {
	k.events = append(k.events, Event{
		At: time.Now().UTC(), ProposalID: p.ID, Actor: p.Actor,
		From: from, To: to, Policy: policy, Note: note,
	})
}
