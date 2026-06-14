# Unboxd Platform

**The open-source AWS alternative anchored on Apache CloudStack.**

A **vendor-neutral, CloudStack-anchored, CNCF-native, multi-tenant, pay-as-you-go** cloud platform. It is a lightweight, framework-agnostic Go **control plane** that composes Apache CloudStack and existing CNCF projects rather than rebuilding infrastructure — and exposes them as AWS-compatible, metered services that run on owned cloud, partner cloud, on-prem, or the edge.

> Status: working scaffold + reference architecture. Builds with Go 1.24,
> stdlib-only, all tests green. See [`docs/tracker.md`](docs/tracker.md) for live
> status and [`docs/roadmap.md`](docs/roadmap.md) for the plan.

## Infrastructure Anchor

```text
Apache CloudStack
  -> Open IaaS anchor
  -> Zones / Pods / Clusters / Hosts
  -> Domains / Accounts / Projects / Roles
  -> Kubernetes / k3s runtime layer
  -> Platform services
  -> AWS-compatible product APIs
```

CloudStack is the default infrastructure anchor for Unboxd Cloud Platform.

Kubernetes remains the workload orchestration layer above CloudStack. The Unboxd control plane provides catalog, tenant management, metering, billing, compliance, service APIs, and operational dashboards.

## Operating Model

```text
GitHub -> CI/CD -> Apache CloudStack -> Kubernetes/k3s -> Platform Services -> Metering/Billing/Compliance
```

The platform control plane treats infrastructure, services, tenants, usage, pricing, and compliance as governed data. GitHub stores desired state, CI/CD validates and publishes artifacts, CloudStack provides the open IaaS substrate, and Kubernetes/k3s runs the platform services.

## Principles

- **Apache CloudStack anchored**: zones, accounts, projects, templates, service offerings, and events provide the open IaaS base.
- **Open-source AWS alternative**, wire-compatible where possible (S3, STS, EC2-style compute, metering) — migrate by changing an endpoint, not your operating model. See [`docs/aws-interop.md`](docs/aws-interop.md).
- **Vendor-neutral**: CloudStack is the default provider anchor; Kubernetes/k3s, edge, and hyperscalers remain provider seams.
- **CNCF-native, compose don't rebuild**: capability -> project map in [`docs/cncf-stack.md`](docs/cncf-stack.md).
- **Multi-tenant + pay-as-you-go**: one `TenantID` axis; one rating engine for tiers, allowances, taxes, and partner settlement.
- **Headless / API-first**: every capability is an API; UIs are optional.
- **Data, not code**: catalog, pricing, taxes, CloudStack inventory, and compliance frameworks load as datasets at deployment time.

## AWS-compatible modules (MVP)

| Module | AWS | Open-source backend |
|--------|-----|---------------------|
| compute | EC2 | Apache CloudStack + Kubernetes/k3s |
| lambda | Lambda | Knative / OpenFaaS |
| sts | STS | Dex + SPIFFE/SPIRE |
| sns | SNS | NATS |
| ses | SES | Postal / Haraka |
| s3 | S3 | CloudStack primary/secondary storage + S3-compatible object gateway |

## Services & binaries (`cmd/`)

| Binary | Role | Port |
|--------|------|------|
| `catalog` | service catalog + category registries | 8083 |
| `metering` | usage ingestion from CloudStack, Kubernetes, and platform services | 8081 |
| `billing` | rating + tax + partner settlement | 8082 |
| `compliance` | frameworks + residency evaluation | 8084 |
| `admin` | platform admin panel (htmx chat + APM + BI) | 8080 |
| `orgconsole` | organization admin console | 8085 |
| `operator` | CloudStack + Kubernetes operations controller | — |
| `platform` | unified CLI (`compose`, catalog, rate, cloudstack, …) | — |

## Quick start

```bash
make check          # vet + tests
make build          # all binaries -> ./bin
make sandbox-up     # run the stack locally via podman (see docs/sandbox.md)
```

Drive it:

```bash
./bin/platform catalog            # list offerings via SDK
./bin/platform cloudstack map     # show CloudStack -> Unboxd mapping
./bin/platform compose up         # podman play kube the sandbox
curl localhost:8083/v1/categories # category-wise registry index
```

## Architecture at a glance

```text
Experience:  admin panel · org console · CLI · SDK · Backstage
Control plane: tenant · catalog · metering · billing · compliance · operator
IaaS anchor: Apache CloudStack zones · domains · accounts · projects · offerings
Runtime layer: Kubernetes/k3s · KubeVirt where needed · Knative/OpenFaaS
Governance: policy · audit · compliance · usage records
```

Full diagrams: [`docs/stack-diagram.md`](docs/stack-diagram.md) · data model: [`docs/data-model.md`](docs/data-model.md).

## Documentation

- Architecture: [`architecture.md`](docs/architecture.md), [`stack-diagram.md`](docs/stack-diagram.md), [`data-model.md`](docs/data-model.md), [`cloudstack-anchor.md`](docs/cloudstack-anchor.md)
- CNCF & registries: [`cncf-stack.md`](docs/cncf-stack.md), [`registries.md`](docs/registries.md)
- Billing: [`meters.md`](docs/meters.md), [`unit-economics.md`](docs/unit-economics.md), [`operating-models.md`](docs/operating-models.md)
- Compliance & standards: [`compliance.md`](docs/compliance.md), [`standards.md`](docs/standards.md)
- AWS & publishing: [`aws-interop.md`](docs/aws-interop.md), [`publishing-routes.md`](docs/publishing-routes.md), [`registry-publish.md`](docs/registry-publish.md)
- Extend & operate: [`extensions.md`](docs/extensions.md), [`gitops.md`](docs/gitops.md), [`observability.md`](docs/observability.md), [`ui.md`](docs/ui.md)
- Deploy & test: [`requirements.md`](docs/requirements.md), [`sandbox.md`](docs/sandbox.md), [`deploy-k3s.md`](docs/deploy-k3s.md), [`versioning.md`](docs/versioning.md)
- Plan: [`roadmap.md`](docs/roadmap.md), [`tracker.md`](docs/tracker.md), ADRs in [`docs/adr/`](docs/adr/)

## License

[Apache 2.0](LICENSE).
