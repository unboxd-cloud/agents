# Unboxd Cloud Platform Reference Architecture

## Purpose

This document maps Unboxd Cloud capabilities to platform services and reference technologies.

Related artifacts:

- [Platform Constitution](./platform-constitution.md)
- [Platform Principles](./platform-principles.md)
- [Business Model](./business-model.md)
- [Platform Capability Model](./platform-capability-model.md)

---

## Architecture Goal

Unboxd Cloud must provide a composable, open-source, Kubernetes-mature, cloud-native, agent-native platform foundation that can support:

- Multiple tenants
- Multiple models
- Multiple clouds
- Multiple frameworks
- Multiple delivery modes
- Multiple channels
- Multiple governance needs
- Multiple industries
- Usage-based pricing
- Managed operations
- Enterprise readiness

---

## Reference Architecture Layers

```text
User / Customer / Partner / Agent Channels
  ↓
Experience & API Layer
  ↓
Control Plane
  ↓
Governance Plane
  ↓
Agent & Automation Plane
  ↓
Data, Memory & Knowledge Plane
  ↓
Deployment & GitOps Plane
  ↓
Runtime Plane
  ↓
Cloud / Edge / Compute Plane
  ↓
Observability, Security & Operations Plane
```

The planes are logical. Some services may span multiple planes.

---

## 1. Experience & API Layer

Purpose:

> Provide interfaces for customers, developers, operators, partners, and agents.

Capabilities served:

- Channel & Integration
- Support & Service Management
- Tenant & Workspace Management
- Marketplace & Partner Ecosystem

Reference services:

| Service | Responsibility |
|---|---|
| Web Console | Customer and operator interface |
| Admin Console | Tenant, policy, billing, support, governance controls |
| Developer Portal | Docs, APIs, SDKs, CLI, templates |
| Partner Portal | Partner onboarding, listings, support routing |
| API Gateway | Public and internal platform APIs |
| CLI | Developer and operator command interface |
| Event Gateway | Webhooks, CloudEvents, event subscriptions |
| Agent Tool Gateway | MCP tools and A2A endpoints |

Reference technologies:

- Web console: Next.js / React / Astro where appropriate
- API: Go / Rust / FastAPI / NestJS depending on service boundary
- Gateway: Traefik / Envoy-ready
- Events: CloudEvents
- Agent protocols: MCP, A2A, OpenAPI

---

## 2. Control Plane

Purpose:

> Coordinate tenants, workspaces, deployments, applications, agents, workflows, usage, billing, and support context.

Capabilities served:

- Tenant & Workspace Management
- Billing & Pricing
- Support & Service Management
- Industry Solution Composition
- Deployment & GitOps

Reference services:

| Service | Responsibility |
|---|---|
| Tenant Service | Organizations, workspaces, environments, tenant isolation |
| Project Service | Apps, agents, workflows, integrations grouped by project |
| Solution Service | Sector and segment solution packages |
| Catalog Service | Templates, modules, integrations, agents, workflows |
| Billing Service | Plans, subscriptions, usage lines, invoices |
| Usage Metering Service | Meter compute, storage, agents, workflows, support, retention |
| Support Service | Tickets, incidents, reviews, escalation, SLAs |
| Configuration Service | Tenant and environment configuration |

Reference technologies:

- SurrealDB / Postgres depending on workload
- OpenAPI for service interfaces
- CloudEvents for platform events
- OpenCost integration for infrastructure usage

---

## 3. Governance Plane

Purpose:

> Govern identity, access, data, cost, agents, security, operations, compliance, and open-source use.

Capabilities served:

- Identity & Access
- Governance & Policy
- Compliance & Audit
- Security & Secrets
- Cost & Usage Management
- Agent Runtime

Reference services:

| Service | Responsibility |
|---|---|
| Identity Service | Human, agent, service, tool, and organization identity |
| Access Service | RBAC, ABAC, ReBAC, relationship authorization |
| Policy Service | Policy-as-code, approvals, enforcement decisions |
| Audit Service | Audit logs, evidence, export, retention |
| Trust Service | Trust scoring, provenance, risk, reputation, decision confidence |
| Governance Service | Reviews, controls, compliance mappings, governance workflows |
| Approval Service | Human approval, change gates, exception handling |

