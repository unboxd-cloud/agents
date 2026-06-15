# Packages, Dependencies, Lineage & Versioning

## Packaging
- **Go module:** `github.com/unboxd-cloud/platform` (one module, stdlib-only).
- **Service artifacts:** one **OCI image** per command (`make images`).
- **Chart:** Helm chart `unboxd-platform` (`deploy/helm/platform`).
- **Datasets:** versioned JSON artifacts (`deploy/datasets/*.json`); the price
  book carries its own `version` field.

## Dependencies
- **Runtime deps:** none beyond the Go stdlib (no third-party modules) → minimal
  attack surface and trivial dependency lineage.
- **Composed deps (deploy-time):** CNCF projects (Crossplane, Argo CD, OpenCost,
  Prometheus, OPA, OpenFGA, …) — versioned via Helm dependencies / image tags,
  reconciled by GitOps.
- **FE:** htmx pinned by CDN version (`htmx.org@1.9.12`).

## Versioning (SemVer 2.0)
| Thing | Scheme | Notes |
|-------|--------|-------|
| Go module / API | `vMAJOR.MINOR.PATCH` | breaking API → MAJOR |
| OCI images | git SHA + SemVer tag | immutable by digest |
| Helm chart | `version` (chart) + `appVersion` | independent of app |
| Price book | dated `version` (e.g. `2026-06-01`) | invoices pin the version used → reproducible |
| Datasets | Git revision | reconciled via GitOps |

## Lineage / provenance (supply chain)
The gold standard is verifiable lineage from source → artifact → deployment:
- **SBOM** per image (SPDX or CycloneDX) generated in CI.
- **SLSA provenance** attestation for each image (who/what/how built).
- **Signing:** images signed (e.g. cosign); digests referenced in the chart.
- **Reproducibility:** static `CGO_ENABLED=0` + `-trimpath` builds; invoices pin
  price-book version and usage events are immutable, so any invoice can be
  re-derived.
- **Traceability:** OTLP traces link a request to the spans/services that served
  it (see `docs/observability.md`).

CI (`.github/workflows/ci.yml`) is the place these attestations attach; the
matrix already builds each image, ready for SBOM/sign/attest steps.
