# Project Tracker

Status legend: ✅ done · 🟡 in progress · ⬜ planned

## Core control plane
| Item | Status |
|------|--------|
| Vendor-neutral provider seam (k8s, cloudstack, edge, aws) | ✅ |
| Tenancy + persona profiles | ✅ |
| Catalog + category-wise registries | ✅ |
| Metering (pull + streaming) | ✅ |
| Pay-as-you-go rating (tiers, allowances) | ✅ |
| Partner settlement (reseller/agency/marketplace/MSP) + publishing | ✅ |
| Taxes (VAT/GST/sales, reverse-charge, compounding) | ✅ |
| Compliance (GDPR/HIPAA/PCI/SOC2/ISO/FedRAMP/DORA/NIS2) + residency | ✅ |
| Authz seam (OPA gate + OpenFGA ReBAC) | ✅ |
| Extensions/plugins (native-runtime) | ✅ |
| Agents: operator, GitOps, k8s orchestrator | ✅ |
| SDK + CLI (`compose`) | ✅ |
| Admin control panel (htmx chat + APM + BI) | ✅ |
| Observability: `/metrics`, OTLP export | ✅ |
| OCI images + Helm + datasets + sandbox (podman) | ✅ |

## AWS-compatible service modules (MVP)
| Module | AWS service | Open-source backend | Catalog | Data-plane API |
|--------|-------------|---------------------|:------:|:--------------:|
| compute | EC2 | Kubernetes + KubeVirt | ✅ | 🟡 |
| lambda | Lambda | Knative / OpenFaaS | ✅ | 🟡 |
| sts | STS | Dex + SPIFFE/SPIRE | ✅ | 🟡 |
| sns | SNS | NATS | ✅ | 🟡 |
| ses | SES | Postal / Haraka SMTP | ✅ | 🟡 |
| s3 | S3 | Rook/Ceph RGW | ✅ | 🟡 |
| bedrock | Bedrock | KServe + llama.cpp/Ollama | ✅ | 🟡 |
| agentcore | Bedrock AgentCore | Dapr Agents | ✅ | 🟡 |
| **REST/API module** (unified AWS-compatible gateway) | — | — | — | 🟡 in progress |

"Catalog" = offering registered, metered, compliance-mapped (done).
"Data-plane API" = the AWS-wire-compatible endpoint (S3 API, STS tokens, etc.)
— in progress; tracked here as the next slice.

## Governance, DX & resilience
| Item | Status |
|------|--------|
| Approval workflows (human + LLM approvers, trusted tools/skills/artifacts) | ✅ engine |
| Trusted registry (official tools/skills/artifacts) | ✅ |
| Notebooks offering (JupyterHub) | ✅ catalog |
| Per-user coding assistant (open-source CPU LLMs) | ✅ catalog; 🟡 runtime + first-login channel config |
| Built-in UIs (admin panel, org console) | ✅ |
| Failsafe clients (Podman Desktop, Headlamp) | ✅ documented |
| Design-for-failure (degradation, probes, offline) | ✅ partial; 🟡 SDK offline queue + ~/.unboxd cache |
| Airgapped/edge local development | ✅ build/test offline; 🟡 image bundle + vendored deps |

## Next (planned)
| Item | Status |
|------|--------|
| Persistence backends (Postgres) behind Store | ⬜ |
| Dex OIDC + tenant scoping from claims | ⬜ |
| OpenCost/Prometheus metering adapters | ⬜ |
| Crossplane compositions per offering | ⬜ |
| AWS Marketplace direct publishing connector | ⬜ |
| SBOM + SLSA provenance + image signing in CI | ⬜ |
| More AWS services (DynamoDB, SQS, KMS, ...) | ⬜ |
