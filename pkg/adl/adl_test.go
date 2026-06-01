package adl

import (
	"os"
	"path/filepath"
	"testing"
)

func loadSample(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "sample.adl"))
	if err != nil {
		t.Fatalf("reading sample: %v", err)
	}
	return string(b)
}

func TestCompileSampleHasNoErrors(t *testing.T) {
	res := Compile(loadSample(t))
	for _, d := range res.Diagnostics {
		t.Logf("diagnostic: %s at %d:%d - %s", d.Severity, d.Pos.Line, d.Pos.Column, d.Message)
	}
	if res.HasErrors() {
		t.Fatalf("expected no errors compiling the sample, got %d diagnostics", len(res.Diagnostics))
	}
	if res.Model == nil {
		t.Fatal("expected a model")
	}
}

func TestDeclarationCountsAndKinds(t *testing.T) {
	res := Compile(loadSample(t))
	counts := map[string]int{}
	for _, d := range res.Model.Declarations {
		switch d.(type) {
		case *Namespace:
			counts["Namespace"]++
		case *Vocabulary:
			counts["Vocabulary"]++
		case *Entity:
			counts["Entity"]++
		case *Relation:
			counts["Relation"]++
		case *Brain:
			counts["Brain"]++
		case *Mind:
			counts["Mind"]++
		case *Belief:
			counts["Belief"]++
		case *Constitution:
			counts["Constitution"]++
		case *Policy:
			counts["Policy"]++
		case *Api:
			counts["Api"]++
		case *Function:
			counts["Function"]++
		case *SurrealMlBinding:
			counts["SurrealMlBinding"]++
		default:
			t.Fatalf("unexpected declaration type %T", d)
		}
	}
	want := map[string]int{
		"Namespace": 1, "Vocabulary": 1, "Entity": 3, "Relation": 1,
		"Brain": 1, "Mind": 1, "Belief": 1, "Constitution": 1,
		"Policy": 1, "Api": 2, "Function": 1, "SurrealMlBinding": 1,
	}
	for k, v := range want {
		if counts[k] != v {
			t.Errorf("declaration %s: got %d, want %d", k, counts[k], v)
		}
	}
}

func TestVocabularyAndTerms(t *testing.T) {
	res := Compile(loadSample(t))
	var voc *Vocabulary
	for _, d := range res.Model.Declarations {
		if v, ok := d.(*Vocabulary); ok {
			voc = v
		}
	}
	if voc == nil {
		t.Fatal("vocabulary not found")
	}
	if len(voc.Terms) != 3 {
		t.Fatalf("expected 3 terms, got %d", len(voc.Terms))
	}
	if voc.Terms[0].Meaning != "a prospective customer" {
		t.Errorf("term 0 meaning = %q", voc.Terms[0].Meaning)
	}
	if voc.Terms[1].Mapping != "acme.crm.Lead" {
		t.Errorf("term 1 mapping = %q", voc.Terms[1].Mapping)
	}
	if voc.Terms[2].Meaning == "" || voc.Terms[2].Mapping == "" {
		t.Errorf("term 2 should have both meaning and mapping: %+v", voc.Terms[2])
	}
}

func TestEntityFieldsAndInheritance(t *testing.T) {
	res := Compile(loadSample(t))
	ents := map[string]*Entity{}
	for _, d := range res.Model.Declarations {
		if e, ok := d.(*Entity); ok {
			ents[e.Name] = e
		}
	}
	person := ents["Person"]
	if person == nil || len(person.Fields) != 3 {
		t.Fatalf("Person should have 3 fields: %+v", person)
	}
	if !person.Fields[0].Required {
		t.Errorf("Person.name should be required")
	}
	if person.Fields[0].Type.Primitive != "string" {
		t.Errorf("Person.name type = %+v", person.Fields[0].Type)
	}
	lead := ents["Lead"]
	if lead == nil || lead.Super == nil {
		t.Fatal("Lead should extend a supertype")
	}
	if lead.Super.Resolved != "acme.crm.Person" {
		t.Errorf("Lead.super resolved = %q, want acme.crm.Person", lead.Super.Resolved)
	}
	// owner: Person is an entity reference, resolved to qualified name.
	var owner *Field
	for i := range lead.Fields {
		if lead.Fields[i].Name == "owner" {
			owner = &lead.Fields[i]
		}
	}
	if owner == nil || owner.Type.Ref == nil {
		t.Fatal("Lead.owner should be an entity reference")
	}
	if owner.Type.Ref.Resolved != "acme.crm.Person" {
		t.Errorf("Lead.owner resolved = %q", owner.Type.Ref.Resolved)
	}
}

