# Reference Architecture

## Goal

Turn heterogeneous infrastructure into a single **self-service, multi-tenant,
pay-as-you-go cloud** — without locking into any vendor and without rebuilding
capabilities that mature CNCF projects already provide.

## Design tenets

- **Composability over construction.** Every box below is either (a) an existing
  CNCF project we deploy and configure, or (b) a small, single-purpose control
  plane service we own. Owned services are thin and talk to each other and to
  CNCF projects through stable interfaces.
- **One abstraction per seam.** Infrastructure is reached through one
  `Provider` interface. Usage data arrives through one `metering.Source`
  interface. This keeps vendors and data sources swappable and keeps our code
  free of duplicated, provider-specific branches.
- **Tenant is a first-class axis.** Every API call, resource, usage record, and
  invoice carries a `TenantID`.

## Layered view

```
┌─────────────────────────────────────────────────────────────────────┐
│  Experience            Backstage portal · CLI · REST/gRPC APIs        │
├─────────────────────────────────────────────────────────────────────┤
│  Control plane (owned, this repo)                                     │
│    tenant │ catalog │ metering │ billing (rating + invoicing)         │
├─────────────────────────────────────────────────────────────────────┤
│  Platform services (composed CNCF projects)                          │
│    Crossplane · Argo CD · Capsule/vCluster · Dex/SPIFFE ·            │
│    OpenCost · Prometheus · OpenTelemetry · Istio/Linkerd · Keda      │
├─────────────────────────────────────────────────────────────────────┤
│  Providers (vendor-neutral seam)                                      │
│    Kubernetes │ Apache CloudStack │ AWS/GCP/Azure │ bare metal        │
└─────────────────────────────────────────────────────────────────────┘
```

## Owned control-plane services

| Service | Responsibility | Key interfaces |
|---------|----------------|----------------|
| **tenant** | Tenants, accounts, isolation metadata | `tenant.Store` |
| **catalog** | Catalog of provisionable offerings (each backed by a CNCF project + a Crossplane composition) | `catalog.Store` |
| **metering** | Ingest usage from many sources, normalize to `UsageEvent` | `metering.Source`, `metering.Store` |
| **billing** | Rate usage against a versioned `PriceBook`, produce invoices | `billing.Rater`, `billing.PriceBook` |

These are deliberately small. They share a single HTTP server package
(`internal/server`) so health checks, JSON encoding, and routing are written
once, not three times.

## Composition seams (where CNCF projects plug in)

- **Provisioning:** the catalog renders a request into a **Crossplane** claim;
  Crossplane compositions target the active **Provider**. Apache CloudStack is a
  provider implementation, not a base layer — satisfying the vendor-neutral
  requirement.
- **Metering:** `metering.Source` adapters pull from **OpenCost** and
  **Prometheus** (and can accept push from **OpenTelemetry** collectors). The
  control plane never scrapes infrastructure directly.
- **Tenancy isolation:** **Capsule** (namespace-as-a-tenant) or **vCluster**
  (virtual clusters) enforce the boundary the `tenant` service records.
- **Identity:** **Dex** (OIDC) for human auth, **SPIFFE/SPIRE** for workload
  identity. The control plane consumes identities; it does not mint them.
- **GitOps delivery:** **Argo CD** reconciles every platform component,
  including this control plane, from Git.

## Multi-tenancy model

A tenant maps to: an isolation boundary (Capsule tenant / vCluster), an
identity group (Dex/SPIFFE), a usage stream (`TenantID` on every `UsageEvent`),
and a billing account (`PriceBook` assignment + invoices). One ID threads
through all four — see ADR-0002.

## Pay-as-you-go billing flow

```
infra usage ──> OpenCost/Prometheus ──> metering.Source ──> UsageEvent store
                                                                  │
                              PriceBook (versioned, tiered) ──> Rater
                                                                  │
                                                            Invoice (per tenant)
```

Rating supports graduated tiers and free allowances per meter, so the same
engine expresses both "first 100 vCPU-hours free, then $0.04" and flat
per-unit pricing. See ADR-0003.
