package agentdb

import "strings"

// Term is one vocabulary entry: a word in a language, its definition, and
// word-by-word mappings of the same concept into other languages. It is the
// persisted form of an ADL `term` (word `means` definition, `maps to` others).
type Term struct {
	Word       string            `json:"word"`
	Lang       string            `json:"lang"`
	Definition string            `json:"definition"`
	Maps       map[string]string `json:"maps,omitempty"` // target lang -> word
}

// Vocab is a stored language-to-language vocabulary map with definitions, backed
// by any agentdb Store (terms are records of kind "term"). It works the same
// whether the Store is in-memory, SurrealDB, or another adapter.
type Vocab struct {
	store Store
}

// NewVocab returns a Vocab over the given Store.
func NewVocab(store Store) *Vocab { return &Vocab{store: store} }

// termID is the stable record id for a (lang, word) pair.
func termID(lang, word string) string {
	return "term:" + strings.ToLower(lang) + ":" + strings.ToLower(word)
}

// Put stores or updates a term.
func (v *Vocab) Put(t Term) error {
	if strings.TrimSpace(t.Word) == "" || strings.TrimSpace(t.Lang) == "" {
		return ErrInvalid
	}
	maps := map[string]any{}
	for k, val := range t.Maps {
		maps[k] = val
	}
	_, err := v.store.PutRecord(Record{
		ID:   termID(t.Lang, t.Word),
		Kind: "term",
		Data: map[string]any{
			"word":       t.Word,
			"lang":       t.Lang,
			"definition": t.Definition,
			"maps":       maps,
		},
	})
	return err
}

// Get returns the term for a (word, lang) pair.
func (v *Vocab) Get(word, lang string) (Term, bool) {
	r, ok := v.store.GetRecord(termID(lang, word))
	if !ok {
		return Term{}, false
	}
	return termFromData(r.Data), true
}

// Define returns the definition of a word in a language.
func (v *Vocab) Define(word, lang string) (string, bool) {
	t, ok := v.Get(word, lang)
	if !ok {
		return "", false
	}
	return t.Definition, t.Definition != ""
}

// Translate maps a word from one language to another, word by word, using the
// stored mappings. It also follows the reverse direction by looking up the
// target-language entry when a direct map is absent.
func (v *Vocab) Translate(word, fromLang, toLang string) (string, bool) {
	if strings.EqualFold(fromLang, toLang) {
		return word, true
	}
	if t, ok := v.Get(word, fromLang); ok {
		if w, ok := t.Maps[toLang]; ok && w != "" {
			return w, true
		}
	}
	return "", false
}

// List returns all terms in a language ("" for every language).
func (v *Vocab) List(lang string) []Term {
	var out []Term
	for _, r := range v.store.ListRecords("term") {
		t := termFromData(r.Data)
		if lang == "" || strings.EqualFold(t.Lang, lang) {
			out = append(out, t)
		}
	}
	return out
}

func termFromData(data map[string]any) Term {
	t := Term{}
	if s, ok := data["word"].(string); ok {
		t.Word = s
	}
	if s, ok := data["lang"].(string); ok {
		t.Lang = s
	}
	if s, ok := data["definition"].(string); ok {
		t.Definition = s
	}
	if m, ok := data["maps"].(map[string]any); ok {
		t.Maps = map[string]string{}
		for k, val := range m {
			if sv, ok := val.(string); ok {
				t.Maps[k] = sv
			}
		}
	}
	return t
}
