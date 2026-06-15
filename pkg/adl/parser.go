package adl

import (
	"fmt"
	"strconv"
)

// parser is a hand-written recursive-descent parser that follows the ADL
// Langium grammar rule for rule. Each top-level declaration is parsed inside a
// recover boundary so a syntax error in one declaration is reported and then
// skipped, letting the rest of the file still parse (important for editor use).
type parser struct {
	toks  []token
	pos   int
	diags []Diagnostic
}

func newParser(toks []token) *parser { return &parser{toks: toks} }

// primitives are the PrimitiveType keywords. They win over an entity reference
// when they appear in a TypeRef position, matching Langium keyword precedence.
var primitives = map[string]bool{
	"string": true, "int": true, "float": true, "bool": true,
	"datetime": true, "object": true, "array": true, "record": true,
}

var runtimeAreas = map[string]bool{
	"data": true, "business_logic": true, "policy": true, "api": true,
	"permissions": true, "events": true, "live_queries": true, "memory": true,
	"beliefs": true, "decisions": true,
}

var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true, "delete": true,
}

var inferencePurposes = map[string]bool{
	"classification": true, "scoring": true, "prediction": true,
	"ranking": true, "embedding": true,
}

var surrealTargets = map[string]bool{"surrealdb": true, "surrealml": true}

// declKeywords are the leading keywords for each top-level declaration; used for
// dispatch and for error-recovery resynchronisation.
var declKeywords = map[string]bool{
	"namespace": true, "vocabulary": true, "entity": true, "relation": true,
	"brain": true, "mind": true, "belief": true, "constitution": true,
	"policy": true, "api": true, "function": true, "surrealml": true,
}

// parseError is the panic payload used to unwind to the nearest declaration
// boundary on a syntax error.
type parseError struct{ d Diagnostic }

func (p *parser) cur() token  { return p.toks[p.pos] }
func (p *parser) atEnd() bool { return p.cur().Kind == tEOF }

func (p *parser) next() token {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

// isKeyword reports whether the current token is the given keyword.
func (p *parser) isKeyword(kw string) bool {
	t := p.cur()
	return t.Kind == tIdent && t.Value == kw
}

// acceptKeyword consumes the current token if it is the given keyword.
func (p *parser) acceptKeyword(kw string) bool {
	if p.isKeyword(kw) {
		p.next()
		return true
	}
	return false
}

func (p *parser) failf(pos Position, format string, args ...any) {
	panic(parseError{Diagnostic{Severity: SeverityError, Message: fmt.Sprintf(format, args...), Pos: pos}})
}

// expectKeyword consumes the given keyword or fails.
func (p *parser) expectKeyword(kw string) {
	if !p.acceptKeyword(kw) {
		p.failf(p.cur().Pos, "expected '%s' but found %s", kw, p.cur().describe())
	}
}

// expect consumes a token of the given kind or fails, returning it.
func (p *parser) expect(kind tokenKind) token {
	if p.cur().Kind != kind {
		p.failf(p.cur().Pos, "expected %s but found %s", kind, p.cur().describe())
	}
	return p.next()
}

// expectID consumes an identifier (that is not used here as a structural
// keyword) and returns its text.
func (p *parser) expectID() (string, Position) {
	t := p.expect(tIdent)
	return t.Value, t.Pos
}

func (p *parser) parseModel() *Model {
	model := &Model{}
	for !p.atEnd() {
		start := p.pos
		p.parseDeclarationSafely(model)
		if p.pos == start {
			// Guarantee forward progress even if recovery left us put.
			p.next()
		}
	}
	return model
}

// parseDeclarationSafely parses one declaration, recovering from a parseError by
// recording it and resynchronising to the next declaration keyword.
func (p *parser) parseDeclarationSafely(model *Model) {
	defer func() {
		if r := recover(); r != nil {
			pe, ok := r.(parseError)
			if !ok {
				panic(r)
			}
			p.diags = append(p.diags, pe.d)
			p.resync()
		}
	}()
	if d := p.parseDeclaration(); d != nil {
		model.Declarations = append(model.Declarations, d)
	}
}

// resync skips tokens until the next top-level declaration keyword or EOF.
func (p *parser) resync() {
	for !p.atEnd() {
		t := p.cur()
		if t.Kind == tIdent && declKeywords[t.Value] {
			return
		}
		p.next()
	}
}

func (p *parser) parseDeclaration() Declaration {
	t := p.cur()
	if t.Kind != tIdent {
		p.failf(t.Pos, "expected a declaration but found %s", t.describe())
	}
	switch t.Value {
	case "namespace":
		return p.parseNamespace()
	case "vocabulary":
		return p.parseVocabulary()
	case "entity":
		return p.parseEntity()
	case "relation":
		return p.parseRelation()
	case "brain":
		return p.parseBrain()
	case "mind":
		return p.parseMind()
	case "belief":
		return p.parseBelief()
	case "constitution":
		return p.parseConstitution()
	case "policy":
		return p.parsePolicy()
	case "api":
		return p.parseApi()
	case "function":
		return p.parseFunction()
	case "surrealml":
		return p.parseSurrealMlBinding()
	default:
		p.failf(t.Pos, "unknown declaration keyword '%s'", t.Value)
		return nil
	}
}

// QualifiedName: ID ('.' ID)*
func (p *parser) parseQualifiedName() (string, Position) {
	name, pos := p.expectID()
	for p.cur().Kind == tDot {
		p.next()
		seg, _ := p.expectID()
		name += "." + seg
	}
	return name, pos
}

func (p *parser) parseReference() Reference {
	name, pos := p.parseQualifiedName()
	return Reference{Name: name, Pos: pos}
}

func (p *parser) parseNamespace() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("namespace")
	name, _ := p.parseQualifiedName()
	return &Namespace{declBase: declBase{Kind: "Namespace", Pos: pos}, Name: name}
}

