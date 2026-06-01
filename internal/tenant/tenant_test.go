package tenant

import (
	"errors"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemStore()
	in := Tenant{ID: "acme", Name: "Acme", Members: []Member{{Subject: "u1", Profile: ProfileSRE}}}
	if _, err := s.Create(in); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("acme")
	if err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt not defaulted")
	}
}

func TestCreateRejectsDuplicateAndInvalid(t *testing.T) {
	s := NewMemStore()
	if _, err := s.Create(Tenant{ID: "a", Name: "A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(Tenant{ID: "a", Name: "A"}); !errors.Is(err, ErrExists) {
		t.Errorf("want ErrExists, got %v", err)
	}
	if _, err := s.Create(Tenant{ID: "", Name: "x"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("want ErrInvalid, got %v", err)
	}
	if _, err := s.Create(Tenant{ID: "b", Name: "B", Members: []Member{{Subject: "u", Profile: "ceo"}}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("want ErrInvalid for bad profile, got %v", err)
	}
}

func TestProfileValid(t *testing.T) {
	for _, p := range ValidProfiles() {
		if !p.Valid() {
			t.Errorf("%s should be valid", p)
		}
	}
	if Profile("nope").Valid() {
		t.Error("unexpected valid profile")
	}
}
