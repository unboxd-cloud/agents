// Package workflow lets organizations define publishing and development
// workflows as ordered approval stages. Each stage is approved by a human or an
// LLM approver, and may only use trusted tools/skills/artifacts (from the
// org's trusted registry). This is the governance layer over publishing routes
// and provisioning: nothing ships without passing its workflow.
package workflow

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Kind classifies a workflow.
type Kind string

const (
	KindPublishing  Kind = "publishing"  // publish an offering to a route/marketplace
	KindDevelopment Kind = "development" // build/deploy a change
)

// ApproverType is who approves a stage.
type ApproverType string

const (
	ApproverHuman ApproverType = "human"
	ApproverLLM   ApproverType = "llm" // defaults to an open-source CPU LLM
)

// Approver identifies the approving party for a stage.
type Approver struct {
	Type ApproverType `json:"type"`
	ID   string       `json:"id"` // human subject, or LLM model id
}

// RefKind is the kind of a trusted reference.
type RefKind string

const (
	RefTool     RefKind = "tool"
	RefSkill    RefKind = "skill"
	RefArtifact RefKind = "artifact"
)

// Ref is a tool/skill/artifact reference used within a stage.
type Ref struct {
	Kind RefKind `json:"kind"`
	Name string  `json:"name"`
}

// Decision is the state of a stage / workflow.
type Decision string

const (
	Pending  Decision = "pending"
	Approved Decision = "approved"
	Rejected Decision = "rejected"
)

// Stage is one approval gate.
type Stage struct {
	Name     string   `json:"name"`
	Approver Approver `json:"approver"`
	// Allowed lists the tools/skills/artifacts this stage may use. Every entry
	// must be in the org's TrustedRegistry, for both human and LLM approvers.
	Allowed  []Ref     `json:"allowed"`
	Decision Decision  `json:"decision"`
	By       string    `json:"by,omitempty"`
	At       time.Time `json:"at,omitempty"`
	Note     string    `json:"note,omitempty"`
}

// Workflow is an ordered set of stages for a tenant.
type Workflow struct {
	ID       string   `json:"id"`
	TenantID string   `json:"tenantId"`
	Kind     Kind     `json:"kind"`
	Subject  string   `json:"subject"` // what is being published/developed
	Stages   []Stage  `json:"stages"`
	Decision Decision `json:"decision"` // overall
}

// Errors.
var (
	ErrUntrusted      = errors.New("reference is not in the trusted registry")
	ErrNoStages       = errors.New("workflow has no stages")
	ErrDone           = errors.New("workflow already completed")
	ErrWrongApprover  = errors.New("approver not authorized for this stage")
	ErrToolNotAllowed = errors.New("tool/skill/artifact not allowed in this stage")
)

// TrustedRegistry holds the org's official trusted tools/skills/artifacts. Only
// trusted refs may be used in any stage (by humans or LLMs alike).
type TrustedRegistry struct {
	mu sync.RWMutex
	m  map[RefKind]map[string]bool
}

// NewTrustedRegistry returns an empty registry.
func NewTrustedRegistry() *TrustedRegistry {
	return &TrustedRegistry{m: map[RefKind]map[string]bool{}}
}

// Trust marks a ref as trusted.
func (t *TrustedRegistry) Trust(r Ref) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.m[r.Kind] == nil {
		t.m[r.Kind] = map[string]bool{}
	}
	t.m[r.Kind][r.Name] = true
}

// IsTrusted reports whether a ref is trusted.
func (t *TrustedRegistry) IsTrusted(r Ref) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.m[r.Kind] != nil && t.m[r.Kind][r.Name]
}

// New validates and creates a workflow: it must have stages, and every
// referenced tool/skill/artifact must be trusted.
func New(id, tenantID string, kind Kind, subject string, stages []Stage, trusted *TrustedRegistry) (*Workflow, error) {
	if len(stages) == 0 {
		return nil, ErrNoStages
	}
	for i := range stages {
		for _, ref := range stages[i].Allowed {
			if !trusted.IsTrusted(ref) {
				return nil, fmt.Errorf("%w: %s/%s", ErrUntrusted, ref.Kind, ref.Name)
			}
		}
		stages[i].Decision = Pending
	}
	return &Workflow{ID: id, TenantID: tenantID, Kind: kind, Subject: subject, Stages: stages, Decision: Pending}, nil
}

// Current returns the index of the first pending stage, or -1 if none.
func (w *Workflow) Current() int {
	for i := range w.Stages {
		if w.Stages[i].Decision == Pending {
			return i
		}
	}
	return -1
}

// Decide records a decision on the current stage by the given approver, having
// used the listed refs. It enforces approver identity and the trusted-tool
// allowlist, then advances or completes the workflow.
func (w *Workflow) Decide(approverID string, approve bool, usedRefs []Ref, note string) error {
	if w.Decision != Pending {
		return ErrDone
	}
	idx := w.Current()
	if idx < 0 {
		return ErrDone
	}
	st := &w.Stages[idx]
	if st.Approver.ID != approverID {
		return ErrWrongApprover
	}
	// Used refs must be within this stage's allowed set.
	for _, u := range usedRefs {
		if !refAllowed(st.Allowed, u) {
			return fmt.Errorf("%w: %s/%s", ErrToolNotAllowed, u.Kind, u.Name)
		}
	}
	st.By, st.At, st.Note = approverID, time.Now().UTC(), note
	if !approve {
		st.Decision = Rejected
		w.Decision = Rejected
		return nil
	}
	st.Decision = Approved
	if w.Current() < 0 {
		w.Decision = Approved // all stages approved
	}
	return nil
}

func refAllowed(allowed []Ref, r Ref) bool {
	for _, a := range allowed {
		if a == r {
			return true
		}
	}
	return false
}

// Judge is the pluggable LLM approver (default: an open-source CPU LLM). It
// returns approve/reject + a note for an LLM-typed stage. The engine stays
// independent of any specific model.
type Judge interface {
	Review(w *Workflow, stage Stage) (approve bool, note string, err error)
}

// AutoDecideLLM runs the current stage's LLM approver via the Judge. Human
// stages are untouched. Returns ErrWrongApprover if the current stage is human.
func (w *Workflow) AutoDecideLLM(j Judge) error {
	idx := w.Current()
	if idx < 0 {
		return ErrDone
	}
	st := w.Stages[idx]
	if st.Approver.Type != ApproverLLM {
		return ErrWrongApprover
	}
	approve, note, err := j.Review(w, st)
	if err != nil {
		return err
	}
	return w.Decide(st.Approver.ID, approve, nil, note)
}

// Store persists workflows per tenant (database-agnostic seam).
type Store interface {
	Put(w *Workflow) error
	Get(tenantID, id string) (*Workflow, bool)
	List(tenantID string) []*Workflow
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]*Workflow // key tenant/id
}

// NewMemStore returns an empty Store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]*Workflow{}} }

func key(t, id string) string { return t + "/" + id }

// Put stores a workflow.
func (s *MemStore) Put(w *Workflow) error {
	if w.TenantID == "" || w.ID == "" {
		return errors.New("workflow needs tenantId and id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key(w.TenantID, w.ID)] = w
	return nil
}

// Get returns a workflow.
func (s *MemStore) Get(tenantID, id string) (*Workflow, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.m[key(tenantID, id)]
	return w, ok
}

// List returns a tenant's workflows sorted by id.
func (s *MemStore) List(tenantID string) []*Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Workflow
	for _, w := range s.m {
		if w.TenantID == tenantID {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
