# Changelog

All notable changes are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/); versions follow
[SemVer](https://semver.org/). Dates are UTC.

## [Unreleased]

### Added
- **Open-source AWS alternative** scaffold: vendor-neutral, CNCF-native,
  multi-tenant, pay-as-you-go control plane (Go, stdlib-only).
- AWS-compatible catalog modules: compute, lambda, sts, sns, ses, s3, bedrock,
  agentcore (open-source backends; AI on open-source CPU LLMs).
- Developer-experience offerings: notebooks, per-user coding-assistant, codespace.
- Billing: tiered pay-as-you-go rating, free allowances, jurisdiction-aware taxes
  (VAT/GST/sales, reverse-charge, compounding), partner settlement
  (direct/reseller/agency/marketplace/MSP) + marketplace & multi-cloud publishing
  routes.
- Compliance: loadable framework registry (GDPR/HIPAA/PCI-DSS/SOC2/ISO-27001/
  FedRAMP/DORA/NIS2) + data residency; OPA/OpenFGA enforcement seam.
- Governance: approval workflows (human + LLM approvers, trusted
  tools/skills/artifacts), native dependency tracker/resolver agent, and a
  production gate (dependency + human-approval + accountability) — everything is
  production/enterprise on publish; the sandbox is local.
- Platform: extension/plugin system (native-runtime), operator agent (GitOps +
  Kubernetes orchestrator), Go SDK, `platform` CLI, admin panel + org console
  (htmx, chat, APM traces), Prometheus `/metrics`, OTLP export.
- Deploy: datasets loaded at deploy (ConfigMap), DB-agnostic stores, OCI images,
  Helm chart, podman sandbox, single-node k3s + Headlamp, GitHub + GitLab CI/CD,
  security scanning (govulncheck + Trivy).
- Docs: architecture, data model, stack diagram, standards, design principles,
  unit economics, meters, operating models, requirements, versioning/lineage,
  observability, gitops, resilience, clients, roadmap + project tracker, ADRs.
- Tooling: `make e2e` (end-to-end test) and `make sanity` (post-deploy checks).

### Changed
- CI builds images with **podman**; security scans are real (Trivy scans the
  podman-exported image tar; govulncheck on the latest patched Go).

### Security
- Static, non-root, read-only-rootfs scratch images; latest patched Go stdlib.

> Not yet certified/verified items are tracked in `docs/roadmap.md` and
> `docs/tracker.md`.
