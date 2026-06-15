// Package community is the platform's community-development engine: the tool that
// taps human intelligence to keep raising the bar. The machine owns excellence
// (deterministic, repeatable conformance); humans own innovation. A closed
// fine-tuning crowd saturates, and real experts won't tune models — so the
// surface is near-zero friction: three everyday gestures (input, suggestion,
// feedback) captured inline, made visible and peer-validated, so one person's
// suggestion seeds another's idea (cross-pollination → synthesis). The asset
// improved is the brain/graph, not the rented cortex.
//
// Significant gestures escalate into a governed innovation funnel — the New
// Product Development (NPD) process — with two honest gates:
//
//	identify → explore → propose → validate → promote → mature → release
//
//   - you may not propose before exploring (no novelty claim without prior art:
//     composition over rebuild), and
//   - an idea is innovation only if it is composable (else it's an unprovable
//     synthesis claim and is not adopted).
//
// A validated idea then graduates through a track-specific publishing process —
// NPD (prototype→mvp→alpha→beta→production) or research publication
// (preprint→submitted→review→accepted→published, à la arXiv/IEEE) — advancing one
// gated stage at a time. The model is generic over the track so each community
// defines its own ordered stages.
package community

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Kind is one of the three everyday gestures.
type Kind string

const (
	Input      Kind = "input"      // a fact the platform did not have
	Suggestion Kind = "suggestion" // a proposed change
	Feedback   Kind = "feedback"   // a judgment on something the platform did
)

// Validation bar: net trust-weighted assessment a contribution or idea must clear
// to be considered validated (or, if negative, rejected).
const DefaultThreshold = 2.0

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalid       = errors.New("invalid")
	ErrNotExplored   = errors.New("explore existing solutions before proposing an idea")
	ErrNotComposable = errors.New("an idea is innovation only if it is a composable composition")
	ErrNotValidated  = errors.New("only a validated idea may be promoted")
	ErrNotMaturing   = errors.New("idea is not in a maturation track")
	ErrUnknownTrack  = errors.New("unknown publishing track")
	ErrTerminal      = errors.New("idea already at the track's terminal stage")
)

// Contributor is a human SME or an agent; both contribute and validate. Trust
// scales the weight their validations carry (propagative trust / TrustGNN).
type Contributor struct {
	Subject   string  `json:"subject"`
	Expertise string  `json:"expertise,omitempty"`
	IsAgent   bool    `json:"isAgent,omitempty"`
	Trust     float64 `json:"trust"`
}

// Contribution is one light gesture, attached inline to a target and, when it
// builds on another, linked by SeedOf — the social edge that lets innovation
// compound across people.
type Contribution struct {
	ID     string    `json:"id"`
	Kind   Kind      `json:"kind"`
	By     string    `json:"by"`
	Target string    `json:"target,omitempty"`
	Body   string    `json:"body"`
	SeedOf string    `json:"seedOf,omitempty"`
	At     time.Time `json:"at"`
}

// Solution is an existing solution found during exploration (prior art).
type Solution struct {
	Name     string `json:"name"`
	Source   string `json:"source,omitempty"`
	Fit      string `json:"fit,omitempty"`
	Reusable bool   `json:"reusable"`
}

// Need is an identified problem: who raised it and the problem statement that
// scopes it. Stage tracks its position in the funnel.
type Need struct {
	ID        string     `json:"id"`
	By        string     `json:"by"`
	Problem   string     `json:"problem"`
	Context   string     `json:"context,omitempty"`
	Priority  string     `json:"priority,omitempty"`
	Stage     string     `json:"stage"`
	Explored  bool       `json:"explored"`
	Solutions []Solution `json:"solutions,omitempty"`
	Gap       string     `json:"gap,omitempty"`
	At        time.Time  `json:"at"`
}

// Assessment is a trust-weighted peer verdict on an idea.
type Assessment struct {
	By         string    `json:"by"`
	Agree      bool      `json:"agree"`
	Confidence float64   `json:"confidence"`
	Weight     float64   `json:"weight"`
	Evidence   string    `json:"evidence,omitempty"`
	At         time.Time `json:"at"`
}

// Step records an idea advancing one gated stage along its track.
type Step struct {
	Stage string    `json:"stage"`
	By    string    `json:"by"`
	At    time.Time `json:"at"`
}

// Idea is the proposed new composition addressing a Need. It moves through
// proposed → validated/rejected → (promoted) maturing → released.
type Idea struct {
	ID          string       `json:"id"`
	NeedID      string       `json:"needId"`
	By          string       `json:"by"`
	Summary     string       `json:"summary"`
	ComposedOf  []string     `json:"composedOf,omitempty"`
	Composable  bool         `json:"composable"`
	State       string       `json:"state"`
	Score       float64      `json:"score"`
	Assessments []Assessment `json:"assessments,omitempty"`
	Track       string       `json:"track,omitempty"`
	Stage       string       `json:"stage,omitempty"`
	History     []Step       `json:"history,omitempty"`
	At          time.Time    `json:"at"`
}

