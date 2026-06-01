# Requirements: compile-time / deployment-time / runtime

The platform separates concerns by *when* they are resolved. This keeps the
binaries small and the system composable: code is fixed at compile time, wiring
at deploy time, and data/behavior at runtime.

## Compile-time requirements
Resolved when the Go binaries are built. Changing these means a rebuild.

| Concern | How |
|--------|-----|
| Toolchain | Go 1.24 |
| Core seams (interfaces) | `provider.Provider`, `metering.Source`/`StreamSource`, `authz.Gate`/`RelationChecker`, `billing` types, `compliance` types |
| **Plugins / extensions** | Native-runtime path: blank-import a plugin package so its `init()` registers it (see `docs/extensions.md`). The set of compiled-in extensions is fixed at build time. |
| Dependencies | stdlib only (no third-party modules); fully static (`CGO_ENABLED=0`) |
| Artifact | one OCI image per service, `--build-arg SERVICE=<name>` |

## Deployment-time requirements
Resolved at `helm install/upgrade` (or via GitOps reconcile). No rebuild needed.

| Concern | How |
|--------|-----|
| Which providers are active | Provider registry selection / config |
| **Datasets** (catalog offerings, price book, tax tables, compliance frameworks) | Loaded from a mounted ConfigMap (`deploy/datasets/*.json`); see `datasets.files` in Helm values |
| Tenancy isolation backend | Capsule / vCluster wiring |
| Identity backend | Dex (OIDC) / SPIFFE config |
| Persistence backend | `Store` implementation chosen by config (database-agnostic) |
| Network/edge/multi-cloud placement | Provider + region config |

## Runtime requirements
Resolved per request while the services run. No redeploy needed.

| Concern | How |
|--------|-----|
| Tenant context | `X-Tenant-ID` header / token claim on every request |
| Usage ingestion | pull (`Ingest`) or push/streaming (`Drain`, NDJSON endpoint) |
| Rating, settlement, tax | computed per request from the active price book + jurisdiction |
| Authorization | OPA policy gate + OpenFGA relationship check per request |
| Compliance evaluation | `compliance.Evaluate(profile, placement)` per placement |
| Price/tax/framework changes | swap the dataset (deployment-time) — runtime picks the active version |
