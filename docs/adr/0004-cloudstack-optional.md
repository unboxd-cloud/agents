# ADR-0004: Apache CloudStack is optional, not foundational

## Status
Accepted

## Context
The original brief named Apache CloudStack as the base. CloudStack is a capable
but **heavyweight** IaaS orchestrator: a JVM management server, a MySQL
database, and agents that drive hypervisors (KVM/VMware/XenServer). For a
platform that is **Kubernetes-native, CNCF-native, multi-cloud, and edge-aware**,
most of that surface overlaps with tools we already compose:

- **Provisioning / lifecycle** → Crossplane + Cluster API
- **Multi-cloud reach** → Kubernetes on any cloud, behind the `Provider` seam
- **Edge** → K3s / KubeEdge
- **Tenancy** → Capsule / vCluster

Running CloudStack *and* the CNCF control plane means two overlapping control
planes to operate, secure, and upgrade.

## Decision
Do **not** make CloudStack foundational. Keep it as **one optional `Provider`
implementation** behind `provider.Provider`, registered only when needed.

**Recommended default:** Kubernetes + Crossplane + Cluster API for compute and
provisioning; add the CloudStack provider only when there is a hard requirement
to orchestrate **traditional VMs / on-prem hypervisors without Kubernetes**
(e.g., an existing CloudStack estate to absorb).

## Consequences
- Lighter default footprint; one cloud-native control plane.
- CloudStack remains a drop-in option — no architectural change to adopt it,
  satisfying the original brief without paying its cost by default.
- We avoid coupling billing/metering/catalog to any CloudStack specifics.
