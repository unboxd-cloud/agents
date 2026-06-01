// Package adl implements ADL, the agent language.
//
// ADL is the executable counterpart to a freeform agent.md / AGENTS.md
// instructions file: instead of natural-language prose, an ADL document declares
// an agent's world in a structured, machine-checked grammar — entities and
// relations, the brains that own runtime areas, the minds and beliefs attached
// to entities, a constitution and policies, HTTP APIs, functions, and
// SurrealDB/SurrealML bindings. This package parses and validates that document
// and projects it into an Agent value the Go backend can act on.
//
// This package is the single source of truth for the language. The TypeScript
// tooling consumes the very same runtime through the WebAssembly build in
// cmd/adl-wasm, so there is exactly one parser and one set of semantics shared
// across the backend and the editor.
//
// Typical use:
//
//	agent, diags := adl.Load(source)
//	if adl.HasErrors(diags) {
//	    // report diags
//	}
//	for _, b := range agent.BeliefsAbout("Lead") { ... }
package adl

import "sort"

// Severity levels for diagnostics.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// Diagnostic is a single problem found while lexing, parsing, or validating.
type Diagnostic struct {
	Severity string   `json:"severity"`
	Message  string   `json:"message"`
	Pos      Position `json:"pos"`
}

// Result is the full outcome of compiling a source string: the parsed model and
// every diagnostic gathered along the way. It is the shape handed to the
// TypeScript tooling (as JSON) by the WASM bridge.
type Result struct {
	Model       *Model       `json:"model"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// HasErrors reports whether any diagnostic is an error (warnings are ignored).
func (r Result) HasErrors() bool { return HasErrors(r.Diagnostics) }

// HasErrors reports whether any diagnostic in the slice is an error.
func HasErrors(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Parse lexes and parses src, returning the model and any lexical/syntactic
// diagnostics. It does not perform cross-reference resolution; use Compile or
// Load for the full pipeline.
func Parse(src string) (*Model, []Diagnostic) {
	lex := newLexer(src)
	toks, lexDiags := lex.tokenize()
	p := newParser(toks)
	model := p.parseModel()
	diags := append(lexDiags, p.diags...)
	sortDiagnostics(diags)
	return model, diags
}

// Compile runs the whole pipeline: lex, parse, and validate (link entity
// references, detect duplicates). It is the canonical entry point shared by the
// Go backend and, via WASM, the TypeScript tooling. For an ergonomic handle on
// the result, use Load instead.
func Compile(src string) Result {
	lex := newLexer(src)
	toks, lexDiags := lex.tokenize()
	p := newParser(toks)
	model := p.parseModel()

	diags := append([]Diagnostic{}, lexDiags...)
	diags = append(diags, p.diags...)
	diags = append(diags, validate(model)...)
	sortDiagnostics(diags)

	return Result{Model: model, Diagnostics: diags}
}

func sortDiagnostics(diags []Diagnostic) {
	sort.SliceStable(diags, func(i, j int) bool {
		a, b := diags[i].Pos, diags[j].Pos
		if a.Offset != b.Offset {
			return a.Offset < b.Offset
		}
		return diags[i].Message < diags[j].Message
	})
}

// Agent is a semantic, queryable view over a compiled ADL document — the
// executable form of an agent definition. Where Model is the raw declaration
// list straight from the parser, Agent groups the declarations by kind and
// indexes them by name, so consumers (Go services and, via WASM, the tooling)
// can ask "give me this entity / the minds for it / the beliefs about it"
// without walking the AST themselves.
//
// One ADL source describes one agent's world, so an Agent is the natural
// top-level handle for an ADL document.
type Agent struct {
	Model *Model `json:"-"`

	Namespaces    []*Namespace        `json:"namespaces,omitempty"`
	Vocabularies  []*Vocabulary       `json:"vocabularies,omitempty"`
	Entities      []*Entity           `json:"entities,omitempty"`
	Relations     []*Relation         `json:"relations,omitempty"`
	Brains        []*Brain            `json:"brains,omitempty"`
	Minds         []*Mind             `json:"minds,omitempty"`
	Beliefs       []*Belief           `json:"beliefs,omitempty"`
	Constitutions []*Constitution     `json:"constitutions,omitempty"`
	Policies      []*Policy           `json:"policies,omitempty"`
	Apis          []*Api              `json:"apis,omitempty"`
	Functions     []*Function         `json:"functions,omitempty"`
	MlBindings    []*SurrealMlBinding `json:"mlBindings,omitempty"`

	entityByName map[string]*Entity
	brainByName  map[string]*Brain
}

// Load compiles ADL source and returns the resulting Agent together with any
// diagnostics. The Agent is always non-nil (empty if the source produced no
// declarations); callers should inspect the diagnostics for errors.
func Load(src string) (*Agent, []Diagnostic) {
	res := Compile(src)
	return NewAgent(res.Model), res.Diagnostics
}

// NewAgent projects a (typically already validated) Model into an Agent. Passing
// a Model from Compile is recommended, so entity references are resolved and
// qualified names are populated.
func NewAgent(model *Model) *Agent {
	a := &Agent{
		Model:        model,
		entityByName: map[string]*Entity{},
		brainByName:  map[string]*Brain{},
	}
	if model == nil {
		return a
	}
	for _, d := range model.Declarations {
		switch n := d.(type) {
		case *Namespace:
			a.Namespaces = append(a.Namespaces, n)
		case *Vocabulary:
			a.Vocabularies = append(a.Vocabularies, n)
		case *Entity:
			a.Entities = append(a.Entities, n)
			a.entityByName[n.Name] = n
			if n.qualified != "" {
				a.entityByName[n.qualified] = n
			}
		case *Relation:
			a.Relations = append(a.Relations, n)
		case *Brain:
			a.Brains = append(a.Brains, n)
			a.brainByName[n.Name] = n
		case *Mind:
			a.Minds = append(a.Minds, n)
		case *Belief:
			a.Beliefs = append(a.Beliefs, n)
		case *Constitution:
			a.Constitutions = append(a.Constitutions, n)
		case *Policy:
			a.Policies = append(a.Policies, n)
		case *Api:
			a.Apis = append(a.Apis, n)
		case *Function:
			a.Functions = append(a.Functions, n)
		case *SurrealMlBinding:
			a.MlBindings = append(a.MlBindings, n)
		}
	}
	return a
}

// Entity looks up an entity by simple or namespace-qualified name.
func (a *Agent) Entity(name string) (*Entity, bool) {
	e, ok := a.entityByName[name]
	return e, ok
}

// Brain looks up a brain by name.
func (a *Agent) Brain(name string) (*Brain, bool) {
	b, ok := a.brainByName[name]
	return b, ok
}

// MindsFor returns the minds whose subject is the named entity (matched by the
// resolved qualified name, falling back to the name as written).
func (a *Agent) MindsFor(entity string) []*Mind {
	target := a.canonicalEntity(entity)
	var out []*Mind
	for _, m := range a.Minds {
		if a.refMatches(&m.Subject, target, entity) {
			out = append(out, m)
		}
	}
	return out
}

// BeliefsAbout returns the beliefs whose subject is the named entity.
func (a *Agent) BeliefsAbout(entity string) []*Belief {
	target := a.canonicalEntity(entity)
	var out []*Belief
	for _, b := range a.Beliefs {
		if a.refMatches(&b.Subject, target, entity) {
			out = append(out, b)
		}
	}
	return out
}

// canonicalEntity returns the qualified name for the given entity name if it is
// known, else the name unchanged.
func (a *Agent) canonicalEntity(name string) string {
	if e, ok := a.entityByName[name]; ok && e.qualified != "" {
		return e.qualified
	}
	return name
}

// refMatches reports whether a reference points at the target entity, comparing
// the resolved qualified name first and the written name as a fallback.
func (a *Agent) refMatches(ref *Reference, qualified, written string) bool {
	if ref.Resolved != "" {
		return ref.Resolved == qualified
	}
	return ref.Name == written || ref.Name == qualified
}