// Idea states.
const (
	Proposed  = "proposed"
	Validated = "validated"
	Rejected  = "rejected"
	Maturing  = "maturing"
	Released  = "released"
)

// Track is a named, ordered publishing process.
type Track struct {
	Name   string   `json:"name"`
	Stages []string `json:"stages"`
}

// Terminal returns the track's final stage.
func (t Track) Terminal() string {
	if len(t.Stages) == 0 {
		return ""
	}
	return t.Stages[len(t.Stages)-1]
}

// next returns the stage after cur, and whether one exists.
func (t Track) next(cur string) (string, bool) {
	for i, s := range t.Stages {
		if s == cur && i+1 < len(t.Stages) {
			return t.Stages[i+1], true
		}
	}
	return "", false
}

// DefaultTracks are the two canonical publishing processes; communities may
// register their own.
func DefaultTracks() map[string]Track {
	return map[string]Track{
		"npd":      {Name: "npd", Stages: []string{"prototype", "mvp", "alpha", "beta", "production"}},
		"research": {Name: "research", Stages: []string{"preprint", "submitted", "review", "accepted", "published"}},
	}
}

// Store is the in-memory community engine (database-agnostic behind this seam).
type Store struct {
	mu           sync.RWMutex
	contributors map[string]Contributor
	contribs     map[string]Contribution
	needs        map[string]Need
	ideas        map[string]Idea
	tracks       map[string]Track
	threshold    float64
	seq          int
}

// NewStore returns an empty engine with the default tracks and validation bar.
func NewStore() *Store {
	return &Store{
		contributors: map[string]Contributor{},
		contribs:     map[string]Contribution{},
		needs:        map[string]Need{},
		ideas:        map[string]Idea{},
		tracks:       DefaultTracks(),
		threshold:    DefaultThreshold,
	}
}

func (s *Store) id(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s-%d", prefix, s.seq)
}

// trustOf returns a contributor's trust weight (1.0 if unknown).
func (s *Store) trustOf(subject string) float64 {
	if c, ok := s.contributors[subject]; ok && c.Trust > 0 {
		return c.Trust
	}
	return 1.0
}

