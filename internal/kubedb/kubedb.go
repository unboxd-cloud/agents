// Package kubedb is the integration layer for managed databases backed by
// KubeDB by AppsCode (https://kubedb.com), the CNCF-landscape Kubernetes
// operator for running production databases as custom resources.
//
// It is the control-plane glue, not a reimplementation: a tenant's
// AWS-RDS-style request (engine, version, size, storage) is translated into a
// KubeDB custom resource and provisioned through the single vendor-neutral
// provider seam (internal/provider) — so KubeDB rides on whichever cluster the
// active Provider targets, and the rest of the platform never imports it
// directly.
//
// Defaults are configurable at deploy time via KUBEDB_* environment variables
// (see FromEnv); provisioned databases are tracked behind the same in-memory
// Store seam used elsewhere (Postgres/Vault-backed later).
package kubedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/unboxd-cloud/platform/internal/provider"
)

// Engine is a database engine KubeDB can manage. Values map 1:1 to a KubeDB
// custom-resource Kind under the kubedb.com API group.
type Engine string

// Supported engines. KubeDB manages each as its own CRD; the platform exposes
// them as a single RDS-compatible offering and selects the engine per request.
const (
	Postgres      Engine = "postgres"
	MySQL         Engine = "mysql"
	MariaDB       Engine = "mariadb"
	MongoDB       Engine = "mongodb"
	Redis         Engine = "redis"
	Elasticsearch Engine = "elasticsearch"
)

// APIVersion is the KubeDB API group/version the manifests target.
const APIVersion = "kubedb.com/v1"

// engineMeta carries the per-engine facts the integration needs: the CRD Kind
// KubeDB watches, the default version provisioned when none is requested, and
// the wire port used to build connection strings.
type engineMeta struct {
	kind           string
	defaultVersion string
	port           int
	scheme         string
}

var engines = map[Engine]engineMeta{
	Postgres:      {kind: "Postgres", defaultVersion: "16.1", port: 5432, scheme: "postgresql"},
	MySQL:         {kind: "MySQL", defaultVersion: "8.4.2", port: 3306, scheme: "mysql"},
	MariaDB:       {kind: "MariaDB", defaultVersion: "11.4.2", port: 3306, scheme: "mysql"},
	MongoDB:       {kind: "MongoDB", defaultVersion: "7.0.5", port: 27017, scheme: "mongodb"},
	Redis:         {kind: "Redis", defaultVersion: "7.4.0", port: 6379, scheme: "redis"},
	Elasticsearch: {kind: "Elasticsearch", defaultVersion: "8.15.0", port: 9200, scheme: "https"},
}

