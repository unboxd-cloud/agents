# Solution Catalog

The Solution Catalog is the customer-facing service catalog for Unboxd Platform.
It is anchored to the uploaded CNCF landscape sheets:

- `projects.csv` provides project maturity, accepted date, security audit count, and last audit date.
- `items.csv` provides broader landscape metadata such as category, subcategory, homepage, repository, license, and organization.

The catalog must expose **services**, not raw projects. Customers consume services; the platform composes projects as backends.

## Catalog model

```text
Solution Catalog
  -> Service
    -> Capability
      -> Backend Project
        -> Runtime
        -> Policy
        -> Meter
        -> Compliance
        -> SLA
```

## Naming rules

- **Solution Catalog** = what the customer consumes.
- **Project Registry** = what the platform composes.
- **Backend Map** = how a service is fulfilled.

A sandbox CNCF project must not be presented as a default enterprise service. It can be exposed as `preview` or `optional` with explicit maturity and audit metadata.

## Service exposure policy

| Backend maturity | Security audit signal | Catalog exposure |
|---|---:|---|
| graduated | one or more audits | `default` |
| graduated | no audit in sheet | `standard` |
| incubating | any | `standard` |
| sandbox | any | `preview` |
| not found in `projects.csv` | unknown | `optional` |

## Seed service catalog

The machine-readable catalog lives in [`deploy/datasets/solution-catalog.json`](../deploy/datasets/solution-catalog.json).

| Service | Category | Compatibility | Backend | Sheet maturity | Exposure |
|---|---|---|---|---|---|
| Compute Service | compute | EC2-compatible | Kubernetes | graduated | default |
| Virtual Machine Service | compute | EC2-compatible | KubeVirt | incubating | standard |
| Batch Scheduling Service | compute | internal-api | Armada | sandbox | preview |
| Serverless Function Service | serverless | Lambda-compatible | Knative | graduated | default |
| Object Storage Service | storage | S3-compatible | Rook | graduated | default |
| Security Token Service | identity | STS-compatible | SPIFFE | graduated | standard |
| Identity Provider Service | identity | OIDC-compatible | Dex | sandbox | preview |
| Event Notification Service | messaging | SNS-compatible | NATS | incubating | standard |
| Streaming Service | messaging | Kafka-compatible where applicable | Strimzi | incubating | standard |
| API Gateway Service | networking | internal-api | APISIX | external/unknown | optional |
| GitOps Deployment Service | operations | GitOps | Argo | graduated | default |
| Policy Service | governance | policy-api | Open Policy Agent | external/unknown | optional |
| Relationship Authorization Service | governance | authz-api | OpenFGA | incubating | standard |
| Observability Service | observability | OTEL/Prometheus | Prometheus | graduated | default |
| Telemetry Pipeline Service | observability | OTEL | OpenTelemetry | graduated | standard |
| Cost Metering Service | billing | OpenCost API | OpenCost | incubating | standard |
| Model Serving Service | ai | Bedrock-compatible facade | KServe | incubating | standard |
| ML Pipelines Service | ai | internal-api | Kubeflow | incubating | standard |
| Agent Runtime Service | ai | AgentCore-compatible facade | Dapr | graduated | default |
| Developer Portal Service | dev | portal-api | Backstage | incubating | standard |
| Container Registry Service | dev | OCI registry | Harbor | graduated | default |
| Secrets Management Service | security | secrets-api | External Secrets Operator | external/unknown | optional |
| Certificate Management Service | security | cert-api | cert-manager | graduated | standard |
| Service Mesh Service | networking | mesh-api | Linkerd | graduated | default |

## Important interpretation

This catalog is intentionally **service-first**. For example, Armada is not the product. The product is **Batch Scheduling Service**. Armada is only the backend candidate, and because the sheet marks it as CNCF sandbox with zero audits, the service exposure is `preview`.
