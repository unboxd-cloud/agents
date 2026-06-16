package controlplane

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
)

func TestStores(t *testing.T) {
	ctx := context.Background()
	t.Run("mem", func(t *testing.T) { testStore(t, ctx, NewMemStore()) })
	t.Run("file", func(t *testing.T) {
		fs, err := NewFileStore(filepath.Join(t.TempDir(), "state.json"))
		if err != nil {
			t.Fatal(err)
		}
		testStore(t, ctx, fs)
	})
}

// testStore exercises the Store contract against any implementation.
func testStore(t *testing.T, ctx context.Context, s Store) {
	t.Helper()
	id1, err := s.NextID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := s.NextID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatalf("NextID not unique: %s", id1)
	}

	rec := Record{VM: cloudstack.VirtualMachine{ID: id1, Account: "t1", Name: "web-1", State: cloudstack.StateStarting}, Target: targetRunning}
	if err := s.Put(ctx, rec); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.Get(ctx, id1)
	if err != nil || !ok {
		t.Fatalf("get: %v ok=%v", err, ok)
	}
	if got.VM.Name != "web-1" || got.Target != targetRunning {
		t.Fatalf("unexpected record: %+v", got)
	}

	list, err := s.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %d records, %v", len(list), err)
	}

	if err := s.Delete(ctx, id1); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.Get(ctx, id1); ok {
		t.Fatal("record not deleted")
	}
}

func TestFileStore_PersistsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.json")

	s1, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s1.NextID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Put(ctx, Record{VM: cloudstack.VirtualMachine{ID: id, Name: "keep", Account: "t1"}, Target: targetStopped}); err != nil {
		t.Fatal(err)
	}

	// Reopening the same path recovers state and the id counter.
	s2, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := s2.Get(ctx, id)
	if err != nil || !ok {
		t.Fatalf("reopened get: %v ok=%v", err, ok)
	}
	if got.VM.Name != "keep" || got.Target != targetStopped {
		t.Fatalf("persisted record wrong: %+v", got)
	}
	if id2, _ := s2.NextID(ctx); id2 == id {
		t.Fatalf("NextID reused id after reopen: %s", id2)
	}
}
