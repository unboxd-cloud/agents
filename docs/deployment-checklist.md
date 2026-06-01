# Deployment Checklist

Everything is production/enterprise on publish. Run through this before and after
every deploy. The pipeline (`/.github/workflows/deploy.yml`) automates the
starred (★) steps.

## Pre-deploy
- [ ] `make check` green (vet + tests) ★ (CI gate)
- [ ] `make e2e` green (end-to-end flow) ★ (CI gate)
- [ ] `govulncheck` clean (Go package/stdlib vulns) ★ (PR gate)
- [ ] Production gate passes for the change: dependencies resolve, **human
      sign-off** recorded, accountability captured (`docs/production-gate.md`)
- [ ] Datasets reviewed (pricing/tax/frameworks) for the target jurisdiction
- [ ] CHANGELOG `Unreleased` updated; version tag chosen (SemVer)

## At deploy (we test while deploying)
- [ ] Images built (podman) for all services ★
- [ ] **Image vuln-scan (Trivy HIGH/CRITICAL) at deploy time — blocks on
      findings** ★
- [ ] Images pushed to the registry ★
- [ ] Helm upgrade with datasets applied ★
- [ ] Rollout status healthy ★

## Post-deploy (sanity)
- [ ] `scripts/post-deploy-sanity.sh` passes: `/healthz`, `/readyz`, `/metrics`
      + functional smoke for every service ★
- [ ] Auto-rollback verified on sanity failure ★
- [ ] Dashboards/alerts green (Prometheus `platform_up`, error rate); traces
      flowing (OTLP)
- [ ] Spot-check a real flow (catalog → rate → invoice; compliance evaluate)

## Release
- [ ] Tag pushed (`vX.Y.Z`), changelog section dated
- [ ] Advisory/notes published if security-relevant

> Why image scan is at deploy (not PR): it kept failing in an unobservable runner
> step; PRs stay gated by `govulncheck` + `build-test` + `e2e`, and the image
> scan runs at deploy time (blocking) and weekly. See `CHANGELOG.md`.
