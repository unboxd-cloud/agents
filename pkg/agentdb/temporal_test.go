package agentdb

import (
	"testing"
	"time"
)

func TestTemporalAsOf(t *testing.T) {
	ts := NewTemporalStore(nil)

	v1, _ := ts.PutRecord(Record{ID: "a", Kind: "agent", Data: map[string]any{"trust": 1.0}})
	time.Sleep(time.Millisecond)
	mid := time.Now().UTC()
	time.Sleep(time.Millisecond)
	v2, _ := ts.PutRecord(Record{ID: "a", Kind: "agent", Data: map[string]any{"trust": 2.0}})

	// As-of before anything existed.
	if _, ok := ts.AsOf("a", v1.CreatedAt.Add(-time.Hour)); ok {
		t.Fatalf("AsOf before creation should be false")
	}
	// As-of mid (after v1, before v2) returns v1.
	got, ok := ts.AsOf("a", mid)
	if !ok || got.Data["trust"] != 1.0 {
		t.Fatalf("AsOf mid = %v (ok=%v), want trust 1.0", got.Data["trust"], ok)
	}
	// As-of now returns v2.
	got, ok = ts.AsOf("a", v2.UpdatedAt)
	if !ok || got.Data["trust"] != 2.0 {
		t.Fatalf("AsOf now = %v, want trust 2.0", got.Data["trust"])
	}
	// Current read (delegated) reflects the latest.
	if cur, _ := ts.GetRecord("a"); cur.Data["trust"] != 2.0 {
		t.Fatalf("current = %v, want 2.0", cur.Data["trust"])
	}
}

func TestTemporalHistoryIsImmutable(t *testing.T) {
	ts := NewTemporalStore(nil)
	ts.PutRecord(Record{ID: "a", Kind: "agent", Data: map[string]any{"trust": 1.0}})
	ts.PutRecord(Record{ID: "a", Kind: "agent", Data: map[string]any{"trust": 2.0}})

	h := ts.History("a")
	if len(h) != 2 {
		t.Fatalf("history len = %d, want 2", len(h))
	}
	// Mutating a returned version's Data must not corrupt stored history.
	h[0].Record.Data["trust"] = 99.0
	if again := ts.History("a"); again[0].Record.Data["trust"] != 1.0 {
		t.Fatalf("stored history was mutated: %v", again[0].Record.Data["trust"])
	}
}

func TestTemporalReplayAndDistance(t *testing.T) {
	ts := NewTemporalStore(nil)
	var stamps []time.Time
	for i := 0; i < 4; i++ {
		r, _ := ts.PutRecord(Record{ID: "a", Kind: "agent", Data: map[string]any{"v": i}})
		stamps = append(stamps, r.UpdatedAt)
		time.Sleep(time.Millisecond)
	}
	// Replay the full span returns all 4 versions in order.
	rep := ts.Replay("a", stamps[0], stamps[3])
	if len(rep) != 4 {
		t.Fatalf("replay len = %d, want 4", len(rep))
	}
	for i, v := range rep {
		if v.Record.Data["v"] != i {
			t.Fatalf("replay[%d] = %v, want %d", i, v.Record.Data["v"], i)
		}
	}
	// Distance between first and last is the number of versions in between.
	if d := ts.Distance("a", stamps[0], stamps[3]); d != 4 {
		t.Fatalf("distance = %d, want 4", d)
	}
}

func TestTemporalChangedSince(t *testing.T) {
	ts := NewTemporalStore(nil)
	ts.PutRecord(Record{ID: "a", Kind: "agent"})
	time.Sleep(time.Millisecond)
	cut := time.Now().UTC()
	time.Sleep(time.Millisecond)
	ts.PutRecord(Record{ID: "b", Kind: "agent"})

	changed := ts.ChangedSince(cut)
	if len(changed) != 1 || changed[0] != "b" {
		t.Fatalf("ChangedSince = %v, want [b]", changed)
	}
}
