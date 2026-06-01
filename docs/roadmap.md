# Roadmap

Each phase is independently shippable. Services stay **single-responsibility**
and **lightweight** (stdlib-first, no framework lock-in).

## Phase 0 — Scaffold (this PR)
- [x] Vendor-neutral `Provider` abstraction (Kubernetes + CloudStack stubs).
- [x] Single-responsibility control-plane services: `tenant`, `catalog`,
      `metering`, `billing`.
- [x] Shared `internal/server` package (no duplicated HTTP boilerplate).
- [x] Pay-as-you-go rating engine with graduated tiers + free allowances.
- [x] Architecture docs, CNCF mapping, ADRs, Helm chart, CI.

## Phase 1 — Real persistence & identity
- [ ] Swap in-memory stores for Postgres (operator from catalog).
- [ ] Dex OIDC on the APIs; tenant scoping enforced from token claims.
- [ ] CloudEvents/NATS bus between metering and billing.

## Phase 2 — Live metering
- [ ] OpenCost `metering.Source` adapter (real cost/usage pull).
- [ ] Prometheus `metering.Source` adapter (query-based meters).
- [ ] OpenTelemetry push ingestion endpoint.

## Phase 3 — Self-service provisioning
- [ ] Catalog → Crossplane claim rendering.
- [ ] Capsule/vCluster tenant isolation wiring.
- [ ] Argo CD app-of-apps for the whole platform.

## Phase 4 — Productized billing
- [x] Taxes (VAT/GST/sales tax, reverse-charge, compounding).
- [ ] Invoice finalization, credits, payment-gateway adapter.
- [ ] Budgets/alerts via Prometheus rules.
- [ ] Backstage plugin for catalog + spend.

## Phase 5 — AWS-compatible service modules
The open-source AWS alternative, completely interoperable with AWS. MVP modules
are registered in the catalog (metered + compliance-mapped); their AWS-wire
data-plane APIs are the current build focus. Live status: `docs/tracker.md`.
- [x] Catalog modules: compute, lambda, sts, sns, ses, s3, bedrock, agentcore.
- [ ] AWS-wire data-plane APIs (S3 API, STS tokens, SNS/SES, Bedrock inference).
- [ ] Unified AWS-compatible REST/API gateway module (in progress).
- [ ] AWS Marketplace direct-publishing connector.
- [ ] More services: DynamoDB, SQS, KMS, etc.

See **`docs/tracker.md`** for the live project tracker.

## Not yet certified / verified (honest caveats)
Declared or scaffolded but **not independently certified** — tracked here until
verified:
- **Compliance certifications** (SOC2/ISO-27001/GDPR/HIPAA/PCI-DSS/FedRAMP/…) are
  declared per-offering as *intent/mapping*, not audited attestations. Real
  certification requires third-party audit per deployment.
- **AWS wire-compatibility** (S3/STS/SNS/SES/Bedrock/AgentCore) is designed for,
  not yet conformance-tested against AWS APIs.
- **Security scans** (govulncheck/Trivy/gitleaks) run **informationally**; results
  are not a certification and gitleaks needs an org license.
- **SBOM / SLSA provenance / image signing** are specified (`docs/versioning.md`)
  but not yet emitted by CI.
- **Providers** (kubernetes/cloudstack/edge/aws) are Phase-0 stubs behind the
  seam; not production-certified.
- **Datasets** (pricing, tax rates, framework controls) are representative samples
  to be replaced with authoritative, jurisdiction-verified data at deployment.

Anything that cannot currently be certified is tracked here by design.