func (p *parser) parseVocabulary() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("vocabulary")
	name, _ := p.expectID()
	p.expect(tLBrace)
	v := &Vocabulary{declBase: declBase{Kind: "Vocabulary", Pos: pos}, Name: name}
	for p.isKeyword("term") {
		v.Terms = append(v.Terms, p.parseTerm())
	}
	p.expect(tRBrace)
	return v
}

// Term: 'term' ID ('means' STRING)? ('maps' 'to' QualifiedName)?
func (p *parser) parseTerm() Term {
	pos := p.cur().Pos
	p.expectKeyword("term")
	name, _ := p.expectID()
	t := Term{Name: name, Pos: pos}
	if p.acceptKeyword("means") {
		t.Meaning = p.expect(tString).Value
	}
	if p.acceptKeyword("maps") {
		p.expectKeyword("to")
		t.Mapping, _ = p.parseQualifiedName()
	}
	return t
}

func (p *parser) parseEntity() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("entity")
	name, _ := p.expectID()
	e := &Entity{declBase: declBase{Kind: "Entity", Pos: pos}, Name: name}
	if p.acceptKeyword("extends") {
		ref := p.parseReference()
		e.Super = &ref
	}
	p.expect(tLBrace)
	for p.cur().Kind == tIdent {
		e.Fields = append(e.Fields, p.parseField())
	}
	p.expect(tRBrace)
	return e
}

// Field: ID ':' TypeRef ('required')?
func (p *parser) parseField() Field {
	name, pos := p.expectID()
	p.expect(tColon)
	f := Field{Name: name, Type: p.parseTypeRef(), Pos: pos}
	if p.acceptKeyword("required") {
		f.Required = true
	}
	return f
}

// TypeRef: PrimitiveType | [Entity:QualifiedName]
func (p *parser) parseTypeRef() TypeRef {
	t := p.cur()
	if t.Kind == tIdent && primitives[t.Value] {
		p.next()
		return TypeRef{Primitive: t.Value, Pos: t.Pos}
	}
	ref := p.parseReference()
	return TypeRef{Ref: &ref, Pos: ref.Pos}
}

// Relation: 'relation' ID source -> target '{' Field* '}'
func (p *parser) parseRelation() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("relation")
	name, _ := p.expectID()
	r := &Relation{declBase: declBase{Kind: "Relation", Pos: pos}, Name: name}
	r.Source = p.parseReference()
	p.expect(tArrow)
	r.Target = p.parseReference()
	p.expect(tLBrace)
	for p.cur().Kind == tIdent {
		r.Fields = append(r.Fields, p.parseField())
	}
	p.expect(tRBrace)
	return r
}

// Brain: 'brain' ID '{' Ownership* '}'  with Ownership: 'owns' RuntimeArea
func (p *parser) parseBrain() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("brain")
	name, _ := p.expectID()
	b := &Brain{declBase: declBase{Kind: "Brain", Pos: pos}, Name: name}
	p.expect(tLBrace)
	for p.isKeyword("owns") {
		p.next()
		area := p.cur()
		if area.Kind != tIdent || !runtimeAreas[area.Value] {
			p.failf(area.Pos, "expected a runtime area but found %s", area.describe())
		}
		p.next()
		b.Owns = append(b.Owns, area.Value)
	}
	p.expect(tRBrace)
	return b
}

// Mind: 'mind' ID 'for' [Entity] '{' Field* '}'
func (p *parser) parseMind() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("mind")
	name, _ := p.expectID()
	p.expectKeyword("for")
	m := &Mind{declBase: declBase{Kind: "Mind", Pos: pos}, Name: name, Subject: p.parseReference()}
	p.expect(tLBrace)
	for p.cur().Kind == tIdent {
		m.Fields = append(m.Fields, p.parseField())
	}
	p.expect(tRBrace)
	return m
}

