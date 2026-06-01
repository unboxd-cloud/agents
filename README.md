# Unboxd Platform

A **vendor-neutral, CNCF-native cloud platform** that turns any underlying
infrastructure (Kubernetes clusters, Apache CloudStack, public clouds) into a
self-service, **multi-tenant**, **pay-as-you-go** cloud.

The platform does **not** reinvent infrastructure primitives. It is a thin
**control plane** that composes existing, graduated/incubating
[CNCF projects](https://landscape.cncf.io/) as building blocks and adds the glue
that an end-to-end product needs: tenancy, a service catalog, metering, rating,
and billing.

> Status: **early scaffold + reference architecture**. The repository contains a
> working, dependency-light Go control plane (tenancy, catalog, metering
> ingestion, and a pay-as-you-go rating/billing engine with tests) plus the
> architecture docs and Kubernetes/Helm deployment manifests that describe how
> the full system fits together. See [`docs/roadmap.md`](docs/roadmap.md).

## Principles

1. **Vendor-neutral.** No hard dependency on any single cloud or hypervisor.
   Infrastructure is reached through a pluggable `Provider` abstraction;
   Apache CloudStack and Kubernetes are two providers, not the foundation.
2. **CNCF-native, compose don't rebuild.** Each platform capability maps to an
   existing CNCF project (see [`docs/cncf-stack.md`](docs/cncf-stack.md)). Our
   code orchestrates them.
3. **Multi-tenant by default.** Every resource, usage record, and invoice is
   scoped to a tenant.
4. **Pay-as-you-go.** Usage is metered continuously and rated against a
   versioned price book with support for graduated tiers and free allowances.

## What's in this repo

| Path | What it is |
|------|------------|
| [`docs/architecture.md`](docs/architecture.md) | Reference architecture & component map |
| [`docs/cncf-stack.md`](docs/cncf-stack.md) | Capability → CNCF project mapping |
| [`docs/roadmap.md`](docs/roadmap.md) | Phased implementation plan |
| [`docs/adr/`](docs/adr/) | Architecture Decision Records |
| `internal/tenant` | Tenant / account model |
| `internal/catalog` | Service catalog of CNCF offerings |
| `internal/metering` | Usage event ingestion + provider adapters |
| `internal/billing` | Pay-as-you-go rating engine + invoicing |
| `internal/provider` | Vendor-neutral infrastructure provider abstraction |
| `cmd/{metering,billing,catalog}` | Control-plane service entrypoints |
| `deploy/helm` | Kubernetes deployment (Helm chart) |

## Quick start

```bash
make build      # build all service binaries into ./bin
make test       # run the test suite
make check      # vet + test (the CI gate)

# run a control-plane service locally
./bin/catalog   # service catalog API on :8083
./bin/metering  # usage ingestion API on :8081
./bin/billing   # rating/billing API on :8082
```

Then explore:

```bash
curl localhost:8083/v1/catalog                       # list CNCF service offerings
curl localhost:8081/healthz                          # liveness
```

## License

[Apache 2.0](LICENSE).
