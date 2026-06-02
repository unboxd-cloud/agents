// temporal.go — time travel on the graph. The store is append-only over time:
// every write of a record is kept as a version, so the graph is queryable as-of
// any moment (replay the past, inspect the present, project from history). Time
// is a distance on the graph — a path through versions — so the distance between
// two states is the number of versions between them.
//
// TemporalStore wraps any Store, delegating the non-temporal operations and
// layering version history on top, so it is a drop-in Store with time travel.
package agentdb

import (
	"sort"
	"sync"
	"time"
)

// Version is one append-only snapshot of a record at a transaction time.
type Version struct {
	Record Record    `json:"record"`
	At     time.Time `json:"at"`
}

// TemporalStore adds bitemporal history to any Store.
type TemporalStore struct {
	Store
	mu      sync.RWMutex
	history map[string][]Version
}

// NewTemporalStore wraps s with version history. If s is nil, a fresh MemStore
// is used.
func NewTemporalStore(s Store) *TemporalStore {
	if s == nil {
		s = NewMemStore()
	}
	return &TemporalStore{Store: s, history: map[string][]Version{}}
}

// cloneRecord copies a record (and its Data map) so a stored version is not
// mutated by later writes.
func cloneRecord(r Record) Record {
	if r.Data != nil {
		d := make(map[string]any, len(r.Data))
		for k, v := range r.Data {
			d[k] = v
		}
		r.Data = d
	}
	return r
}

// PutRecord writes through to the underlying store and appends a version.
func (t *TemporalStore) PutRecord(r Record) (Record, error) {
	out, err := t.Store.PutRecord(r)
	if err != nil {
		return out, err
	}
	t.mu.Lock()
	t.history[out.ID] = append(t.history[out.ID], Version{Record: cloneRecord(out), At: out.UpdatedAt})
	t.mu.Unlock()
	return out, nil
}

// AsOf returns the state of record id as of time at — the latest version whose
// transaction time is at or before at. Returns false if the record did not yet
// exist at that moment.
func (t *TemporalStore) AsOf(id string, at time.Time) (Record, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	vs := t.history[id]
	var found Record
	ok := false
	for _, v := range vs {
		if v.At.After(at) {
			break
		}
		found, ok = v.Record, true
	}
	return cloneRecord(found), ok
}

// History returns every version of id, oldest first.
func (t *TemporalStore) History(id string) []Version {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return cloneVersions(t.history[id])
}

// cloneVersions deep-copies versions so callers cannot mutate stored history.
func cloneVersions(vs []Version) []Version {
	out := make([]Version, len(vs))
	for i, v := range vs {
		out[i] = Version{Record: cloneRecord(v.Record), At: v.At}
	}
	return out
}

// Replay returns the versions of id within [from, to], oldest first — walking
// the path forward through time.
func (t *TemporalStore) Replay(id string, from, to time.Time) []Version {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Version
	for _, v := range t.history[id] {
		if !v.At.Before(from) && !v.At.After(to) {
			out = append(out, Version{Record: cloneRecord(v.Record), At: v.At})
		}
	}
	return out
}

// Distance returns how many versions separate the states at from and to — time
// as a distance on the graph.
func (t *TemporalStore) Distance(id string, from, to time.Time) int {
	if to.Before(from) {
		from, to = to, from
	}
	return len(t.Replay(id, from, to))
}

// ChangedSince returns the ids of records with at least one version strictly
// after t, ordered by id — the working set the operator must reconcile.
func (t *TemporalStore) ChangedSince(at time.Time) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var ids []string
	for id, vs := range t.history {
		if len(vs) > 0 && vs[len(vs)-1].At.After(at) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
