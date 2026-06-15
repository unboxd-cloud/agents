# Clients & Control Panels

## Primary: our own UIs (built-in, recommended)
The platform ships its own lightweight (htmx) UIs — these are the **primary**
interface and are kept as the default:
- **Admin control panel** (`cmd/admin`, :8080): chat, APM flow traces, OTLP
  export, embedded BI, providers/modes/catalog/compliance overview.
- **Org admin console** (`cmd/orgconsole`, :8085): per-organization members,
  compliance posture, available catalog, spend preview.

They are server-rendered, framework-agnostic, and headless-backed (consume the
SDK). See `docs/ui.md`.

## Complementary external clients (optional)
For Kubernetes/container-level operations you can also use:

| Client | Type | Use |
|--------|------|-----|
| **Podman Desktop** | desktop | manage containers/pods/k8s locally; run `deploy/sandbox/pod.yaml`; build/push images |
| **Headlamp** (CNCF) | browser | inspect the cluster the platform runs on; lightweight, extensible |

## Two cloud control panels — and which is better
If you want a general Kubernetes control panel beyond ours:

| | **Headlamp** (recommended) | Rancher |
|---|---|---|
| Footprint | lightweight, single-pane, CNCF sandbox | heavy, multi-cluster fleet manager |
| Extensible | plugin API | extensive but opinionated |
| Fit | matches our lightweight ethos; great for single-node k3s + small fleets | better for large multi-cluster fleet ops |

**Better default: Headlamp** — it aligns with the platform's lightweight,
vendor-neutral design and the single-node k3s path (`docs/deploy-k3s.md`). Choose
Rancher only when you need large-scale multi-cluster fleet management.

Bottom line: **use our built-in admin panel + org console as the product UI**;
add Podman Desktop (desktop) and Headlamp (browser) for infrastructure-level
visibility.

## Failsafe route (design for failure)
Our UI is primary, but it must never be a single point of failure. Podman Desktop
and Headlamp are kept as a **failsafe/benchmark route**:
- If our FE has a bug or the admin/orgconsole **server is down**, operators still
  run and inspect the platform via Podman Desktop (the sandbox `pod.yaml`) or
  Headlamp (the cluster) — the services are headless/API-first, so every action
  is reachable without our FE.
- The sandbox `pod.yaml` and Helm chart are kept podman-/Kubernetes-compatible
  precisely so this fallback always works (and doubles as a benchmark).

See `docs/resilience.md` for the broader design-for-failure approach.