Reference technologies:

- OpenFGA for relationship authorization
- OPA for policy-as-code
- AuthZEN-ready authorization API patterns
- DID / VC-ready identity layer
- JSON-LD / schema.org for evidence and metadata

---

## 4. Agent & Automation Plane

Purpose:

> Run agents, workflows, automations, tool calls, approvals, and human-agent collaboration safely.

Capabilities served:

- Agent Runtime
- Workflow & Automation
- Model Management
- Memory & Knowledge
- Governance & Policy
- Observability & Operations

Reference services:

| Service | Responsibility |
|---|---|
| Agent Runtime Service | Agent lifecycle, runs, tools, policy bindings, escalation |
| Model Router Service | Model selection, routing, fallback, cost and quality tracking |
| Tool Registry Service | Tool definitions, permissions, schemas, MCP endpoints |
| Workflow Service | Durable workflows, jobs, approvals, retries, replay |
| Evaluation Service | Agent, model, workflow, and output evaluation |
| Human Approval Service | Human-in-the-loop approval and escalation |
| Agent Trace Service | Runs, steps, tool calls, model calls, evidence, outputs |

Reference technologies:

- MCP
- A2A
- OpenAPI
- CloudEvents
- Temporal / Argo Workflows / Dapr Workflows where appropriate
- LangGraph / LlamaIndex / custom agent runtime where appropriate

---

## 5. Data, Memory & Knowledge Plane

Purpose:

> Represent business reality, memory, knowledge, evidence, facts, claims, predictions, simulations, and relationships.

Capabilities served:

- Memory & Knowledge
- Data & Graph
- Compliance & Audit
- Agent Runtime
- Industry Solution Composition

Reference services:

| Service | Responsibility |
|---|---|
| Graph Service | Entities, relationships, actions, events, timelines |
| Memory Service | Human, agent, organization, and project memory |
| Knowledge Service | Knowledge bases, documents, sources, evidence |
| Search Service | Full-text, semantic, graph, and filtered search |
| Lineage Service | Data provenance, source tracking, transformation history |
| Reality Classification Service | Fact, claim, inference, prediction, simulation, goal typing |
| Data Governance Service | Classification, retention, residency, export, access rules |

Reference technologies:

- SurrealDB for graph, document, relation, and temporal data
- JSON-LD / schema.org for semantic metadata
- Vector search where needed
- CloudEvents for event representation

---

## 6. Deployment & GitOps Plane

Purpose:

> Keep desired state versioned, reviewed, reconciled, and auditable.

Capabilities served:

- Deployment & GitOps
- OCI Artifact Management
- Cloud & Infrastructure
- Application Runtime
- Governance & Policy

Reference services:

| Service | Responsibility |
|---|---|
| GitOps Service | Desired state repositories, reconciliation, drift detection |
| Release Service | Release records, promotion, rollback, environment history |
| Infrastructure State Service | OpenTofu state, plan, apply, drift, environment mapping |
| Environment Service | Dev, staging, production, tenant environments |
| Deployment Policy Service | Deployment gates, approvals, policy checks |
| Change Management Service | Change requests, reviews, impact, rollback path |

Reference technologies:

- Flux / ArgoCD
- OpenTofu
- Forgejo / GitHub / Forgejo Actions
- Helm / Kustomize
- Kubernetes manifests

---

## 7. Runtime Plane

Purpose:

> Run applications, APIs, agents, workflows, services, jobs, and internal tools reliably.

Capabilities served:

- Application Runtime
- Agent Runtime
- Workflow & Automation
- Cloud & Infrastructure
- Observability & Operations
- Security & Secrets

Reference services:

