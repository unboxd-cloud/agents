# CNCF Stack — Capability → Project Mapping

The platform is assembled from existing [CNCF landscape](https://landscape.cncf.io/)
projects. We deploy and configure these; we do not rebuild them. Our owned code
(the control plane) is only the glue that a complete product needs.

| Capability | CNCF project(s) | How the platform uses it |
|------------|-----------------|--------------------------|
| Container orchestration | **Kubernetes** | The substrate every other component runs on; also a `Provider`. |
| Infra provisioning / IaC control plane | **Crossplane** | Catalog requests become Crossplane claims; compositions target the active provider (incl. Apache CloudStack). |
| GitOps delivery | **Argo CD** / **Flux** | Reconciles all platform components, including this control plane, from Git. |
| Packaging | **Helm** | Charts for platform + tenant workloads (`deploy/helm`). |
| Multi-tenancy isolation | **Capsule**, **vCluster** | Enforces per-tenant namespace/cluster boundaries. |
| Cost / usage metering | **OpenCost** | Authoritative source of resource cost & usage; a `metering.Source`. |
| Metrics | **Prometheus** | Usage signals (vCPU-hours, GB, requests); a `metering.Source`. |
| Telemetry pipeline | **OpenTelemetry** | Collector pushes usage/events into metering. |
| Human identity | **Dex** | OIDC broker for portal/API auth. |
| Workload identity | **SPIFFE / SPIRE** | mTLS identity between services. |
| Service mesh | **Istio** / **Linkerd** | Traffic policy, mTLS, per-tenant network policy. |
| Autoscaling | **KEDA** | Event-driven scaling of tenant workloads. |
| Developer portal | **Backstage** | Self-service front door over the catalog API. |
| Secrets | **External Secrets Operator** | Pulls credentials for providers/tenants. |
| Policy injection layer | **OPA / Gatekeeper**, **Kyverno** | The cross-cutting decision point: admission control, API request authz, and per-tenant guardrails/quotas are injected as policy, not hard-coded. Every control-plane request can be gated by an OPA decision. |
| Fine-grained authorization | **OpenFGA** | Relationship-based access control (ReBAC) for "which member/profile may act on which tenant resource". The control plane asks OpenFGA `check(user, relation, object)`; OPA enforces the coarse gate, OpenFGA answers the fine-grained relationship. |
| Eventing | **NATS**, **CloudEvents** | Usage and lifecycle events between services. |
| Storage abstraction | **Rook**, **Container Storage Interface** | Backing storage offerings in the catalog. |
| Managed databases | **KubeDB** (by AppsCode) | Runs Postgres/MySQL/MariaDB/MongoDB/Redis/Elasticsearch as CRDs; backs the RDS-compatible `rds` offering. The control-plane glue lives in `internal/kubedb` and provisions through the `provider` seam. |

## "Services for CNCF projects"

The **service catalog** (`internal/catalog`) offers CNCF projects *as
provisionable, metered services*. Each catalog entry binds:

1. a CNCF project (what the tenant gets),
2. a Crossplane composition (how it's provisioned on the active provider), and
3. one or more **meters** (how it's billed pay-as-you-go).

Example seeded entries: managed Kubernetes (vCluster), managed Prometheus,
managed NATS, managed databases (KubeDB), object storage (Rook). Adding an
offering is data, not code — keeping the system composable.
