// Package authz is the composable authorization seam. It separates the two
// concerns cleanly (ADR direction): a coarse policy gate (OPA) and a
// fine-grained relationship check (OpenFGA). Services depend only on these
// interfaces, so the engines are swappable and never hard-coded.
package authz

import "context"

// Request is the subject/action/resource tuple every decision is made on.
// TenantID keeps authorization on the same join key as the rest of the platform.
type Request struct {
	TenantID string
	Subject  string // identity from Dex/SPIFFE
	Profile  string // persona profile
	Action   string // e.g. "catalog.order", "billing.read"
	Resource string // e.g. "offering:managed-inference"
}

// Gate is the coarse policy-injection layer (OPA/Gatekeeper). It decides whether
// a request is allowed by policy at all.
type Gate interface {
	Allow(ctx context.Context, r Request) (bool, error)
}

// RelationChecker is fine-grained relationship-based access (OpenFGA):
// "does Subject have Relation on Object?".
type RelationChecker interface {
	Check(ctx context.Context, tenantID, subject, relation, object string) (bool, error)
}

// Authorizer composes both: policy gate first, then relationship check. This is
// the single entry point services call.
type Authorizer struct {
	Gate     Gate
	Relation RelationChecker
}

// Authorize runs the gate, then (if a relation is named) the relationship check.
func (a Authorizer) Authorize(ctx context.Context, r Request, relation, object string) (bool, error) {
	if a.Gate != nil {
		ok, err := a.Gate.Allow(ctx, r)
		if err != nil || !ok {
			return false, err
		}
	}
	if a.Relation != nil && relation != "" {
		return a.Relation.Check(ctx, r.TenantID, r.Subject, relation, object)
	}
	return true, nil
}

// AllowAll is a no-op Gate used as the Phase 0 default (real OPA in Phase 1).
type AllowAll struct{}

// Allow always permits.
func (AllowAll) Allow(context.Context, Request) (bool, error) { return true, nil }