| Service | Responsibility |
|---|---|
| Kubernetes Runtime | Workload scheduling, services, jobs, config, storage |
| Ingress Service | HTTP routing, TLS termination, external exposure |
| Certificate Service | TLS certificate provisioning and renewal |
| Storage Service | Persistent volumes, object storage, backups |
| Secrets Service | Secrets storage, injection, rotation |
| Service Mesh Layer | Optional mTLS, retries, traffic policy, service identity |
| Serverless Layer | Optional event-driven and scale-to-zero workloads |

Reference technologies:

- K3s / Kubernetes
- Traefik
- cert-manager
- containerd
- local-path provisioner / CSI-ready storage
- Infisical
- Linkerd / Cilium / Istio when needed
- Knative when needed

---

## 8. OCI Artifact Plane

Purpose:

> Store, distribute, verify, promote, and govern deployable artifacts.

Capabilities served:

- OCI Artifact Management
- Deployment & GitOps
- Security & Secrets
- Compliance & Audit

Reference services:

| Service | Responsibility |
|---|---|
| Registry Service | Container, Helm, WASM, and OCI artifacts |
| Artifact Metadata Service | Versions, provenance, SBOMs, signatures, ownership |
| Artifact Promotion Service | Dev to staging to production promotion |
| Artifact Policy Service | Signing, vulnerability, license, provenance gates |
| Artifact Retention Service | Cleanup, lifecycle, retention policy |

Reference technologies:

- Harbor
- OCI
- Cosign-ready signing
- SBOM-ready metadata
- Vulnerability scanning through registry integrations

---

## 9. Observability, Security & Operations Plane

Purpose:

> Make the platform measurable, secure, recoverable, supportable, and continuously improvable.

Capabilities served:

- Observability & Operations
- Security & Secrets
- Cost & Usage Management
- Support & Service Management
- Compliance & Audit
- Learning & Improvement Loop

Reference services:

| Service | Responsibility |
|---|---|
| Telemetry Service | Metrics, logs, traces, events |
| Dashboard Service | Grafana dashboards, customer and operator views |
| Alerting Service | Alerts, routing, incident triggers |
| Incident Service | Incident records, response, postmortems |
| Backup Service | Backup execution, verification, restore evidence |
| Cost Service | Usage, cost allocation, budgets, showback / chargeback |
| Security Posture Service | Vulnerability, configuration, secrets, policy posture |
| Improvement Service | Feedback, incidents, usage insights, roadmap items |

Reference technologies:

- OpenTelemetry
- Prometheus
- Grafana
- Loki / Tempo-ready
- OpenCost
- Infisical
- Kubernetes events and metrics

---

## Capability to Service Mapping

| Capability Domain | Primary Services | Reference Technologies |
|---|---|---|
| Tenant & Workspace Management | Tenant Service, Project Service, Configuration Service | SurrealDB / Postgres, OpenAPI |
| Identity & Access | Identity Service, Access Service | OpenFGA, OPA, SSO / SCIM-ready |
| Governance & Policy | Policy Service, Governance Service, Approval Service | OPA, OpenFGA, policy-as-code |
| Agent Runtime | Agent Runtime Service, Tool Registry, Agent Trace Service | MCP, A2A, OpenAPI, CloudEvents |
| Model Management | Model Router Service, Evaluation Service | Provider APIs, local models, evaluation store |
| Memory & Knowledge | Memory Service, Knowledge Service, Search Service | SurrealDB, JSON-LD, vector search |
| Data & Graph | Graph Service, Lineage Service, Data Governance Service | SurrealDB, schema.org, CloudEvents |
| Workflow & Automation | Workflow Service, Approval Service | Temporal, Argo Workflows, Dapr Workflows |
| Application Runtime | Kubernetes Runtime, Ingress Service, Certificate Service | K3s, Traefik, cert-manager |
| Cloud & Infrastructure | Infrastructure State Service, Environment Service | OpenTofu, Kubernetes |
| Deployment & GitOps | GitOps Service, Release Service, Change Management Service | Flux, ArgoCD, Forgejo, Helm, Kustomize |
| OCI Artifact Management | Registry Service, Artifact Policy Service | Harbor, OCI, Cosign-ready signing |
| Observability & Operations | Telemetry Service, Dashboard Service, Incident Service | OpenTelemetry, Prometheus, Grafana |
| Security & Secrets | Secrets Service, Security Posture Service | Infisical, OPA, Harbor scanning |
| Cost & Usage Management | Usage Metering Service, Cost Service | OpenCost, Kubernetes metrics |
| Billing & Pricing | Billing Service, Usage Metering Service | Usage records, invoices, plans |
| Support & Service Management | Support Service, Incident Service | Tickets, SLAs, runbooks |
| Channel & Integration | API Gateway, Event Gateway, Agent Tool Gateway | API, CLI, Webhooks, MCP, A2A |
| Industry Solution Composition | Solution Service, Catalog Service | Templates, blueprints, modules |
| Marketplace & Partner Ecosystem | Partner Portal, Catalog Service | Listings, certifications, approvals |
| Compliance & Audit | Audit Service, Governance Service | Audit logs, evidence records |
| Learning & Improvement Loop | Improvement Service, Evaluation Service | Feedback, incidents, telemetry, roadmap |

