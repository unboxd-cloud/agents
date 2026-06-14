# Unboxd Platform

**The open-source AWS alternative — completely interoperable with AWS.**

A **vendor-neutral, CNCF-native, multi-tenant, pay-as-you-go** cloud platform. It
is a lightweight, framework-agnostic Go **control plane** that *composes existing
CNCF projects* (Crossplane, Argo CD, OpenCost, Prometheus, OPA, OpenFGA, KServe,
…) rather than rebuilding infrastructure — and exposes them as AWS-compatible,
metered services that run on any cloud, on-prem, or the edge.

> Status: working scaffold + reference architecture. Builds with Go 1.24,
> stdlib-only, all tests green. See [`docs/tracker.md`](docs/tracker.md) for live
> status and [`docs/roadmap.md`](docs/roadmap.md) for the plan.

## Operating Model

```text
GitHub → CI/CD → k3s → Agent CRD → Java Reconciler Pod → SurrealDB → Fabric Runtime
```

The platform control plane treats agents as governed data. GitHub stores desired
state, CI/CD validates and publishes artifacts, k3s runs the reconciler, and
SurrealDB becomes the runtime source of truth for Fabric.

See [`docs/agent-as-data-operating-model.md`](docs/agent-as-data-operating-model.md).

## Principles
- **Open-source AWS alternative**, wire-compatible where possible (S3, STS, …) —
  migrate by changing an endpoint, not your code. See [`docs/aws-interop.md`](docs/aws-interop.md).
- **Vendor-neutral**: one `provider` seam (Kubernetes, Apache CloudStack, edge,
  AWS). CloudStack is optional, not foundational ([ADR-0004](docs/adr/0004-cloudstack-optional.md)).
- **CNCF-native, compose don't rebuild**: capability → project map in [`docs/cncf-stack.md`](docs/cncf-stack.md).
- **Multi-tenant + pay-as-you-go**: one `TenantID` axis; one rating engine for
  tiers, allowances, taxes, and partner settlement.
- **Headless / API-first**: every capability is an API; UIs are optional.
- **Data, not code**: catalog, pricing, taxes, and compliance frameworks load as
  datasets at deployment time.

## AWS-compatible modules (MVP)
| Module | AWS | Open-source backend |
|--------|-----|---------------------|
| compute | EC2 | Kubernetes + KubeVirt |
| lambda | Lambda | Knative / OpenFaaS |
| sts | STS | Dex + SPIFFE/SPIRE |
| sns | SNS | NATS |
| ses | SES | Postal / Haraka |
| s3 | S3 | Rook/Ceph RGW |
| bedrock | Bedrock | KServe + open-source **CPU** LLMs (llama.cpp/Ollama) |
| agentcore | Bedrock AgentCore | Dapr Agents |

## Services & binaries (`cmd/`)
| Binary | Role | Port |
|--------|------|------|
| `catalog` | service catalog + category registries | 8083 |
| `metering` | usage ingestion (pull + streaming) | 8081 |
| `billing` | rating + tax + partner settlement | 8082 |
| `compliance` | frameworks + residency evaluation | 8084 |
| `admin` | platform admin panel (htmx chat + APM + BI) | 8080 |
| `orgconsole` | organization admin console | 8085 |
| `operator` | GitOps + Kubernetes-orchestrator agents | — |
| `platform` | unified CLI (`compose`, catalog, rate, …) | — |

## Quick start
```bash
make check          # vet + tests
make build          # all binaries -> ./bin
make sandbox-up     # run the stack locally via podman (see docs/sandbox.md)
```
Drive it:
```bash
./bin/platform catalog            # list offerings (via SDK)
./bin/platform compose up         # podman play kube the sandbox
curl localhost:8083/v1/categories # category-wise registry index
```

## Architecture at a glance
```
Experience:  admin panel · org console · CLI · SDK · Backstage
Control plane (this repo): tenant · catalog · metering · billing · compliance · operator
Platform services (CNCF): Crossplane · Argo CD · Capsule · OpenCost · Prometheus ·
                          OpenTelemetry · Dex/SPIFFE · OPA · OpenFGA · KEDA
Providers (one seam): Kubernetes/k3s · CloudStack · edge · AWS/GCP/Azure
```
Full diagrams: [`docs/stack-diagram.md`](docs/stack-diagram.md) ·
data model: [`docs/data-model.md`](docs/data-model.md).

## Documentation
- Architecture: [`architecture.md`](docs/architecture.md), [`stack-diagram.md`](docs/stack-diagram.md), [`data-model.md`](docs/data-model.md), [`agent-as-data-operating-model.md`](docs/agent-as-data-operating-model.md)
- CNCF & registries: [`cncf-stack.md`](docs/cncf-stack.md), [`registries.md`](docs/registries.md)
- Billing: [`meters.md`](docs/meters.md), [`unit-economics.md`](docs/unit-economics.md), [`operating-models.md`](docs/operating-models.md)
- Compliance & standards: [`compliance.md`](docs/compliance.md), [`standards.md`](docs/standards.md)
- AWS & publishing: [`aws-interop.md`](docs/aws-interop.md), [`publishing-routes.md`](docs/publishing-routes.md), [`registry-publish.md`](docs/registry-publish.md)
- Extend & operate: [`extensions.md`](docs/extensions.md), [`gitops.md`](docs/gitops.md), [`observability.md`](docs/observability.md), [`ui.md`](docs/ui.md)
- Deploy & test: [`requirements.md`](docs/requirements.md), [`sandbox.md`](docs/sandbox.md), [`deploy-k3s.md`](docs/deploy-k3s.md), [`versioning.md`](docs/versioning.md)
- Plan: [`roadmap.md`](docs/roadmap.md), [`tracker.md`](docs/tracker.md), ADRs in [`docs/adr/`](docs/adr/)

## License
[Apache 2.0](LICENSE).
