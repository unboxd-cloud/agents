# Per-user Coding Assistants

Every user gets a coding assistant **by default**, backed by **open-source,
CPU-based LLMs** (no proprietary dependency, runnable on edge/k3s). Offered via
the catalog as `coding-assistant` (Continue / Tabby / OpenHands).

## First-login channel configuration
On a user's first login, the assistant is provisioned and a **channel config** is
loaded so it works with complete accuracy for that deployment:
- **Identity/context**: tenant, persona profile, and entitlements from Dex/SPIFFE
  (the same `TenantID`/`Profile` axes used everywhere).
- **Environment**: the deployment's endpoints (catalog/billing/compliance/SDK),
  active provider/region, and available catalog — so suggestions match what the
  user can actually deploy.
- **Trusted tools/skills**: only the org's trusted refs (see
  `docs/workflows.md`) are exposed to the assistant — identical guardrails to
  human and LLM approvers.
- **Model**: defaults to the `bedrock` open-source CPU LLM endpoint; GPU optional.

## Metering
Billed pay-as-you-go on `assistant.hour` (free monthly allowance) plus
`ai.tokens.million` and `ai.cpu.hour` (see `docs/meters.md`). Per-user usage rolls
up to the tenant invoice.

## Accuracy per deployment
Because the channel config is loaded from the live deployment (endpoints,
provider, catalog, entitlements) rather than hard-coded, each assistant reflects
*that* environment exactly — no stale or cross-deployment context.

> Status: catalog offering + config model defined; the assistant runtime
> integration is tracked in `docs/tracker.md`.