// AddContributor registers (or updates) a contributor, defaulting trust to 1.0.
func (s *Store) AddContributor(c Contributor) (Contributor, error) {
	if strings.TrimSpace(c.Subject) == "" {
		return Contributor{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.Trust <= 0 {
		c.Trust = 1.0
	}
	s.contributors[c.Subject] = c
	return c, nil
}

// Contribute records a light gesture (input/suggestion/feedback). SeedOf links it
// to the contribution that sparked it.
func (s *Store) Contribute(c Contribution) (Contribution, error) {
	if c.By == "" || strings.TrimSpace(c.Body) == "" {
		return Contribution{}, ErrInvalid
	}
	switch c.Kind {
	case Input, Suggestion, Feedback:
	default:
		return Contribution{}, fmt.Errorf("%w: kind %q", ErrInvalid, c.Kind)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.SeedOf != "" {
		if _, ok := s.contribs[c.SeedOf]; !ok {
			return Contribution{}, fmt.Errorf("%w: seed %s", ErrNotFound, c.SeedOf)
		}
	}
	c.ID = s.id("c")
	c.At = time.Now().UTC()
	s.contribs[c.ID] = c
	return c, nil
}

// RaiseNeed identifies a problem and defines its statement (funnel: identify).
// from optionally names a contribution that escalated into this need.
func (s *Store) RaiseNeed(by, problem, context, priority, from string) (Need, error) {
	if by == "" || strings.TrimSpace(problem) == "" {
		return Need{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if from != "" {
		if _, ok := s.contribs[from]; !ok {
			return Need{}, fmt.Errorf("%w: contribution %s", ErrNotFound, from)
		}
	}
	n := Need{
		ID: s.id("need"), By: by, Problem: problem, Context: context,
		Priority: priority, Stage: "identify", At: time.Now().UTC(),
	}
	s.needs[n.ID] = n
	return n, nil
}

// Explore surveys existing solutions for a need (funnel: explore) — the prior-art
// gate that must precede proposing an idea.
func (s *Store) Explore(needID, gap string, solutions ...Solution) (Need, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.needs[needID]
	if !ok {
		return Need{}, fmt.Errorf("%w: need %s", ErrNotFound, needID)
	}
	n.Solutions = append(n.Solutions, solutions...)
	n.Gap = gap
	n.Explored = true
	n.Stage = "explore"
	s.needs[needID] = n
	return n, nil
}

// Propose proposes a new idea against an explored need (funnel: propose). Gated:
// the need must have been explored, and the idea must be composable to count as
// innovation.
func (s *Store) Propose(needID, by, summary string, composedOf []string, composable bool) (Idea, error) {
	if by == "" || strings.TrimSpace(summary) == "" {
		return Idea{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.needs[needID]
	if !ok {
		return Idea{}, fmt.Errorf("%w: need %s", ErrNotFound, needID)
	}
	if !n.Explored {
		return Idea{}, ErrNotExplored
	}
	if !composable {
		return Idea{}, ErrNotComposable
	}
	idea := Idea{
		ID: s.id("idea"), NeedID: needID, By: by, Summary: summary,
		ComposedOf: composedOf, Composable: composable, State: Proposed, At: time.Now().UTC(),
	}
	s.ideas[idea.ID] = idea
	n.Stage = "propose"
	s.needs[needID] = n
	return idea, nil
}

// Assess records a trust-weighted peer verdict on an idea (funnel: validate).
// When the net weighted score crosses the bar the idea becomes validated; if it
// falls below the negative bar it is rejected.
func (s *Store) Assess(ideaID, by string, agree bool, confidence float64, evidence string) (Idea, error) {
	if confidence <= 0 {
		confidence = 1.0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	idea, ok := s.ideas[ideaID]
	if !ok {
		return Idea{}, fmt.Errorf("%w: idea %s", ErrNotFound, ideaID)
	}
	if idea.State != Proposed && idea.State != Validated && idea.State != Rejected {
		return Idea{}, fmt.Errorf("%w: idea %s is %s", ErrInvalid, ideaID, idea.State)
	}
	w := s.trustOf(by) * confidence
	idea.Assessments = append(idea.Assessments, Assessment{
		By: by, Agree: agree, Confidence: confidence, Weight: w, Evidence: evidence, At: time.Now().UTC(),
	})
	idea.Score = 0
	for _, a := range idea.Assessments {
		if a.Agree {
			idea.Score += a.Weight
		} else {
			idea.Score -= a.Weight
		}
	}
	switch {
	case idea.Score >= s.threshold:
		idea.State = Validated
		if n, ok := s.needs[idea.NeedID]; ok {
			n.Stage = "validate"
			s.needs[idea.NeedID] = n
		}
	case idea.Score <= -s.threshold:
		idea.State = Rejected
	default:
		if idea.State != Rejected {
			idea.State = Proposed
		}
	}
	s.ideas[ideaID] = idea
	return idea, nil
}

// Promote adopts a validated idea into a publishing track under governance
// (funnel: promote), starting it at the track's first stage.
func (s *Store) Promote(ideaID, track, approvedBy string) (Idea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idea, ok := s.ideas[ideaID]
	if !ok {
		return Idea{}, fmt.Errorf("%w: idea %s", ErrNotFound, ideaID)
	}
	if idea.State != Validated {
		return Idea{}, ErrNotValidated
	}
	t, ok := s.tracks[track]
	if !ok || len(t.Stages) == 0 {
		return Idea{}, fmt.Errorf("%w: %s", ErrUnknownTrack, track)
	}
	idea.Track = track
	idea.Stage = t.Stages[0]
	idea.State = Maturing
	idea.History = append(idea.History, Step{Stage: idea.Stage, By: approvedBy, At: time.Now().UTC()})
	s.ideas[ideaID] = idea
	if n, ok := s.needs[idea.NeedID]; ok {
		n.Stage = "promote"
		s.needs[idea.NeedID] = n
	}
	return idea, nil
}

// Advance moves a maturing idea to the next gated stage on its track. Reaching
// the terminal stage releases (ships/publishes) it.
func (s *Store) Advance(ideaID, by string) (Idea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idea, ok := s.ideas[ideaID]
	if !ok {
		return Idea{}, fmt.Errorf("%w: idea %s", ErrNotFound, ideaID)
	}
	if idea.State == Released {
		return Idea{}, ErrTerminal
	}
	if idea.State != Maturing {
		return Idea{}, ErrNotMaturing
	}
	t := s.tracks[idea.Track]
	next, ok := t.next(idea.Stage)
	if !ok {
		return Idea{}, ErrTerminal
	}
	idea.Stage = next
	idea.History = append(idea.History, Step{Stage: next, By: by, At: time.Now().UTC()})
	if next == t.Terminal() {
		idea.State = Released
	}
	s.ideas[ideaID] = idea
	return idea, nil
}

// RegisterTrack adds or replaces a publishing track (a community's own process).
func (s *Store) RegisterTrack(t Track) error {
	if strings.TrimSpace(t.Name) == "" || len(t.Stages) == 0 {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks[t.Name] = t
	return nil
}

// Tracks returns the registered publishing tracks, ordered by name.
func (s *Store) Tracks() []Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Track, 0, len(s.tracks))
	for _, t := range s.tracks {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Needs returns all needs, newest first.
func (s *Store) Needs() []Need {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Need, 0, len(s.needs))
	for _, n := range s.needs {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// Ideas returns the ideas proposed against a need (all needs if needID == ""),
// newest first.
func (s *Store) Ideas(needID string) []Idea {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Idea{}
	for _, i := range s.ideas {
		if needID == "" || i.NeedID == needID {
			out = append(out, i)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// Contributions returns the gesture feed, newest first.
func (s *Store) Contributions() []Contribution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Contribution, 0, len(s.contribs))
	for _, c := range s.contribs {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

// Idea returns one idea by id.
func (s *Store) Idea(id string) (Idea, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.ideas[id]
	return i, ok
}
