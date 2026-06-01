# Approval Workflows (publishing & development)

Organizations define **workflows** as ordered approval **stages**. Each stage is
approved by a **human** or an **LLM** approver, and may only use **trusted**
tools/skills/artifacts. Nothing publishes or deploys until its workflow is
approved. Implemented in `internal/workflow`.

## Model
- **Workflow**: `Kind` (publishing | development), a subject, ordered `Stages`,
  scoped to a `TenantID`.
- **Stage**: a `Name`, an `Approver` (`human`/`llm` + id), and an `Allowed` list
  of refs usable in that stage.
- **Ref**: `{kind: tool|skill|artifact, name}`. Every ref in every stage must be
  in the org's **TrustedRegistry** — enforced at creation (`ErrUntrusted`).

## Approvers
- **Human**: `Decide(approverID, approve, usedRefs, note)` — must match the
  stage's approver id; `usedRefs` must be within the stage's allow-list.
- **LLM**: `AutoDecideLLM(judge)` runs a pluggable `Judge` (defaults to an
  **open-source CPU LLM**). The engine is model-agnostic; the judge returns
  approve/reject + a note.

Both approver types are constrained to the **same trusted allow-list** — humans
and LLMs operate under identical guardrails.

## Semantics (no hidden logic)
- Stages run in order; the first `pending` stage is current.
- A rejection short-circuits: the workflow becomes `Rejected`, further
  `Decide`/`AutoDecideLLM` return `ErrDone`.
- All stages approved → workflow `Approved`.
- Wrong approver → `ErrWrongApprover`; using a non-allowed ref → `ErrToolNotAllowed`.
- Decisions record `By`, `At`, `Note` for audit.

## How it composes
- **Publishing**: a `publishing` workflow gates a publishing route
  (`docs/publishing-routes.md`) before a listing goes to AWS/GCP/Azure/etc.
- **Development**: a `development` workflow gates the GitOps/orchestrator apply.
- Enforcement integrates with the `authz` OPA/OpenFGA seam; trusted refs are the
  org's official tools/skills/artifacts (extensions live in `plugin`).
