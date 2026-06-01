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
