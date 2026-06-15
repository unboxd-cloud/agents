package authz

import (
	"context"
	"testing"
)

type denyGate struct{}

func (denyGate) Allow(context.Context, Request) (bool, error) { return false, nil }

type fakeRel struct{ allow bool }

func (f fakeRel) Check(context.Context, string, string, string, string) (bool, error) {
	return f.allow, nil
}

func TestAuthorize_GateDenies(t *testing.T) {
	a := Authorizer{Gate: denyGate{}, Relation: fakeRel{allow: true}}
	ok, err := a.Authorize(context.Background(), Request{}, "owner", "offering:x")
	if err != nil || ok {
		t.Fatalf("gate deny should block: ok=%v err=%v", ok, err)
	}
}

func TestAuthorize_RelationDecides(t *testing.T) {
	a := Authorizer{Gate: AllowAll{}, Relation: fakeRel{allow: false}}
	ok, _ := a.Authorize(context.Background(), Request{}, "owner", "offering:x")
	if ok {
		t.Fatal("relation deny should block")
	}
	a.Relation = fakeRel{allow: true}
	ok, _ = a.Authorize(context.Background(), Request{}, "owner", "offering:x")
	if !ok {
		t.Fatal("relation allow should pass")
	}
}

func TestAuthorize_NoRelationPassesGate(t *testing.T) {
	a := Authorizer{Gate: AllowAll{}}
	ok, _ := a.Authorize(context.Background(), Request{}, "", "")
	if !ok {
		t.Fatal("gate-only allow should pass")
	}
}