func TestReferencesResolve(t *testing.T) {
	res := Compile(loadSample(t))
	for _, d := range res.Model.Declarations {
		switch n := d.(type) {
		case *Relation:
			if n.Source.Resolved == "" || n.Target.Resolved == "" {
				t.Errorf("relation %s endpoints unresolved: %+v", n.Name, n)
			}
		case *Mind:
			if n.Subject.Resolved == "" {
				t.Errorf("mind %s subject unresolved", n.Name)
			}
		case *Belief:
			if n.Subject.Resolved == "" {
				t.Errorf("belief %s subject unresolved", n.Name)
			}
			if n.Confidence == nil || *n.Confidence != 0.82 {
				t.Errorf("belief %s confidence = %v", n.Name, n.Confidence)
			}
		case *SurrealMlBinding:
			if n.Input.Resolved == "" || n.Output.Resolved == "" {
				t.Errorf("surrealml %s refs unresolved", n.Name)
			}
		}
	}
}

func TestBrainOwnsAndApi(t *testing.T) {
	res := Compile(loadSample(t))
	for _, d := range res.Model.Declarations {
		switch n := d.(type) {
		case *Brain:
			if len(n.Owns) != 6 {
				t.Errorf("brain owns count = %d, want 6", len(n.Owns))
			}
		case *Api:
			if n.Target != "surrealdb" {
				t.Errorf("api target = %q", n.Target)
			}
			if n.Method != "get" && n.Method != "post" {
				t.Errorf("api method = %q", n.Method)
			}
		}
	}
}

func TestUnresolvedReferenceIsAnError(t *testing.T) {
	src := `entity A { other: B }`
	res := Compile(src)
	if !res.HasErrors() {
		t.Fatal("expected an unresolved-reference error")
	}
	found := false
	for _, d := range res.Diagnostics {
		if d.Severity == SeverityError && contains(d.Message, "unresolved") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an 'unresolved' error, got %+v", res.Diagnostics)
	}
}

func TestSyntaxErrorRecoversToNextDeclaration(t *testing.T) {
	// The first entity is malformed (missing closing brace contents), but the
	// parser should recover and still see the second, well-formed entity.
	src := `entity Broken { name string }
entity Good { name: string }`
	res := Compile(src)
	if !res.HasErrors() {
		t.Fatal("expected a syntax error for the malformed entity")
	}
	var good *Entity
	for _, d := range res.Model.Declarations {
		if e, ok := d.(*Entity); ok && e.Name == "Good" {
			good = e
		}
	}
	if good == nil {
		t.Errorf("parser failed to recover and parse the 'Good' entity: %+v", res.Model.Declarations)
	}
}

func TestParseExposesLexErrors(t *testing.T) {
	_, diags := Parse(`namespace a.b $`)
	if len(diags) == 0 {
		t.Fatal("expected a lexical diagnostic for the stray '$'")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestLoadAgentGroupsAndQueries(t *testing.T) {
	agent, diags := Load(loadSample(t))
	if HasErrors(diags) {
		t.Fatalf("unexpected errors: %+v", diags)
	}
	if len(agent.Entities) != 3 || len(agent.Brains) != 1 || len(agent.Apis) != 2 {
		t.Fatalf("grouping wrong: entities=%d brains=%d apis=%d",
			len(agent.Entities), len(agent.Brains), len(agent.Apis))
	}

	lead, ok := agent.Entity("Lead")
	if !ok || lead.Name != "Lead" {
		t.Fatal("Entity(\"Lead\") not found")
	}
	if _, ok := agent.Entity("acme.crm.Lead"); !ok {
		t.Error("Entity lookup by qualified name failed")
	}

	if minds := agent.MindsFor("Lead"); len(minds) != 1 || minds[0].Name != "LeadMind" {
		t.Errorf("MindsFor(Lead) = %+v", minds)
	}
	if beliefs := agent.BeliefsAbout("Lead"); len(beliefs) != 1 || beliefs[0].Name != "HotLead" {
		t.Errorf("BeliefsAbout(Lead) = %+v", beliefs)
	}
	if brain, ok := agent.Brain("CrmBrain"); !ok || len(brain.Owns) != 6 {
		t.Errorf("Brain(CrmBrain) = %+v ok=%v", brain, ok)
	}
}

func TestLoadEmptyIsNonNil(t *testing.T) {
	agent, _ := Load("")
	if agent == nil || len(agent.Entities) != 0 {
		t.Fatal("Load(\"\") should return a non-nil empty agent")
	}
}
