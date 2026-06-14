package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
)

// Record is a VM plus the lifecycle target the reconciler drives it toward — the
// unit of desired state the control plane persists.
type Record struct {
	VM     cloudstack.VirtualMachine `json:"vm"`
	Target target                    `json:"target"`
}

// Store persists the control plane's desired state. Multiple implementations
// (in-memory, file-backed, …) let one control-plane core run over different
// backends — "single core, multi store". Implementations must be safe for
// concurrent use.
type Store interface {
	// NextID allocates a unique, monotonically increasing VM id.
	NextID(ctx context.Context) (string, error)
	Put(ctx context.Context, rec Record) error
	Get(ctx context.Context, id string) (Record, bool, error)
	List(ctx context.Context) ([]Record, error)
	Delete(ctx context.Context, id string) error
}

// MemStore is an in-memory Store (the default; no persistence).
type MemStore struct {
	mu   sync.Mutex
	recs map[string]Record
	seq  int64
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{recs: map[string]Record{}} }

func (s *MemStore) NextID(_ context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("vm-%d", s.seq), nil
}

func (s *MemStore) Put(_ context.Context, rec Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recs[rec.VM.ID] = rec
	return nil
}

func (s *MemStore) Get(_ context.Context, id string) (Record, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.recs[id]
	return rec, ok, nil
}

func (s *MemStore) List(_ context.Context) ([]Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedRecords(s.recs), nil
}

func (s *MemStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.recs, id)
	return nil
}

// FileStore is a JSON file-backed Store: desired state (and the id counter)
// survive restarts, and it can be the shared source of truth between the API
// service and an operator core (the platform's GitHub→…→store→runtime model).
// Writes are atomic (temp file + rename); a mutex serializes in-process access.
type FileStore struct {
	mu   sync.Mutex
	path string
}

type fileState struct {
	Seq     int64             `json:"seq"`
	Records map[string]Record `json:"records"`
}

// NewFileStore returns a Store persisted at path, creating the parent directory
// and an empty state file as needed.
func NewFileStore(path string) (*FileStore, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("controlplane: file store dir: %w", err)
		}
	}
	s := &FileStore{path: path}
	if _, err := s.load(); err != nil { // validate readability up front
		return nil, err
	}
	return s, nil
}

func (s *FileStore) load() (fileState, error) {
	st := fileState{Records: map[string]Record{}}
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, fmt.Errorf("controlplane: read store: %w", err)
	}
	if len(b) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, fmt.Errorf("controlplane: parse store: %w", err)
	}
	if st.Records == nil {
		st.Records = map[string]Record{}
	}
	return st, nil
}

func (s *FileStore) save(st fileState) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("controlplane: encode store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("controlplane: write store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("controlplane: commit store: %w", err)
	}
	return nil
}

func (s *FileStore) NextID(_ context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load()
	if err != nil {
		return "", err
	}
	st.Seq++
	if err := s.save(st); err != nil {
		return "", err
	}
	return fmt.Sprintf("vm-%d", st.Seq), nil
}

func (s *FileStore) Put(_ context.Context, rec Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load()
	if err != nil {
		return err
	}
	st.Records[rec.VM.ID] = rec
	return s.save(st)
}

func (s *FileStore) Get(_ context.Context, id string) (Record, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load()
	if err != nil {
		return Record{}, false, err
	}
	rec, ok := st.Records[id]
	return rec, ok, nil
}

func (s *FileStore) List(_ context.Context) ([]Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load()
	if err != nil {
		return nil, err
	}
	return sortedRecords(st.Records), nil
}

func (s *FileStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load()
	if err != nil {
		return err
	}
	delete(st.Records, id)
	return s.save(st)
}

func sortedRecords(m map[string]Record) []Record {
	out := make([]Record, 0, len(m))
	for _, r := range m {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VM.ID < out[j].VM.ID })
	return out
}
