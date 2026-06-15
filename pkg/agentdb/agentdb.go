// Package agentdb is the Go port of the AGenNext agent-db-runtime: a governed
// state machine for agents. It treats stored state as something agents propose
// changes to, policies govern, humans approve, workers apply, and the store
// reconciles — the same operational model as the SurrealDB-backed runtime, here
// behind a database-agnostic in-memory store seam (SurrealDB-backed later).
//
//	Agents propose. Policies govern. Humans approve. Workers execute. The store reconciles.
//
// Records are typed nodes (the agent-db "entity"); Edges are typed relations
// (the agent-db "relation"). The governance lifecycle lives in governance.go.
package agentdb

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// Record is a stored entity: a typed node with arbitrary attributes.
type Record struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Edge is a typed relation between two records.
type Edge struct {
	ID   string         `json:"id"`
	Kind string         `json:"kind"`
	From string         `json:"from"`
	To   string         `json:"to"`
	Data map[string]any `json:"data,omitempty"`
}

// Errors returned by a Store.
var (
	ErrNotFound = errors.New("record not found")
	ErrInvalid  = errors.New("invalid")
)

// Store is the persistence seam for records and edges (in-memory now, SurrealDB
// later — the same database-agnostic seam used across the platform).
type Store interface {
	PutRecord(r Record) (Record, error)
	GetRecord(id string) (Record, bool)
	ListRecords(kind string) []Record
	PutEdge(e Edge) (Edge, error)
	Edges(from string) []Edge
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu      sync.RWMutex
	records map[string]Record
	edges   map[string]Edge
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore {
	return &MemStore{records: map[string]Record{}, edges: map[string]Edge{}}
}

// PutRecord creates or replaces a record, preserving CreatedAt across updates.
func (s *MemStore) PutRecord(r Record) (Record, error) {
	if r.ID == "" || r.Kind == "" {
		return Record{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if existing, ok := s.records[r.ID]; ok {
		r.CreatedAt = existing.CreatedAt
	} else if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	s.records[r.ID] = r
	return r, nil
}

// GetRecord returns the record with the given id.
func (s *MemStore) GetRecord(id string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[id]
	return r, ok
}

// ListRecords returns records of the given kind ("" for all), ordered by id.
func (s *MemStore) ListRecords(kind string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Record
	for _, r := range s.records {
		if kind == "" || r.Kind == kind {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// PutEdge creates or replaces an edge.
func (s *MemStore) PutEdge(e Edge) (Edge, error) {
	if e.ID == "" || e.Kind == "" || e.From == "" || e.To == "" {
		return Edge{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.edges[e.ID] = e
	return e, nil
}

// Edges returns edges originating at from ("" for all), ordered by id.
func (s *MemStore) Edges(from string) []Edge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Edge
	for _, e := range s.edges {
		if from == "" || e.From == from {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
