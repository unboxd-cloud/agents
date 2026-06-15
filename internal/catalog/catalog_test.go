package catalog

import "testing"

func TestSeededHasOfferings(t *testing.T) {
	s := Seeded()
	if len(s.List()) == 0 {
		t.Fatal("seeded catalog is empty")
	}
	if _, ok := s.Get("bedrock"); !ok {
		t.Error("expected AWS-compatible bedrock offering")
	}
	if _, ok := s.Get("s3"); !ok {
		t.Error("expected AWS-compatible s3 offering")
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
		// sns is tech-only and must not appear for product_manager.
		if o.ID == "sns" {
			t.Error("product_manager should not see sns")
		}
	}
}

func TestForCategoryAndCategories(t *testing.T) {
	s := Seeded()
	ai := s.ForCategory("ai")
	if len(ai) == 0 {
		t.Fatal("expected ai-category offerings")
	}
	cats := s.Categories()
	if len(cats) == 0 {
		t.Fatal("expected non-empty category index")
	}
}
