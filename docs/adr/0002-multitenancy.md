# ADR-0002: One tenant ID across four planes

## Status
Accepted

## Context
Multi-tenancy touches isolation, identity, usage, and billing. If each plane
invents its own tenant key, correlation (and correct billing) becomes fragile
and duplicative.

## Decision
A single `TenantID` is the join key across all four planes:

| Plane | Binding |
|-------|---------|
| Isolation | Capsule tenant / vCluster named from `TenantID` |
| Identity | Dex group / SPIFFE trust domain carries `TenantID` |
| Usage | every `metering.UsageEvent` carries `TenantID` |
| Billing | `PriceBook` assignment + invoices keyed by `TenantID` |

Every control-plane API requires a `TenantID` (no implicit "default" tenant in
multi-tenant mode).

## Consequences
- Correlation is trivial and consistent; no per-plane translation tables.
- Cross-tenant access is a single, auditable check, not scattered logic.
- Deleting a tenant is a fan-out over a known set of bindings.
