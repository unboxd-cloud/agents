# Codespaces

A **codespace** is a per-user/per-organization cloud **dev workspace** where
**agents and humans** create, track, and deliver **requirements, files, and
projects** with **acceptable quality** and a **target deadline**.

It is not a new silo — it composes primitives the platform already has:

| Codespace concept | Platform primitive |
|-------------------|--------------------|
| Workspace runtime | `codespace` catalog offering (DevContainers + code-server), metered `codespace.hour` |
| Humans | tenant `Member`s + persona `Profile`s |
| Agents | `operator` agents, LLM `Judge`, per-user `coding-assistant` (open-source CPU LLMs) |
| Files | Git-backed datasets/artifacts (offline-friendly, GitOps-synced) |
| Requirements & projects | tracked work items with status + owner + **deadline** |
| Acceptable quality (gate) | approval **workflows** (human + LLM approvers, trusted tools/skills/artifacts) + CI gate |
| Delivery | publishing/development workflow → publishing routes / GitOps apply |
| Visibility | admin panel + org console + `docs/tracker.md` + OTLP traces/metrics |

## Lifecycle (status & next steps, per codespace)
```
created → in-progress → in-review (quality gate) → delivered | rejected
```
- **Status & next steps** are tracked per codespace (and per requirement/project
  inside it) the same way modules are tracked in `docs/tracker.md`: a status mark
  plus the next action and a target deadline.
- **Quality is enforced**, not assumed: nothing is "delivered" until its workflow
  is `Approved` (human + LLM approvers, using only trusted tools/skills/artifacts)
  and CI is green.
- **Deadlines** are first-class metadata on requirements/projects; overruns are
  visible in the console and via metrics.

## Agents + humans together
- Humans own requirements and final sign-off (human approver stage).
- Agents (coding assistant, LLM judge, operator) draft, review, and reconcile —
  always within the org's trusted allow-list and the same guardrails as humans.
- Both act inside the codespace; the workflow records who/what decided and when
  (auditable, no hidden logic).

## Airgapped / edge
A codespace runs on any provider including single-node **k3s** at the edge, and
supports **airgapped** development: stdlib-only offline builds, locally saved
files/datasets, image save/load, and Git-based sync when connectivity returns
(see `docs/resilience.md`, `docs/deploy-k3s.md`).

> Status: codespace offering + model defined; the workspace runtime, per-item
> deadline tracking, and requirement objects are tracked in `docs/tracker.md`.
