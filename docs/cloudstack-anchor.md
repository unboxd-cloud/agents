# Apache CloudStack Anchor

Unboxd Cloud Platform uses Apache CloudStack as the default open IaaS anchor.

CloudStack provides the infrastructure substrate. Unboxd adds AWS-compatible APIs, tenant experience, service catalog, metering, billing, compliance, and operational dashboards above it.

## Positioning

```text
Apache CloudStack = open IaaS anchor
Kubernetes/k3s    = workload runtime layer
Unboxd control    = catalog, tenant, metering, billing, compliance
Unboxd APIs       = AWS-compatible product surface
```

## Why CloudStack

CloudStack is the anchor because the platform needs an open, production-proven cloud foundation that can be operated by:

- regional cloud providers
- on-prem enterprise teams
- labs and universities
- public sector clouds
- managed service providers
- edge operators
- community cloud operators

The goal is not to rebuild IaaS. The goal is to compose it into a governed, metered cloud product.

## Layering

```text
Experience Layer
  admin panel · org console · CLI · SDK · Backstage

Product API Layer
  AWS-compatible APIs · tenant APIs · catalog APIs · billing APIs

Control Plane Layer
  catalog · tenant · metering · billing · compliance · operator

Runtime Layer
  Kubernetes/k3s · Knative/OpenFaaS · KubeVirt when needed

IaaS Anchor
  Apache CloudStack zones · pods · clusters · hosts · storage · networks
```

## CloudStack mapping

| CloudStack primitive | Unboxd Platform meaning |
| --- | --- |
| Zone | Region or provider location |
| Pod | Availability boundary inside a zone |
| Cluster | Compute pool |
| Host | Physical or virtual capacity provider |
| Domain | Organization or reseller boundary |
| Account | Tenant account |
| Project | Shared workspace or customer project |
| Role | Permission boundary |
| Template | Approved base image |
| Service Offering | Compute product SKU |
| Disk Offering | Storage product SKU |
| Network Offering | Network product SKU |
| Event | Metering, audit, and operations source |

## MVP capabilities

| Capability | CloudStack source | Unboxd surface |
| --- | --- | --- |
| Compute | Virtual machines, templates, service offerings | EC2-style compute API |
| Tenant management | Domains, accounts, projects | Org and tenant console |
| Metering | Usage records and events | Usage API and billing engine |
| Storage | Volumes, snapshots, disk offerings | Storage SKU and object gateway integration |
| Network | Networks, IPs, firewall, load balancing | Network product API |
| Images | Templates and ISOs | Image catalog |
| Audit | Events and async jobs | Audit and compliance records |

## Operating flow

```text
CloudStack inventory/events
  -> Unboxd ingestion
  -> normalized platform records
  -> metering
  -> rating
  -> billing
  -> compliance evidence
  -> dashboard/API output
```

## Rules

- CloudStack is the default IaaS provider.
- Kubernetes/k3s runs platform services and cloud-native workloads.
- Billing and metering must use normalized records, not provider-specific leakage.
- CloudStack events must be preserved for audit.
- Product APIs should hide provider complexity.
- Provider seams must remain open for edge and additional clouds.

## First implementation target

1. CloudStack client package
2. CloudStack inventory sync
3. CloudStack usage ingestion
4. CloudStack service-offering catalog mapping
5. CloudStack tenant/account mapping
6. CLI command: `platform cloudstack map`
7. Metering records from CloudStack usage data
