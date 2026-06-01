package agentql

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
func (r Result) HasErrors() bool {
	for _, d := range r.Diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Parse lexes and parses src, returning the model and any lexical/syntactic
// diagnostics. It does not perform cross-reference resolution; use Compile for
// the full pipeline.
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
// references, detect duplicates). This is the canonical entry point shared by
// the Go backend and, via WASM, the TypeScript tooling.
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
