package catalog

import "testing"

func TestSeededHasOfferings(t *testing.T) {
	s := Seeded()
	if len(s.List()) == 0 {
		t.Fatal("seeded catalog is empty")
	}
	if _, ok := s.Get("managed-inference"); !ok {
		t.Error("expected AI-native managed-inference offering")
	}
}

func TestForProfileFilters(t *testing.T) {
	s := Seeded()
	pm := s.ForProfile("product_manager")
	for _, o := range pm {
		found := false
		for _, p := range o.Profiles {
			if p == "product_manager" {
				found = true
			}
		}
		if !found {
			t.Errorf("offering %s leaked to product_manager", o.ID)
		}
	}
	// product_manager should not see SRE-only managed-prometheus
	if _, ok := s.Get("managed-prometheus"); !ok {
		t.Fatal("fixture missing")
	}
	for _, o := range pm {
		if o.ID == "managed-prometheus" {
			t.Error("product_manager should not see managed-prometheus")
		}
	}
}
