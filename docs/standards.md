# Industry & Gold Standards

The platform deliberately adopts open, widely-accepted standards over bespoke
formats — the "gold standard" is *interoperable by default, vendor-neutral,
exportable*.

## Standards adopted

| Domain | Standard | Where |
|--------|----------|-------|
| Containers/artifacts | **OCI** image & distribution spec | `Dockerfile`, all images |
| Orchestration | **Kubernetes API** | Helm chart, sandbox Pod, providers |
| Packaging | **Helm** v3 | `deploy/helm` |
| App/provisioning model | **OAM** / Crossplane | catalog compositions |
| Telemetry | **OpenTelemetry / OTLP** | `internal/observe` export |
| Metrics | **Prometheus exposition** (0.0.4) | `/metrics` |
| Cost/FinOps | **FOCUS** (FinOps Open Cost & Usage Spec) | billing/metering model |
| Eventing | **CloudEvents** | usage/lifecycle events (roadmap) |
| Human identity | **OIDC** (via Dex) | auth (roadmap) |
| Workload identity | **SPIFFE/SPIRE** | mTLS (roadmap) |
| Policy | **OPA/Rego** | authz gate |
| Authorization | **Zanzibar-style ReBAC** (OpenFGA) | fine-grained authz |
| API contracts | **OpenAPI** / JSON | services + SDK |
| Versioning | **SemVer 2.0** | modules, chart, datasets |
| Supply chain | **SLSA**, **SBOM** (SPDX/CycloneDX) | CI (see `docs/versioning.md`) |
| Licensing | **Apache-2.0** (SPDX) | `LICENSE` |
| Data residency/privacy | **GDPR/CCPA**, **ISO-27001**, **SOC2**, **HIPAA**, **PCI-DSS**, **FedRAMP** | compliance datasets |

## Gold-standard practices
- **12-factor**: config via env/datasets, stateless services, logs to stdout.
- **Headless/API-first**: every capability is an API; UI is optional.
- **Database-agnostic**: persistence behind `Store` interfaces.
- **Vendor-neutral**: one provider seam; CloudStack optional (ADR-0004).
- **Compose, don't rebuild**: capabilities map to CNCF projects.
- **Reproducible & exportable**: pure rating, OTLP traces, open formats.
- **Secure defaults**: non-root, read-only rootfs, dropped caps, static binaries.
- **GitOps**: declarative desired state, reconcile not push.