// Engines returns the supported engines, sorted, for catalog/UI listing.
func Engines() []Engine {
	out := make([]Engine, 0, len(engines))
	for e := range engines {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Errors returned by the package.
var (
	ErrInvalid    = errors.New("invalid database spec")
	ErrUnknownEng = errors.New("unknown database engine")
	ErrNotFound   = errors.New("database not found")
	ErrNoProvider = errors.New("kubedb: provider required")
)

// DeletionPolicy controls what KubeDB does with data when the resource is
// deleted. Mirrors KubeDB's spec.deletionPolicy.
type DeletionPolicy string

const (
	DeletionDelete    DeletionPolicy = "Delete"  // keep the PVC's snapshot, drop the workload
	DeletionWipeOut   DeletionPolicy = "WipeOut" // delete everything, including data
	DeletionHalt      DeletionPolicy = "Halt"    // keep PVCs and secrets
	DeletionDoNotTerm DeletionPolicy = "DoNotTerminate"
)

// Spec is a vendor-neutral managed-database request. It is intentionally small
// and RDS-shaped; KubeDB-specific detail is derived in Manifest.
type Spec struct {
	TenantID       string         `json:"tenantId"`
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace,omitempty"`
	Engine         Engine         `json:"engine"`
	Version        string         `json:"version,omitempty"`        // defaults per engine
	Replicas       int            `json:"replicas,omitempty"`       // defaults to 1
	StorageGB      int            `json:"storageGb,omitempty"`      // defaults to 10
	StorageClass   string         `json:"storageClass,omitempty"`   // cluster default if empty
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"` // defaults to Delete
}

// withDefaults returns a copy of the spec with empty fields filled in.
func (s Spec) withDefaults() Spec {
	if s.Replicas <= 0 {
		s.Replicas = 1
	}
	if s.StorageGB <= 0 {
		s.StorageGB = 10
	}
	if s.DeletionPolicy == "" {
		s.DeletionPolicy = DeletionDelete
	}
	if meta, ok := engines[s.Engine]; ok && strings.TrimSpace(s.Version) == "" {
		s.Version = meta.defaultVersion
	}
	if strings.TrimSpace(s.Namespace) == "" {
		s.Namespace = "default"
	}
	return s
}

// Validate checks the required fields and bounds.
func (s Spec) Validate() error {
	if strings.TrimSpace(s.TenantID) == "" || strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w: tenantId and name required", ErrInvalid)
	}
	if _, ok := engines[s.Engine]; !ok {
		return fmt.Errorf("%w: %q", ErrUnknownEng, s.Engine)
	}
	if s.Replicas < 0 || s.StorageGB < 0 {
		return fmt.Errorf("%w: replicas and storageGb must be non-negative", ErrInvalid)
	}
	return nil
}

// Manifest renders the KubeDB custom resource for the spec as a structured
// object (kubectl/Crossplane accept JSON, a subset of YAML, so no YAML
// dependency is needed). It applies defaults first.
func (s Spec) Manifest() (map[string]any, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	s = s.withDefaults()
	meta := engines[s.Engine]

	storage := map[string]any{
		"accessModes": []string{"ReadWriteOnce"},
		"resources": map[string]any{
			"requests": map[string]any{
				"storage": fmt.Sprintf("%dGi", s.StorageGB),
			},
		},
	}
	if s.StorageClass != "" {
		storage["storageClassName"] = s.StorageClass
	}

	return map[string]any{
		"apiVersion": APIVersion,
		"kind":       meta.kind,
		"metadata": map[string]any{
			"name":      s.Name,
			"namespace": s.Namespace,
			"labels": map[string]any{
				"app.kubernetes.io/managed-by": "unboxd-platform",
				"platform.unboxd/tenant":       s.TenantID,
			},
		},
		"spec": map[string]any{
			"version":        s.Version,
			"replicas":       s.Replicas,
			"storageType":    "Durable",
			"storage":        storage,
			"deletionPolicy": string(s.DeletionPolicy),
		},
	}, nil
}

// ManifestJSON renders the KubeDB custom resource as indented JSON bytes.
func (s Spec) ManifestJSON() ([]byte, error) {
	m, err := s.Manifest()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(m, "", "  ")
}

// ToResource maps the spec onto the vendor-neutral provider.Resource the seam
// provisions, carrying the rendered manifest so any Provider implementation can
// apply it (Kubernetes today, a Crossplane claim later).
func (s Spec) ToResource() (provider.Resource, error) {
	manifest, err := s.ManifestJSON()
	if err != nil {
		return provider.Resource{}, err
	}
	s = s.withDefaults()
	return provider.Resource{
		Kind: "database",
		Name: s.Name,
		Params: map[string]string{
			"engine":    string(s.Engine),
			"version":   s.Version,
			"namespace": s.Namespace,
			"replicas":  strconv.Itoa(s.Replicas),
			"manifest":  string(manifest),
		},
	}, nil
}

// Instance is a handle to a provisioned managed database. It embeds the
// provider Instance and adds the connection facts a client needs.
type Instance struct {
	provider.Instance
	TenantID  string    `json:"tenantId"`
	Name      string    `json:"name"`
	Engine    Engine    `json:"engine"`
	Namespace string    `json:"namespace"`
	Host      string    `json:"host"` // in-cluster service DNS
	Port      int       `json:"port"`
	CreatedAt time.Time `json:"createdAt"`
}

// DSN returns a connection string for the instance, with the supplied
// credentials. The host/port point at the KubeDB-created Service; KubeDB stores
// the generated auth secret as <name>-auth in the same namespace.
func (i Instance) DSN(user, password string) string {
	scheme := "tcp"
	if meta, ok := engines[i.Engine]; ok {
		scheme = meta.scheme
	}
	cred := ""
	if user != "" {
		cred = user
		if password != "" {
			cred += ":" + password
		}
		cred += "@"
	}
	return fmt.Sprintf("%s://%s%s:%d", scheme, cred, i.Host, i.Port)
}

// AuthSecretName is the Secret KubeDB generates with the root credentials.
func (i Instance) AuthSecretName() string { return i.Name + "-auth" }

// serviceHost is the in-cluster DNS name KubeDB exposes for the primary.
func serviceHost(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}

// Client provisions and tracks KubeDB-managed databases through a Provider.
type Client struct {
	prov  provider.Provider
	store Store
}

// NewClient returns a Client that provisions through prov and tracks instances
// in store (an in-memory Store is used when store is nil).
func NewClient(prov provider.Provider, store Store) (*Client, error) {
	if prov == nil {
		return nil, ErrNoProvider
	}
	if store == nil {
		store = NewMemStore()
	}
	return &Client{prov: prov, store: store}, nil
}

// Provision validates the spec, renders the KubeDB resource, provisions it via
// the Provider, records it, and returns the Instance.
func (c *Client) Provision(ctx context.Context, s Spec) (Instance, error) {
	if err := s.Validate(); err != nil {
		return Instance{}, err
	}
	s = s.withDefaults()
	res, err := s.ToResource()
	if err != nil {
		return Instance{}, err
	}
	pi, err := c.prov.Provision(ctx, s.TenantID, res)
	if err != nil {
		return Instance{}, fmt.Errorf("kubedb: provision %s/%s: %w", s.Engine, s.Name, err)
	}
	inst := Instance{
		Instance:  pi,
		TenantID:  s.TenantID,
		Name:      s.Name,
		Engine:    s.Engine,
		Namespace: s.Namespace,
		Host:      serviceHost(s.Name, s.Namespace),
		Port:      engines[s.Engine].port,
		CreatedAt: time.Now().UTC(),
	}
	c.store.Put(inst)
	return inst, nil
}

// Deprovision removes the database via the Provider and forgets the instance.
func (c *Client) Deprovision(ctx context.Context, tenantID, name string) error {
	inst, ok := c.store.Get(tenantID, name)
	if !ok {
		return ErrNotFound
	}
	if err := c.prov.Deprovision(ctx, tenantID, inst.ID); err != nil {
		return fmt.Errorf("kubedb: deprovision %s: %w", name, err)
	}
	c.store.Delete(tenantID, name)
	return nil
}

// Get returns a tracked instance.
func (c *Client) Get(tenantID, name string) (Instance, bool) { return c.store.Get(tenantID, name) }

// List returns a tenant's tracked instances.
func (c *Client) List(tenantID string) []Instance { return c.store.List(tenantID) }

// Store is the persistence seam for provisioned databases, keyed by tenant+name.
type Store interface {
	Put(i Instance)
	Get(tenantID, name string) (Instance, bool)
	Delete(tenantID, name string)
	List(tenantID string) []Instance
}

// MemStore is an in-memory Store.
type MemStore struct {
	mu sync.RWMutex
	m  map[string]Instance
}

// NewMemStore returns an empty in-memory Store.
func NewMemStore() *MemStore { return &MemStore{m: map[string]Instance{}} }

func key(tenantID, name string) string { return tenantID + "/" + name }

// Put inserts or replaces an instance.
func (s *MemStore) Put(i Instance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key(i.TenantID, i.Name)] = i
}

