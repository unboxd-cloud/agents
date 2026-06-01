// Package metering ingests usage from many sources and normalizes it to
// immutable UsageEvents (ADR-0003). The control plane never scrapes
// infrastructure directly; it reads through the Source seam (OpenCost,
// Prometheus, OpenTelemetry).
package metering

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

// UsageEvent is an immutable record of consumption for one tenant + meter.
type UsageEvent struct {
	TenantID string    `json:"tenantId"`
	Meter    string    `json:"meter"`
	Quantity float64   `json:"quantity"`
	At       time.Time `json:"at"`
}

// ErrInvalid is returned for malformed events.
var ErrInvalid = errors.New("invalid usage event")

// Source is the single seam for usage ingestion. Adapters wrap OpenCost,
// Prometheus, or an OpenTelemetry collector.
type Source interface {
	Name() string
	Collect(ctx context.Context, since time.Time) ([]UsageEvent, error)
}

// Store persists usage events, queryable per tenant and period.
type Store interface {
	Record(e UsageEvent) error
	Query(tenantID string, from, to time.Time) []UsageEvent
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu     sync.RWMutex
	events []UsageEvent
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{} }

// Record validates and appends a usage event.
func (s *MemStore) Record(e UsageEvent) error {
	if e.TenantID == "" || e.Meter == "" || e.Quantity < 0 {
		return ErrInvalid
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	return nil
}

// Query returns a tenant's events within [from, to), sorted by time.
// A zero from/to bound is treated as unbounded on that side.
func (s *MemStore) Query(tenantID string, from, to time.Time) []UsageEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []UsageEvent
	for _, e := range s.events {
		if e.TenantID != tenantID {
			continue
		}
		if !from.IsZero() && e.At.Before(from) {
			continue
		}
		if !to.IsZero() && !e.At.Before(to) {
			continue
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

// Ingest pulls events from a Source and records them, returning the count.
// This is the composition point for OpenCost/Prometheus adapters.
func Ingest(ctx context.Context, src Source, store Store, since time.Time) (int, error) {
	events, err := src.Collect(ctx, since)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range events {
		if err := store.Record(e); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
