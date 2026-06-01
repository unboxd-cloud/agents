// Package agentql is the canonical runtime for the AgentQL language.
//
// AgentQL is a declarative DSL (originally authored as a Langium grammar in the
// Agent-Platform project) for describing agentic domain models: entities and
// relations, brains and minds, beliefs, constitutions and policies, HTTP APIs,
// functions, and SurrealDB/SurrealML bindings.
//
// This package is the single source of truth for parsing and validating
// AgentQL. The TypeScript tooling consumes the very same runtime by loading the
// WebAssembly build in cmd/agentql-wasm, so there is exactly one parser and one
// set of semantics shared across the Go backend and the editor tooling.
package agentql

// Position is a 1-based line/column location in the source, plus the 0-based
// byte offset, suitable for editor diagnostics.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

// Model is the grammar entry rule: a sequence of top-level declarations.
type Model struct {
	Declarations []Declaration `json:"declarations"`
}

// Declaration is implemented by every top-level construct. Each concrete node
// carries a "kind" discriminator in its JSON form so the TypeScript side can
// switch on it without reflection.
type Declaration interface {
	isDecl()
	// Loc returns the position of the declaration keyword.
	Loc() Position
}

// declBase is embedded (anonymously) by every declaration so its Kind and Pos
// fields are inlined into the node's JSON object.
type declBase struct {
	Kind string   `json:"kind"`
	Pos  Position `json:"pos"`
}

func (declBase) isDecl()         {}
func (b declBase) Loc() Position { return b.Pos }

// Reference is a cross-reference to an Entity, written as a QualifiedName. After
// validation, Resolved holds the qualified name of the entity it bound to (empty
// if it could not be resolved).
type Reference struct {
	Name     string   `json:"name"`
	Resolved string   `json:"resolved,omitempty"`
	Pos      Position `json:"pos"`
}

// TypeRef is either a primitive type keyword or a reference to an Entity.
type TypeRef struct {
	Primitive string     `json:"primitive,omitempty"`
	Ref       *Reference `json:"ref,omitempty"`
	Pos       Position   `json:"pos"`
}

// Field is a named, typed member of an entity, relation, or mind.
type Field struct {
	Name     string   `json:"name"`
	Type     TypeRef  `json:"type"`
	Required bool     `json:"required"`
	Pos      Position `json:"pos"`
}

// Namespace: 'namespace' QualifiedName
type Namespace struct {
	declBase
	Name string `json:"name"`
}

// Term: 'term' ID ('means' STRING)? ('maps' 'to' QualifiedName)?
type Term struct {
	Name    string   `json:"name"`
	Meaning string   `json:"meaning,omitempty"`
	Mapping string   `json:"mapping,omitempty"`
	Pos     Position `json:"pos"`
}

// Vocabulary: 'vocabulary' ID '{' Term* '}'
type Vocabulary struct {
	declBase
	Name  string `json:"name"`
	Terms []Term `json:"terms"`
}

// Entity: 'entity' ID ('extends' [Entity])? '{' Field* '}'
type Entity struct {
	declBase
	Name   string     `json:"name"`
	Super  *Reference `json:"super,omitempty"`
	Fields []Field    `json:"fields"`

	// qualified is the namespace-qualified name, computed during validation.
	qualified string
}

// Relation: 'relation' ID source -> target '{' Field* '}'
type Relation struct {
	declBase
	Name   string    `json:"name"`
	Source Reference `json:"source"`
	Target Reference `json:"target"`
	Fields []Field   `json:"fields"`
}

// Brain: 'brain' ID '{' Ownership* '}'
type Brain struct {
	declBase
	Name string   `json:"name"`
	Owns []string `json:"owns"`
}

// Mind: 'mind' ID 'for' [Entity] '{' Field* '}'
type Mind struct {
	declBase
	Name    string    `json:"name"`
	Subject Reference `json:"subject"`
	Fields  []Field   `json:"fields"`
}

// Belief: 'belief' ID '{' subject claim (confidence)? (source)? '}'
type Belief struct {
	declBase
	Name       string    `json:"name"`
	Subject    Reference `json:"subject"`
	Claim      string    `json:"claim"`
	Confidence *float64  `json:"confidence,omitempty"`
	Source     string    `json:"source,omitempty"`
}

// Rule: 'rule' ID ':' STRING
type Rule struct {
	Name      string   `json:"name"`
	Statement string   `json:"statement"`
	Pos       Position `json:"pos"`
}

// Constitution: 'constitution' ID '{' Rule* '}'
type Constitution struct {
	declBase
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

// Policy: 'policy' ID '{' Rule* '}'
type Policy struct {
	declBase
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

// Api: 'api' HttpMethod STRING 'returns' SurrealTarget
type Api struct {
	declBase
	Method string `json:"method"`
	Path   string `json:"path"`
	Target string `json:"target"`
}

// Param: ID ':' TypeRef
type Param struct {
	Name string   `json:"name"`
	Type TypeRef  `json:"type"`
	Pos  Position `json:"pos"`
}

// Function: 'function' QualifiedName '(' params ')' 'returns' TypeRef 'as' SurrealTarget
type Function struct {
	declBase
	Name   string  `json:"name"`
	Params []Param `json:"params"`
	Return TypeRef `json:"return"`
	Target string  `json:"target"`
}

// SurrealMlBinding: 'surrealml' ID 'for' InferencePurpose 'input' [Entity] 'output' [Entity]
type SurrealMlBinding struct {
	declBase
	Name    string    `json:"name"`
	Purpose string    `json:"purpose"`
	Input   Reference `json:"input"`
	Output  Reference `json:"output"`
}
