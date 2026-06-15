package agentdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SurrealConfig configures a SurrealDB-backed Store.
type SurrealConfig struct {
	Endpoint  string // e.g. http://localhost:8000
	Namespace string
	Database  string
	User      string
	Pass      string
	HTTP      *http.Client
}

// SurrealStore is the SurrealDB-backed Store: the production counterpart to
// MemStore, matching the agent-db-runtime's SurrealDB model. It speaks SurrealQL
// over SurrealDB's HTTP /sql endpoint using only the standard library (no driver
// dependency, honoring the repo's stdlib-only rule).
//
// Records live in a `record` table and edges in an `edge` table, each carrying
// the original id so the Store contract round-trips (native graph RELATE is a
// future refinement). Like the upstream agent-db-runtime, this has not yet been
// exercised against a live SurrealDB server: the SurrealQL generation and HTTP
// transport are unit-tested, but treat it as schema-first until validated end to
// end.
type SurrealStore struct {
	cfg SurrealConfig
}

// compile-time check that SurrealStore satisfies the Store seam.
var _ Store = (*SurrealStore)(nil)

// NewSurrealStore returns a SurrealDB-backed Store.
func NewSurrealStore(cfg SurrealConfig) *SurrealStore {
	if cfg.HTTP == nil {
		cfg.HTTP = &http.Client{Timeout: 10 * time.Second}
	}
	return &SurrealStore{cfg: cfg}
}

// PutRecord upserts a record via SurrealQL UPSERT.
func (s *SurrealStore) PutRecord(r Record) (Record, error) {
	if r.ID == "" || r.Kind == "" {
		return Record{}, ErrInvalid
	}
	now := time.Now().UTC()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	q := "UPSERT " + thing("record", r.ID) + " CONTENT " + mustJSON(recordDoc(r)) + " RETURN AFTER;"
	res, err := s.query(context.Background(), q)
	if err != nil {
		return Record{}, err
	}
	if err := firstErr(res); err != nil {
		return Record{}, err
	}
	return r, nil
}

// GetRecord selects a single record by id.
func (s *SurrealStore) GetRecord(id string) (Record, bool) {
	res, err := s.query(context.Background(), "SELECT * FROM "+thing("record", id)+";")
	if err != nil {
		return Record{}, false
	}
	recs := recordsFrom(res)
	if len(recs) == 0 {
		return Record{}, false
	}
	return recs[0], true
}

// ListRecords selects records, optionally filtered by kind.
func (s *SurrealStore) ListRecords(kind string) []Record {
	q := "SELECT * FROM record"
	if kind != "" {
		q += " WHERE kind = " + mustJSON(kind)
	}
	q += " ORDER BY recordId;"
	res, err := s.query(context.Background(), q)
	if err != nil {
		return nil
	}
	return recordsFrom(res)
}

// PutEdge upserts an edge via SurrealQL UPSERT.
func (s *SurrealStore) PutEdge(e Edge) (Edge, error) {
	if e.ID == "" || e.Kind == "" || e.From == "" || e.To == "" {
		return Edge{}, ErrInvalid
	}
	q := "UPSERT " + thing("edge", e.ID) + " CONTENT " + mustJSON(edgeDoc(e)) + " RETURN AFTER;"
	res, err := s.query(context.Background(), q)
	if err != nil {
		return Edge{}, err
	}
	if err := firstErr(res); err != nil {
		return Edge{}, err
	}
	return e, nil
}

// Edges selects edges originating at from.
func (s *SurrealStore) Edges(from string) []Edge {
	q := "SELECT * FROM edge"
	if from != "" {
		q += " WHERE \"from\" = " + mustJSON(from)
	}
	q += " ORDER BY edgeId;"
	res, err := s.query(context.Background(), q)
	if err != nil {
		return nil
	}
	return edgesFrom(res)
}

// query POSTs SurrealQL to the /sql endpoint and decodes the result envelope.
func (s *SurrealStore) query(ctx context.Context, surql string) ([]surrealResult, error) {
	url := strings.TrimRight(s.cfg.Endpoint, "/") + "/sql"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(surql))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "application/json")
	if s.cfg.Namespace != "" {
		req.Header.Set("NS", s.cfg.Namespace)
		req.Header.Set("surreal-ns", s.cfg.Namespace)
	}
	if s.cfg.Database != "" {
		req.Header.Set("DB", s.cfg.Database)
		req.Header.Set("surreal-db", s.cfg.Database)
	}
	if s.cfg.User != "" {
		req.SetBasicAuth(s.cfg.User, s.cfg.Pass)
	}
	resp, err := s.cfg.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("surrealdb: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return decodeEnvelope(body)
}
