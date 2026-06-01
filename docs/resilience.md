# Design for Failure

Assume every dependency can fail. The platform degrades gracefully and never
relies on a single point.

## Graceful degradation
- **UIs**: the admin panel and org console catch per-service errors and render a
  "service unavailable" note instead of failing the whole page — partial data
  still renders.
- **Headless core**: every capability is an API, so a broken FE never blocks
  operations (use the CLI/SDK or the failsafe clients in `docs/clients.md`).
- **Pure billing**: rating/tax/settlement are pure functions — given the same
  inputs they always reproduce, independent of any live service.

## Self-healing & probes
- Every service exposes `/healthz`, `/readyz`, `/metrics` (Kubernetes restarts
  unhealthy pods; Prometheus alerts on `platform_up`/errors).
- Agents are **level-triggered**: the operator re-reconciles each cycle, so a
  missed event self-corrects; reconcile errors are logged, not fatal.
- Stateless services + database-agnostic `Store` seam → restart/replace freely.

## Failsafe routes
- Podman Desktop / Headlamp inspect and run the platform if our FE is down.
- The podman `pod.yaml` and Helm chart stay portable as an always-available
  fallback and benchmark.

## Clients: communicate & save work locally (offline-first)
Clients can keep working when the backend is unreachable:
- **Save locally**: CLI commands emit JSON to stdout — redirect to save work
  (`platform catalog > catalog.json`, `platform rate < req.json > invoice.json`).
  Desired-state edits live as local dataset files under `deploy/datasets/`.
- **Local cache**: clients use a local dir (e.g. `~/.unboxd/`) to cache the last
  good catalog/pricebook/frameworks so they render offline.
- **Store-and-forward**: usage/events queue locally and replay when connectivity
  returns (CloudEvents/NATS store-and-forward); the metering API is idempotent on
  immutable `UsageEvent`s.
- **Sync via Git**: saved datasets/desired state reconcile through GitOps when
  back online — Git is the durable, offline-friendly source of truth.

> The CLI JSON-redirect and Git-sync paths exist today; the SDK offline queue and
> `~/.unboxd` cache are tracked in `docs/tracker.md`.

## Failure-mode checklist
| Failure | Mitigation |
|---------|-----------|
| A service is down | UI degrades; other services unaffected; pod restarts |
| Our FE is down | CLI/SDK + Podman Desktop/Headlamp failsafe |
| Backend unreachable (client) | local save/cache + store-and-forward + Git sync |
| Dataset malformed | GitOps agent validates and refuses to apply |
| Price/tax change | versioned price book; invoices reproducible |
| Node/cluster loss | stateless + GitOps re-reconcile onto new capacity |
