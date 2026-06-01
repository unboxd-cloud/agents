# ADR-0006: Extensions, agents, UI & observability

## Status
Accepted

## Context
The platform needs to be extensible (adapters/protocols/publishing routes),
automated (control loops), operable (a UI), and observable — without heavy
frameworks or vendor lock-in.

## Decision
- **Extensions:** one `plugin` registry with the `database/sql` driver model —
  plugins self-register in `init()` and are blank-imported to compile natively
  into a binary. Seams: provider, meter source, authz, tax/compliance source,
  protocol, publish route.
- **Agents:** a minimal reconcile-loop runtime (`agent`) with an `Operator` that
  supervises many agents. GitOps and the Kubernetes orchestrator are agents.
- **UI:** htmx over stdlib `html/template` — lightweight, no build step,
  framework-agnostic. Chat interface + flow-trace (APM) panel; all read paths go
  through the SDK. Services stay headless/API-first; the UI is optional.
- **Observability:** Prometheus `/metrics` on every service; a dependency-free
  tracer that exports OTLP/JSON for Jaeger/Tempo/Prometheus; BI links embedded.

## Consequences
- New integrations are plugins; new automation is agents — core code is stable.
- One client surface (SDK) is reused by CLI, agents, and UI (no duplication).
- Everything is exportable to CNCF tooling; insights live in your BI stack.