---

## Logical Deployment Modes

### Shared Multi-Tenant

Best for:

- Small businesses
- Developers
- Early-stage startups
- Low-regulation workloads

Characteristics:

- Shared control plane
- Shared runtime clusters
- Tenant isolation through namespace, policy, identity, data, and billing boundaries

### Dedicated Tenant

Best for:

- Growing companies
- Mid-market customers
- Higher isolation needs

Characteristics:

- Dedicated namespace or cluster
- Dedicated configuration
- Stronger isolation
- Optional dedicated data layer

### Private Cloud

Best for:

- Enterprises
- Regulated customers
- Data-sensitive workloads

Characteristics:

- Customer-owned environment
- Dedicated runtime
- Enterprise identity integration
- Custom governance

### On-Prem / Edge

Best for:

- Factories
- Logistics
- Clinics
- Retail branches
- Low-latency or local-control use cases

Characteristics:

- K3s / edge cluster
- Local workloads
- Optional sync to central control plane

### Air-Gapped

Best for:

- Government
- Defense
- Highly regulated sovereign environments

Characteristics:

- No public internet dependency
- Offline artifact distribution
- Local registry
- Local observability
- Manual or controlled sync

---

## Reference Technical Foundation

```text
Compute / VPS / Bare Metal / Cloud / Edge
  ↓
K3s / Kubernetes
  ↓
Traefik + cert-manager
  ↓
Flux / ArgoCD + OpenTofu
  ↓
Forgejo + Harbor
  ↓
Infisical + OPA + OpenFGA
  ↓
SurrealDB
  ↓
OpenTelemetry + Prometheus + Grafana + OpenCost
  ↓
Agent Runtime + Model Router + Workflow Engine
  ↓
Solutions, Channels, APIs, Agents, Apps
```

---

## Minimum Viable Production Stack

For a single-node or small-cluster foundation:

```text
K3s
Traefik
cert-manager
Flux or ArgoCD
Forgejo
Harbor
Infisical
OpenFGA
OPA
SurrealDB
OpenTelemetry
Prometheus
Grafana
OpenCost
Agent Runtime
```

Optional by need:

```text
Temporal
Argo Workflows
Dapr
Linkerd / Cilium / Istio
Knative
Loki
Tempo
Object storage
External managed database
```

---

## Architecture Principles

The reference architecture must remain:

- Reality-anchored
- Open-source first
- Kubernetes-mature
- CNCF-aligned
- Cloud-native mature
- Agent-native by design
- Multi-tenant
- Multi-model
- Multi-cloud
- Multi-framework
- Multi-delivery
- Multi-channel
- Multi-governance
- Multi-industry
- Composable
- Usage-based
- Enterprise-ready

---

## Final Statement

> The Unboxd Cloud reference architecture maps business capabilities into composable platform services and open-source technologies, creating a Kubernetes-mature, cloud-native, agent-native foundation that can be delivered across tenants, clouds, models, frameworks, channels, industries, and governance needs.
