package kubedb

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/unboxd-cloud/platform/internal/provider"
)

func validSpec() Spec {
	return Spec{TenantID: "t1", Name: "orders-db", Engine: Postgres}
}

func TestValidate(t *testing.T) {
	if err := validSpec().Validate(); err != nil {
		t.Fatalf("valid spec rejected: %v", err)
	}
	if err := (Spec{Name: "x", Engine: Postgres}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Errorf("want ErrInvalid for missing tenant, got %v", err)
	}
	if err := (Spec{TenantID: "t1", Name: "x", Engine: "cassandra"}).Validate(); !errors.Is(err, ErrUnknownEng) {
		t.Errorf("want ErrUnknownEng, got %v", err)
	}
}

func TestManifestDefaults(t *testing.T) {
	m, err := validSpec().Manifest()
	if err != nil {
		t.Fatal(err)
	}
	if m["apiVersion"] != APIVersion {
		t.Errorf("apiVersion = %v", m["apiVersion"])
	}
	if m["kind"] != "Postgres" {
		t.Errorf("kind = %v, want Postgres", m["kind"])
	}
	spec := m["spec"].(map[string]any)
	if spec["version"] != "16.1" {
		t.Errorf("default version = %v, want 16.1", spec["version"])
	}
	if spec["replicas"] != 1 {
		t.Errorf("default replicas = %v, want 1", spec["replicas"])
	}
	if spec["deletionPolicy"] != "Delete" {
		t.Errorf("deletionPolicy = %v", spec["deletionPolicy"])
	}
	storage := spec["storage"].(map[string]any)
	req := storage["resources"].(map[string]any)["requests"].(map[string]any)
	if req["storage"] != "10Gi" {
		t.Errorf("default storage = %v, want 10Gi", req["storage"])
	}
	if _, set := storage["storageClassName"]; set {
		t.Error("storageClassName should be omitted when empty (cluster default)")
	}
}

func TestManifestJSONValid(t *testing.T) {
	s := validSpec()
	s.StorageClass = "fast-ssd"
	b, err := s.ManifestJSON()
	if err != nil {
		t.Fatal(err)
	}
	var back map[string]any
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}
}

func TestToResource(t *testing.T) {
	s := validSpec()
	s.Replicas = 3
	r, err := s.ToResource()
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind != "database" || r.Name != "orders-db" {
		t.Errorf("unexpected resource: %+v", r)
	}
	if r.Params["engine"] != "postgres" || r.Params["replicas"] != "3" {
		t.Errorf("params = %v", r.Params)
	}
	if r.Params["manifest"] == "" {
		t.Error("manifest param empty")
	}
}

func TestProvisionAndDeprovision(t *testing.T) {
	c, err := NewClient(provider.NewKubernetes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	inst, err := c.Provision(context.Background(), validSpec())
	if err != nil {
		t.Fatal(err)
	}
	if inst.ID == "" || inst.Provider != "kubernetes" {
		t.Errorf("unexpected instance: %+v", inst)
	}
	if inst.Port != 5432 {
		t.Errorf("port = %d, want 5432", inst.Port)
	}
	if inst.Host != "orders-db.default.svc.cluster.local" {
		t.Errorf("host = %q", inst.Host)
	}
	if got := inst.AuthSecretName(); got != "orders-db-auth" {
		t.Errorf("auth secret = %q", got)
	}

	if _, ok := c.Get("t1", "orders-db"); !ok {
		t.Error("instance not tracked after provision")
	}
	if got := c.List("t1"); len(got) != 1 {
		t.Errorf("List returned %d, want 1", len(got))
	}

	if err := c.Deprovision(context.Background(), "t1", "orders-db"); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get("t1", "orders-db"); ok {
		t.Error("instance still tracked after deprovision")
	}
	if err := c.Deprovision(context.Background(), "t1", "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestNewClientRequiresProvider(t *testing.T) {
	if _, err := NewClient(nil, nil); !errors.Is(err, ErrNoProvider) {
		t.Errorf("want ErrNoProvider, got %v", err)
	}
}

func TestDSN(t *testing.T) {
	inst := Instance{Engine: Postgres, Host: "orders-db.default.svc.cluster.local", Port: 5432}
	if got := inst.DSN("admin", "s3cret"); got != "postgresql://admin:s3cret@orders-db.default.svc.cluster.local:5432" {
		t.Errorf("DSN = %q", got)
	}
	if got := inst.DSN("", ""); got != "postgresql://orders-db.default.svc.cluster.local:5432" {
		t.Errorf("DSN without creds = %q", got)
	}
}

func TestEnginesSorted(t *testing.T) {
	got := Engines()
	if len(got) != len(engines) {
		t.Fatalf("Engines returned %d, want %d", len(got), len(engines))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("engines not sorted: %v", got)
		}
	}
}
