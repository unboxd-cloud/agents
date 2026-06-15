package agentdb

import "testing"

func TestRecordPutGetList(t *testing.T) {
	s := NewMemStore()
	if _, err := s.PutRecord(Record{ID: "agent:1", Kind: "agent", Data: map[string]any{"name": "orchestrator"}}); err != nil {
		t.Fatalf("PutRecord: %v", err)
	}
	if _, err := s.PutRecord(Record{Kind: "agent"}); err == nil {
		t.Error("PutRecord without id should be invalid")
	}
	got, ok := s.GetRecord("agent:1")
	if !ok || got.Kind != "agent" {
		t.Fatalf("GetRecord = %+v ok=%v", got, ok)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("timestamps should be set")
	}
	// Update preserves CreatedAt.
	created := got.CreatedAt
	up, _ := s.PutRecord(Record{ID: "agent:1", Kind: "agent", Data: map[string]any{"name": "dr"}})
	if !up.CreatedAt.Equal(created) {
		t.Error("update should preserve CreatedAt")
	}
	_, _ = s.PutRecord(Record{ID: "human:1", Kind: "human"})
	if l := s.ListRecords("agent"); len(l) != 1 || l[0].ID != "agent:1" {
		t.Errorf("ListRecords(agent) = %+v", l)
	}
	if l := s.ListRecords(""); len(l) != 2 {
		t.Errorf("ListRecords(all) = %d, want 2", len(l))
	}
}

func TestEdges(t *testing.T) {
	s := NewMemStore()
	if _, err := s.PutEdge(Edge{ID: "e1", Kind: "supervises", From: "agent:1", To: "agent:2"}); err != nil {
		t.Fatalf("PutEdge: %v", err)
	}
	if _, err := s.PutEdge(Edge{ID: "e2", Kind: "x", From: ""}); err == nil {
		t.Error("PutEdge without from/to should be invalid")
	}
	if e := s.Edges("agent:1"); len(e) != 1 || e[0].Kind != "supervises" {
		t.Errorf("Edges(agent:1) = %+v", e)
	}
	if e := s.Edges("other"); len(e) != 0 {
		t.Errorf("Edges(other) = %+v, want none", e)
	}
}
