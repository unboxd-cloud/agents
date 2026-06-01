# Operating & Partnership Models

The same multi-tenant, pay-as-you-go engine supports several commercial models
simultaneously. Modes are a thin overlay on a rated `Invoice` (see
`internal/billing/partner.go`) — the rater is never duplicated per model.

| Mode | Who is billed | Money flow | Field |
|------|---------------|-----------|-------|
| **Direct** | End customer | List price → platform | `ModeDirect` |
| **Reseller** | Partner | Partner pays base; resells with **markup** they keep | `ModeReseller` (markup `Rate`) |
| **Service provider (MSP)** | MSP | MSP wraps under own brand; **markup** on base | `ModeServiceProvider` |
| **Agency** | End customer | Customer pays list; agency earns a **commission** | `ModeAgency` (commission `Rate`) |
| **Marketplace** | End customer (via marketplace) | Marketplace takes a **commission** | `ModeMarketplace` |

## How it composes

1. Usage is metered and rated into a base `Invoice` per tenant (unchanged).
2. A `Partner{Mode, Rate}` is attached to the tenant's account.
3. `Settle(invoice, partner)` produces a `Settlement` with:
   - `GrossToCustomer` — what the end customer pays,
   - `NetToPlatform` — what the platform receives,
   - `Adjustment` — the markup added or commission taken.

This keeps tenancy + rating as the single source of truth and expresses every
partnership model as data (`Mode`, `Rate`), not new billing code — consistent
with the "compose, don't duplicate" tenet.

## Marketplace publishing model

Third parties **publish** offerings into the catalog. Each catalog `Offering`
carries a `Publisher` and a `RevShare` (the publisher's share of rated revenue;
the platform keeps the remainder). Publishing flow:

1. A publisher submits an `Offering` (project + Crossplane composition + meters).
2. Policy gate (OPA) + relationship check (OpenFGA) authorize the listing.
3. When a tenant consumes a published offering, usage is metered and rated by the
   **same** engine, then split: `publisher = total * RevShare`, platform keeps
   the rest. This reuses `Settle` semantics — publishing adds no new billing code.

First-party offerings simply omit `Publisher` (treated as `platform`,
`RevShare = 0`).

## Tenant hierarchies (reseller trees)

Resellers and MSPs own sub-tenants. The `TenantID` join key (ADR-0002) extends
to a parent/child relationship: a partner's rollup invoice is the sum of its
sub-tenants' base invoices, then `Settle`d once at the partner level. Deep
hierarchies (distributor → reseller → customer) chain the same overlay.
