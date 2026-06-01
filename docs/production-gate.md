# Production Gate (everything is production / enterprise)

The local **sandbox is local** and ungated (fast dev with podman). The moment you
**publish or deploy**, it is **production and enterprise** — and must pass the
production gate. There is no ungated publish path.

Implemented in `internal/release` (`Gate.Check`), composing existing primitives:

| Gate | Enforces | Backed by |
|------|----------|-----------|
| **Dependency gate** | all deps resolve (no cycle) and none are unmet | `depgraph` |
| **Human approval gate** | the workflow is `Approved` and a **human** signed off | `workflow` |
| **Accountability gate** | every approved stage records **who** (`By`) and **when** (`At`) | `workflow` |

`Check(target, workflow)` returns `{ok, reasons}`; publish/deploy proceeds only
when `ok` is true, otherwise the reasons explain exactly what is missing.

## Flow
```
local sandbox (ungated, podman)  ──►  publish/deploy request
                                         │  release.Gate.Check
            ┌────────────────────────────┼────────────────────────────┐
            ▼                            ▼                            ▼
   dependency gate            human approval gate          accountability gate
   (depgraph resolves)        (workflow Approved +         (By + At recorded
                               human sign-off)              on every stage)
            └────────────────────────────┬────────────────────────────┘
                                          ▼
                       all green ─► publishing route / GitOps apply
                       else      ─► blocked with reasons
```

## Why
- **Enterprise by default**: no "dev shortcut" reaches production; the same gate
  applies to first-party and marketplace publishing routes.
- **LLM ≠ human sign-off**: LLM approvers can review/advance stages, but the gate
  still requires a recorded human sign-off for production accountability.
- **Auditable**: the workflow records who/what decided and when — no hidden
  logic; the gate's reasons are explicit.

This gate sits in front of `docs/publishing-routes.md` and the GitOps/orchestrator
apply, making "everything is production while publishing" concrete.