// Belief: 'belief' ID '{' subject claim (confidence)? (source)? '}'
func (p *parser) parseBelief() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("belief")
	name, _ := p.expectID()
	b := &Belief{declBase: declBase{Kind: "Belief", Pos: pos}, Name: name}
	p.expect(tLBrace)
	p.expectKeyword("subject")
	b.Subject = p.parseReference()
	p.expectKeyword("claim")
	b.Claim = p.expect(tString).Value
	if p.acceptKeyword("confidence") {
		num := p.expect(tNumber)
		v, err := strconv.ParseFloat(num.Value, 64)
		if err != nil {
			p.failf(num.Pos, "invalid confidence value %q", num.Value)
		}
		b.Confidence = &v
	}
	if p.acceptKeyword("source") {
		b.Source = p.expect(tString).Value
	}
	p.expect(tRBrace)
	return b
}

func (p *parser) parseConstitution() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("constitution")
	name, _ := p.expectID()
	c := &Constitution{declBase: declBase{Kind: "Constitution", Pos: pos}, Name: name}
	p.expect(tLBrace)
	for p.isKeyword("rule") {
		c.Rules = append(c.Rules, p.parseRule())
	}
	p.expect(tRBrace)
	return c
}

func (p *parser) parsePolicy() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("policy")
	name, _ := p.expectID()
	pol := &Policy{declBase: declBase{Kind: "Policy", Pos: pos}, Name: name}
	p.expect(tLBrace)
	for p.isKeyword("rule") {
		pol.Rules = append(pol.Rules, p.parseRule())
	}
	p.expect(tRBrace)
	return pol
}

// Rule: 'rule' ID ':' STRING
func (p *parser) parseRule() Rule {
	pos := p.cur().Pos
	p.expectKeyword("rule")
	name, _ := p.expectID()
	p.expect(tColon)
	stmt := p.expect(tString).Value
	return Rule{Name: name, Statement: stmt, Pos: pos}
}

// Api: 'api' HttpMethod STRING 'returns' SurrealTarget
func (p *parser) parseApi() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("api")
	method := p.cur()
	if method.Kind != tIdent || !httpMethods[method.Value] {
		p.failf(method.Pos, "expected an HTTP method but found %s", method.describe())
	}
	p.next()
	path := p.expect(tString).Value
	p.expectKeyword("returns")
	target := p.parseSurrealTarget()
	return &Api{declBase: declBase{Kind: "Api", Pos: pos}, Method: method.Value, Path: path, Target: target}
}

// Function: 'function' QualifiedName '(' (Param (',' Param)*)? ')' 'returns' TypeRef 'as' SurrealTarget
func (p *parser) parseFunction() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("function")
	name, _ := p.parseQualifiedName()
	fn := &Function{declBase: declBase{Kind: "Function", Pos: pos}, Name: name}
	p.expect(tLParen)
	if p.cur().Kind != tRParen {
		fn.Params = append(fn.Params, p.parseParam())
		for p.cur().Kind == tComma {
			p.next()
			fn.Params = append(fn.Params, p.parseParam())
		}
	}
	p.expect(tRParen)
	p.expectKeyword("returns")
	fn.Return = p.parseTypeRef()
	p.expectKeyword("as")
	fn.Target = p.parseSurrealTarget()
	return fn
}

// Param: ID ':' TypeRef
func (p *parser) parseParam() Param {
	name, pos := p.expectID()
	p.expect(tColon)
	return Param{Name: name, Type: p.parseTypeRef(), Pos: pos}
}

// SurrealMlBinding: 'surrealml' ID 'for' InferencePurpose 'input' [Entity] 'output' [Entity]
func (p *parser) parseSurrealMlBinding() Declaration {
	pos := p.cur().Pos
	p.expectKeyword("surrealml")
	name, _ := p.expectID()
	p.expectKeyword("for")
	purpose := p.cur()
	if purpose.Kind != tIdent || !inferencePurposes[purpose.Value] {
		p.failf(purpose.Pos, "expected an inference purpose but found %s", purpose.describe())
	}
	p.next()
	binding := &SurrealMlBinding{declBase: declBase{Kind: "SurrealMlBinding", Pos: pos}, Name: name, Purpose: purpose.Value}
	p.expectKeyword("input")
	binding.Input = p.parseReference()
	p.expectKeyword("output")
	binding.Output = p.parseReference()
	return binding
}

func (p *parser) parseSurrealTarget() string {
	t := p.cur()
	if t.Kind != tIdent || !surrealTargets[t.Value] {
		p.failf(t.Pos, "expected 'surrealdb' or 'surrealml' but found %s", t.describe())
	}
	p.next()
	return t.Value
}
