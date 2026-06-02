// Package s3 holds the S3 settings for k3s cluster snapshots: the bucket and
// credentials the platform uses to store and restore cluster (etcd/datastore)
// snapshots for disaster recovery.
//
// Settings are configurable at deploy time via S3_* environment variables (see
// FromEnv) and editable through the admin consoles — platform-wide on the admin
// control panel and per-org on the org console — behind the same
// database-agnostic Store seam used elsewhere (in-memory now; Postgres with a
// sealed-secret/Vault-backed secret field later).
package s3

import (
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Settings configures where k3s cluster snapshots are stored in S3. Scope is the
// owner: the literal "platform" for the platform-wide default, or a tenant/org
// ID for a per-org override.
type Settings struct {
	Scope         string    `json:"scope"`
	Bucket        string    `json:"bucket"`
	Region        string    `json:"region"`
	Endpoint      string    `json:"endpoint,omitempty"` // S3-compatible endpoint override
	Prefix        string    `json:"prefix,omitempty"`   // key prefix for snapshot objects
	AccessKeyID   string    `json:"accessKeyId,omitempty"`
	SecretKey     string    `json:"secretKey,omitempty"` // sensitive: never render in clear
	Schedule      string    `json:"schedule,omitempty"`  // cron, e.g. "0 */6 * * *"
	RetentionDays int       `json:"retentionDays,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// secretMask is shown in place of a stored secret so the UI never echoes it.
const secretMask = "********"

// Errors returned by a Store.
var (
	ErrNotFound = errors.New("s3 settings not found")
	ErrInvalid  = errors.New("invalid s3 settings")
)

// Validate checks the required fields (bucket and region) and bounds.
func (s Settings) Validate() error {
	if strings.TrimSpace(s.Scope) == "" || strings.TrimSpace(s.Bucket) == "" || strings.TrimSpace(s.Region) == "" {
		return ErrInvalid
	}
	if s.RetentionDays < 0 {
		return ErrInvalid
	}
	return nil
}

// Redacted returns a copy with the secret masked, safe to render in a UI.
func (s Settings) Redacted() Settings {
	if s.SecretKey != "" {
		s.SecretKey = secretMask
	}
	return s
}

// SecretOrKeep returns the submitted secret, or the existing one when the
// submission is blank or still the mask — so editing other fields in a form that
// shows the masked secret does not wipe the stored value.
func SecretOrKeep(submitted, existing string) string {
	submitted = strings.TrimSpace(submitted)
	if submitted == "" || submitted == secretMask {
		return existing
	}
	return submitted
}

// FromEnv builds Settings for scope from the deploy-time S3_* environment
// variables, returning the settings and whether any S3 variable was set (so the
// caller can decide whether to seed the store).
func FromEnv(scope string) (Settings, bool) {
	s := Settings{
		Scope:       scope,
		Bucket:      os.Getenv("S3_BUCKET"),
		Region:      os.Getenv("S3_REGION"),
		Endpoint:    os.Getenv("S3_ENDPOINT"),
		Prefix:      os.Getenv("S3_PREFIX"),
		AccessKeyID: os.Getenv("S3_ACCESS_KEY_ID"),
		SecretKey:   os.Getenv("S3_SECRET_ACCESS_KEY"),
		Schedule:    os.Getenv("S3_SNAPSHOT_SCHEDULE"),
	}
	if v := os.Getenv("S3_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.RetentionDays = n
		}
	}
	set := s.Bucket != "" || s.Region != "" || s.Endpoint != "" || s.AccessKeyID != "" || s.Schedule != ""
	return s, set
}

// Store is the S3-settings persistence seam.
type Store interface {
	Set(s Settings) error
	Get(scope string) (Settings, bool)
	List() []Settings
}

// MemStore is an in-memory Store keyed by scope.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Settings
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Settings{}} }

// Set validates and stores settings, defaulting UpdatedAt.
func (s *MemStore) Set(in Settings) error {
	if err := in.Validate(); err != nil {
		return err
	}
	if in.UpdatedAt.IsZero() {
		in.UpdatedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[in.Scope] = in
	return nil
}

// Get returns the settings for scope.
func (s *MemStore) Get(scope string) (Settings, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[scope]
	return v, ok
}

// List returns all settings, ordered by scope.
func (s *MemStore) List() []Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Settings, 0, len(s.m))
	for _, v := range s.m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Scope < out[j].Scope })
	return out
}