// Get returns an instance by tenant and name.
func (s *MemStore) Get(tenantID, name string) (Instance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.m[key(tenantID, name)]
	return i, ok
}

// Delete forgets an instance.
func (s *MemStore) Delete(tenantID, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key(tenantID, name))
}

// List returns a tenant's instances, ordered by name.
func (s *MemStore) List(tenantID string) []Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Instance
	for _, i := range s.m {
		if i.TenantID == tenantID {
			out = append(out, i)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Name < out[b].Name })
	return out
}

// FromEnv builds a default Spec from the deploy-time KUBEDB_* environment
// variables, filling tenantID/name from the arguments. The returned spec still
// needs Validate before use.
func FromEnv(tenantID, name string) Spec {
	s := Spec{
		TenantID:     tenantID,
		Name:         name,
		Namespace:    os.Getenv("KUBEDB_NAMESPACE"),
		Engine:       Engine(strings.ToLower(os.Getenv("KUBEDB_ENGINE"))),
		Version:      os.Getenv("KUBEDB_VERSION"),
		StorageClass: os.Getenv("KUBEDB_STORAGE_CLASS"),
	}
	if s.Engine == "" {
		s.Engine = Postgres
	}
	if v := os.Getenv("KUBEDB_REPLICAS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.Replicas = n
		}
	}
	if v := os.Getenv("KUBEDB_STORAGE_GB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.StorageGB = n
		}
	}
	if v := DeletionPolicy(os.Getenv("KUBEDB_DELETION_POLICY")); v != "" {
		s.DeletionPolicy = v
	}
	return s
}
