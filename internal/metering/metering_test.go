package metering

import (
	"context"
	"testing"
	"time"
)

func TestMemStore_RecordAndQuery(t *testing.T) {
	s := NewMemStore()
	if err := s.Record(UsageEvent{TenantID: "t1", Meter: "m", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(UsageEvent{TenantID: "t2", Meter: "m", Quantity: 5}); err != nil {
		t.Fatal(err)
	}
	got := s.Query("t1", time.Time{}, time.Time{})
	if len(got) != 1 || got[0].TenantID != "t1" {
		t.Fatalf("query isolation failed: %+v", got)
	}
}

func TestMemStore_RejectsInvalid(t *testing.T) {
	s := NewMemStore()
	if err := s.Record(UsageEvent{Meter: "m", Quantity: 1}); err == nil {
		t.Error("expected error for missing tenant")
	}
	if err := s.Record(UsageEvent{TenantID: "t1", Quantity: 1}); err == nil {
		t.Error("expected error for missing meter")
	}
	if err := s.Record(UsageEvent{TenantID: "t1", Meter: "m", Quantity: -1}); err == nil {
		t.Error("expected error for negative quantity")
	}
}

type fakeSource struct{ events []UsageEvent }

func (f fakeSource) Name() string { return "fake" }
func (f fakeSource) Collect(_ context.Context, _ time.Time) ([]UsageEvent, error) {
	return f.events, nil
}

func TestIngest(t *testing.T) {
	store := NewMemStore()
	src := fakeSource{events: []UsageEvent{
		{TenantID: "t1", Meter: "m", Quantity: 1},
		{TenantID: "t1", Meter: "m", Quantity: 2},
	}}
	n, err := Ingest(context.Background(), src, store, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("ingested: want 2, got %d", n)
	}
}
