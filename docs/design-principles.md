# Design Principles

The principles every change is held to. They are intentionally few, concrete,
and testable.

1. **Open-source AWS alternative, interoperable.** Prefer AWS-wire compatibility
   so migration is endpoint-level, not a rewrite. Never lock into a vendor.
2. **Compose, don't rebuild.** Each capability maps to an existing CNCF/landscape
   project; we own thin control-plane glue only.
3. **Vendor-neutral via one seam.** Infrastructure is reached through a single
   `provider.Provider`. CloudStack is optional, not foundational.
4. **One tenant axis.** A single `TenantID` joins tenancy, identity, usage,
   billing, and compliance.
5. **Data, not code.** Catalog, pricing, taxes, and compliance frameworks load as
   datasets at deployment; new pricing/regulations are data edits.
6. **Single responsibility, lightweight, framework-agnostic.** Small stdlib-first
   services; one job each; no heavy frameworks.
7. **Headless / API-first.** Every capability is an API; UIs are optional and
   separable (our htmx UI is primary, external clients are failsafe).
8. **Pure where it counts.** Rating/tax/settlement are pure functions →
   reproducible invoices.
9. **Database-agnostic.** Persistence lives behind `Store` interfaces.
10. **Extensible by registration.** Adapters/plugins/protocols/publish-routes
    self-register (native-runtime), no core edits.
11. **Automate via agents.** Level-triggered reconcile loops (GitOps,
    orchestrator, dependency-resolver) under one operator.
12. **Everything is production/enterprise on publish.** The sandbox is local and
    ungated; publishing always passes the production gate (dependency + human
    approval + accountability). LLMs assist; a human signs off.
13. **Open standards.** OCI, OTLP, Prometheus, FOCUS, OIDC, SPIFFE, SemVer, SLSA.
14. **Observable & exportable.** `/metrics`, OTLP traces, OpenCost cost — insights
    live in your BI tool.
15. **Design for failure.** Graceful degradation, probes, offline/airgapped,
    local save, GitOps re-reconcile.
16. **No leaks, no hidden logic, no stale state.** `defer` closes, bounded
    goroutines, explicit gate reasons, auditable decisions.
17. **AI-native on open-source CPU LLMs.** Default inference is CPU-based
    open-source models; GPU is an optional tier.
18. **Honest status.** What can't be certified is tracked in the roadmap.

These map directly to the ADRs in `docs/adr/` and are enforced by tests, the CI
gate, and the production gate.
