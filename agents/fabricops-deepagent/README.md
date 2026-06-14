# FabricOps DeepAgent

FabricOps DeepAgent is the VPS infrastructure operator agent for the Unboxd Platform / Fabric runtime.

It wraps LangChain DeepAgents as the reasoning harness and keeps platform authority in the existing operating model:

```text
GitHub -> CI/CD -> k3s -> Agent CRD -> Reconciler -> SurrealDB -> Fabric Runtime
```

## Mission

Continuously reconcile declared platform intent against VPS, Kubernetes, SurrealDB, GitHub, and security reality.

```text
Desired State
  -> Fabric Graph
  -> FabricOps DeepAgent
  -> k3s / VPS / GitHub / SurrealDB
  -> Observed Reality
  -> Fabric Evidence
  -> Human Gate when needed
```

## Runtime role

DeepAgents is used as the agent harness only. It does not own infrastructure authority.

FabricOps must pass every privileged action through:

- Fabric policy
- OPA decision check
- OpenFGA relationship check where applicable
- human approval for destructive or high-risk actions
- SurrealDB audit/event write

## Sub-agents

| Sub-agent | Responsibility |
| --- | --- |
| `vps_agent` | Host health, systemd, disk, memory, network, firewall, SSH posture |
| `kubernetes_agent` | k3s node/pod/event health, deployment rollout, namespace safety |
| `surrealdb_agent` | SurrealDB health, schema checks, backup state, runtime evidence |
| `github_agent` | PRs, actions, releases, security alerts, deployment provenance |
| `security_agent` | open ports, failed logins, privilege escalation, leaked secrets |
| `cost_agent` | CPU, memory, storage, network, utilization and savings recommendations |
| `deployment_agent` | build, scan, deploy, verify, promote loop |
| `human_approval_agent` | approval gates and decision records |

## Deny-by-default gates

FabricOps must never auto-execute these without explicit human approval:

- delete namespace
- delete PVC / PV
- delete database
- delete node
- rotate secrets
- open firewall port
- disable security controls
- force push branch
- deploy unscanned image

## Repository layout

```text
agents/fabricops-deepagent/
  README.md
  agent.yaml
  policies/
    fabricops.rego
  prompts/
    system.md
  tools/
    allowlist.yaml
  manifests/
    namespace.yaml
    configmap.yaml
    deployment.yaml
    serviceaccount.yaml
  scripts/
    run-local.sh
```

## Environment

Required runtime variables:

```bash
SURREAL_URL=ws://surrealdb.fabric.svc.cluster.local:8000
SURREAL_NS=agennext
SURREAL_DB=fabric
OPA_URL=http://opa.open-policy-agent.svc.cluster.local:8181
OPENFGA_API_URL=http://openfga.openfga.svc.cluster.local:8080
FABRICOPS_MODE=observe
```

`FABRICOPS_MODE` values:

| Mode | Meaning |
| --- | --- |
| `observe` | Read-only monitoring and evidence writing |
| `recommend` | Produce repair plans, no execution |
| `approve` | Execute only actions with explicit approval record |
| `auto-safe` | Execute allowlisted non-destructive repairs |

Production default is `observe`.

## First VPS checks

```bash
kubectl get nodes -o wide
kubectl get pods -A
kubectl get events -A --sort-by=.lastTimestamp
kubectl top nodes
kubectl top pods -A
systemctl status k3s
journalctl -u k3s --since "1 hour ago"
df -h
free -h
ss -tulpn
```

## Output contract

Every action must emit a Fabric decision event:

```json
{
  "agent": "fabricops-deepagent",
  "mode": "observe",
  "action": "check_k3s_pods",
  "risk": "low",
  "decision": "allowed",
  "approval_required": false,
  "evidence": {},
  "timestamp": "RFC3339"
}
```

## Current status

Scaffold only. Wire this to the DeepAgents Python runtime and the existing platform operator loop next.
