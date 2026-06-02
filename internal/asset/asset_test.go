package asset

import "testing"

func TestUpsertDefaultsAndValidation(t *testing.T) {
	s := NewMemStore()
	if _, err := s.Upsert(Asset{Kind: "service"}); err == nil {
		t.Error("missing id/name should be invalid")
	}
	a, err := s.Upsert(Asset{ID: "service:catalog", Kind: "service", Name: "catalog", Source: "http://catalog"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if a.Status != "active" || a.DiscoveredAt.IsZero() {
		t.Errorf("defaults not applied: %+v", a)
	}
}

func TestTagsPersistOnReupsert(t *testing.T) {
	s := NewMemStore()
	_, _ = s.Upsert(Asset{ID: "service:catalog", Kind: "service", Name: "catalog", Tags: []string{"prod"}})
	_, _ = s.Upsert(Asset{ID: "service:catalog", Kind: "service", Name: "catalog"}) // no tags
	g, _ := s.Get("service:catalog")
	if len(g.Tags) != 1 || g.Tags[0] != "prod" {
		t.Errorf("tags should persist across discovery: %+v", g.Tags)
	}
}

func TestListAndKinds(t *testing.T) {
	s := NewMemStore()
	_, _ = s.Upsert(Asset{ID: "service:catalog", Kind: "service", Name: "catalog"})
	_, _ = s.Upsert(Asset{ID: "provider:k8s", Kind: "provider", Name: "kubernetes"})
	if len(s.List("service")) != 1 {
		t.Error("List(service) should be 1")
	}
	if len(s.List("")) != 2 {
		t.Error("List(all) should be 2")
	}
	if k := s.Kinds(); k["service"] != 1 || k["provider"] != 1 {
		t.Errorf("Kinds = %+v", k)
	}
}

func TestDiscover(t *testing.T) {
	s := NewMemStore()
	n := Discover(s, func() []Asset {
		return []Asset{{ID: "a", Kind: "k", Name: "a"}, {ID: "b", Kind: "k", Name: "b"}}
	})
	if n != 2 || len(s.List("")) != 2 {
		t.Errorf("Discover cataloged %d, store has %d", n, len(s.List("")))
	}
}
