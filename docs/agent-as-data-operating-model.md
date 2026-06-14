# Agent-as-Data Operating Model

## Core loop

```text
GitHub → CI/CD → k3s → Agent CRD → Java Reconciler Pod → SurrealDB → Fabric Runtime
```

## Principles

- **Agent = Data**: an agent is declared, versioned, reviewed, and reconciled as data.
- **Fabric = Runtime**: Fabric executes governed work from trusted runtime state.
- **Kubernetes = Reconciliation Engine**: Kubernetes stores desired state and continuously drives actual state toward it.
- **SurrealDB = Source of Runtime Truth**: reconciled agents become queryable runtime records.
- **Human at Gate**: GitHub PRs, CI checks, and policy gates control what enters the runtime.

## Reconciliation flow

1. GitHub stores desired state as code and data.
2. CI/CD validates source, manifests, build outputs, and publish artifacts.
3. GitHub publishes the reconciler image and package artifacts.
4. k3s pulls the reconciler image and runs the platform control plane.
5. Kubernetes hosts the Agent CRD and Agent objects.
6. The Java reconciler watches `agents.fabric.agennext.io`.
7. When an Agent is added, changed, or deleted, the reconciler updates SurrealDB.
8. SurrealDB exposes the runtime graph of agents, identity, policy, state, and audit.
9. Fabric Runtime reads SurrealDB and executes governed work.
10. The loop repeats continuously.

## Platform placement

This operating model belongs in the platform repo because the platform is the control plane. The `Agent-As-Data` repo can hold a focused reference implementation, while `unboxd-cloud/platform` defines how the full platform composes CI/CD, Kubernetes, policy, billing, observability, and runtime reconciliation.

## Runtime contract

An Agent must carry enough data to be reconciled, governed, and executed:

- identity
- owner
- objective
- lifecycle state
- trust score
- runtime mode
- approvals
- skills
- tools
- policies
- constraints
- Kubernetes metadata
- audit timestamps

## Proof command

For the reference implementation:

```bash
cd ~/Agent-As-Data && git pull && chmod +x scripts/self-check.sh scripts/k3s-reconcile-check.sh && scripts/k3s-reconcile-check.sh
```

Expected proof:

```text
Agent CRD exists
fabric-architect Agent exists
Java reconciler pod is Running
SurrealDB contains agent:fabric_architect
```
