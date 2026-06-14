# ADR-0007: Apache CloudStack as the contract, Kubernetes as the reconciler

## Status
Accepted

## Context
The original brief named Apache CloudStack as the base. ADR-0004 decided not to
run CloudStack as foundational infrastructure (its JVM management server + MySQL
+ hypervisor agents duplicate what the CNCF stack already gives us) and kept it
as one optional `provider.Provider`.

That still leaves a real requirement: **CloudStack clients, SDKs, and tooling**
should be able to drive the platform. The platform is already "AWS-compatible,
CNCF-native" (speak the AWS API, implement it with open-source CNCF projects).
We want the same move for CloudStack: speak the CloudStack API, implement it on
Kubernetes — without adopting the heavyweight CloudStack server ADR-0004 rejected.

## Decision
Adopt the **Apache CloudStack API as the northbound contract** and **Kubernetes
as the reconciler** that realizes it.

- `internal/cloudstack` — the **contract**: CloudStack-shaped request/response
  types and the VM lifecycle (`Starting → Running → Stopping → Stopped →
  Destroyed`). It says nothing about the substrate.
- `internal/kube` — the **pod manager**: a `PodManager` seam over Kubernetes pod
  lifecycle (in-memory stub in Phase 0, like `internal/provider`; a real
  client-go/KubeVirt manager drops in behind the same interface).
- `internal/controlplane` — the **cloud control plane**: implements
  `cloudstack.Contract` and is an `agent.Agent`. Writes record desired state; a
  level-triggered `Reconcile` loop converges actual pods toward it. Each VM is
  backed by a pod (template → image, service offering → CPU/memory).
- `cmd/cloud` — the control-plane service (`:8086`). It serves a clean REST API
  under `/v1` **and** a CloudStack-compatible `/client/api?command=...` endpoint,
  and runs the operators in the background.

## Single core, multi store, multi operator, multi mode
The control plane keeps no state of its own — desired state lives in a `Store`:

- **single core** — one `agent.Operator` runtime drives the reconcilers.
- **multi store** — `Store` has `MemStore` (default) and `FileStore` (JSON,
  persistent) implementations, selected by `CLOUD_STORE` (`mem`|`file`). A
  persistent store can be the shared source of truth across processes.
- **multi operator** — the core runs the control-plane reconciler **and** a pod
  `Reaper` (garbage-collects pods whose VM is gone) over the same store.
- **multi mode** — `CLOUD_MODE` selects which planes run: `all` (default),
  `api` (serve the API, write desired state), or `operator` (reconcile only).
  With `CLOUD_STORE=file`, `api` + `operator` split like apiserver /
  controller-manager over one store.

## Relationship to ADR-0004
This **complements** ADR-0004; it does not reverse it. ADR-0004 rejects the
CloudStack *server* as a foundation. ADR-0007 uses the CloudStack *API* as a
compatibility contract while **Kubernetes does the reconciliation**. Both hold
at once: CloudStack is the API, Kubernetes is the engine. The `cloudstack`
provider stub (`provider.NewCloudStack`) is unchanged and orthogonal.

## Consequences
- Existing CloudStack clients can target the platform via `/client/api`.
- Compute is realized as Kubernetes pods, reusing the platform's CNCF substrate;
  no second control plane to operate (the ADR-0004 concern).
- Billing/metering/catalog stay decoupled from CloudStack specifics — the
  contract lives only in the new control-plane packages.
- Phase 0 ships an in-memory `PodManager`; the reconcile path is real and tested.

## Follow-ups
- Real `PodManager` (client-go; KubeVirt for full VMs per the compute module).
- Wire `cloud` into the Helm chart and local sandbox.
- Grow the contract: networks, volumes, snapshots, async job IDs.
