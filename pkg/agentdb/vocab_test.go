package agentdb

import "testing"

func TestVocabPutGetDefine(t *testing.T) {
	v := NewVocab(NewMemStore())
	if err := v.Put(Term{
		Word:       "agent",
		Lang:       "en",
		Definition: "an autonomous actor that proposes and executes governed changes",
		Maps:       map[string]string{"es": "agente", "de": "Agent"},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := v.Put(Term{Word: "", Lang: "en"}); err == nil {
		t.Error("Put with empty word should be invalid")
	}

	got, ok := v.Get("agent", "en")
	if !ok || got.Definition == "" || got.Maps["es"] != "agente" {
		t.Fatalf("Get = %+v ok=%v", got, ok)
	}
	if def, ok := v.Define("agent", "en"); !ok || def == "" {
		t.Errorf("Define = %q ok=%v", def, ok)
	}
	if _, ok := v.Define("missing", "en"); ok {
		t.Error("Define(missing) should be false")
	}
}

func TestVocabTranslate(t *testing.T) {
	v := NewVocab(NewMemStore())
	_ = v.Put(Term{Word: "policy", Lang: "en", Definition: "a governing rule", Maps: map[string]string{"es": "política"}})

	if w, ok := v.Translate("policy", "en", "es"); !ok || w != "política" {
		t.Errorf("Translate en->es = %q ok=%v", w, ok)
	}
	if w, ok := v.Translate("policy", "en", "en"); !ok || w != "policy" {
		t.Errorf("same-language translate should echo: %q", w)
	}
	if _, ok := v.Translate("policy", "en", "fr"); ok {
		t.Error("missing mapping should be false")
	}
	if _, ok := v.Translate("nope", "en", "es"); ok {
		t.Error("unknown word should be false")
	}
}

func TestVocabList(t *testing.T) {
	v := NewVocab(NewMemStore())
	_ = v.Put(Term{Word: "agent", Lang: "en"})
	_ = v.Put(Term{Word: "agente", Lang: "es"})
	if l := v.List("en"); len(l) != 1 || l[0].Word != "agent" {
		t.Errorf("List(en) = %+v", l)
	}
	if l := v.List(""); len(l) != 2 {
		t.Errorf("List(all) = %d, want 2", len(l))
	}
}
