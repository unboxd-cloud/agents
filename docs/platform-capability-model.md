# Unboxd Cloud Platform Capability Model

## Purpose

This document defines the capabilities Unboxd Cloud must provide to deliver on its constitution, principles, business model, and Solutions-as-a-Service delivery model.

Related artifacts:

- [Platform Constitution](./platform-constitution.md)
- [Platform Principles](./platform-principles.md)
- [Business Model](./business-model.md)

---

## Capability Architecture

```text
Business Architecture
  ↓
Capability Architecture
  ↓
Solution Architecture
  ↓
Technical Architecture
  ↓
Implementation
```

The capability model answers:

> What must the platform be able to do?

---

## Capability Domains

```text
1. Tenant & Workspace Management
2. Identity & Access
3. Governance & Policy
4. Agent Runtime
5. Model Management
6. Memory & Knowledge
7. Data & Graph
8. Workflow & Automation
9. Application Runtime
10. Cloud & Infrastructure
11. Deployment & GitOps
12. OCI Artifact Management
13. Observability & Operations
14. Security & Secrets
15. Cost & Usage Management
16. Billing & Pricing
17. Support & Service Management
18. Channel & Integration
19. Industry Solution Composition
20. Marketplace & Partner Ecosystem
21. Compliance & Audit
22. Learning & Improvement Loop
```

---

## 1. Tenant & Workspace Management

Purpose:

> Support many customers, teams, workspaces, environments, and deployment models while preserving isolation.

Capabilities:

- Organization onboarding
- Tenant creation
- Workspace creation
- Environment management
- Team management
- Tenant isolation
- Dedicated tenant support
- Private deployment support
- Edge / on-prem / air-gapped tenant support
- Tenant-level configuration
- Tenant-level policy
- Tenant-level billing
- Tenant-level support context

Key objects:

```text
Tenant
Workspace
Organization
Team
Environment
Project
```

---

## 2. Identity & Access

Purpose:

> Make every human, agent, service, tool, and organization addressable, governable, and auditable.

Capabilities:

- User identity
- Agent identity
- Organization identity
- Service identity
- Tool identity
- API identity
- SSO integration
- SCIM provisioning
- RBAC
- ABAC
- ReBAC
- Delegation
- Access review
- Session management
- API keys
- Token lifecycle
- DID / VC readiness

Key objects:

```text
Actor
Human
Agent
Organization
Service
Tool
Role
Permission
Relationship
Credential
```

Reference technologies:

- OpenFGA
- OPA
- AuthZEN-ready APIs
- DID / VC-ready identity layer

---

## 3. Governance & Policy

Purpose:

> Ensure every action can be governed, explained, approved, denied, audited, and improved.

Capabilities:

- Policy-as-code
- Approval workflows
- Governance rules
- Change controls
- Risk scoring
- Trust scoring
- Access governance
- Data governance
- Agent governance
- Cost governance
- Operational governance
- Open-source governance
- Policy evaluation
- Policy simulation
- Policy audit
- Kill switch for agents and automations

Key objects:

```text
Policy
Rule
Approval
Risk
TrustScore
Decision
Evidence
AuditRecord
```

Reference technologies:

- OPA
- OpenFGA
- Policy-as-code

---

## 4. Agent Runtime

Purpose:

> Run agents as governed actors that can observe, reason, act, learn, and collaborate with humans and systems.

Capabilities:

- Agent creation
- Agent identity
- Agent lifecycle
- Agent configuration
- Tool binding
- Memory binding
- Policy binding
- Model binding
- Agent-to-agent communication
- Human approval
- Escalation
- Evaluation
- Traceability
- Agent suspension
- Agent termination
- Agent audit

Key objects:

```text
Agent
AgentIdentity
AgentPolicy
AgentMemory
AgentTool
AgentRun
AgentTrace
AgentEvaluation
```

Protocols:

- MCP
- A2A
- OpenAPI
- Webhooks
- CloudEvents

---

## 5. Model Management

Purpose:

> Support the right model for the right task across cloud, open-source, local, domain-specific, and enterprise-approved models.

Capabilities:

- Model registry
- Model approval
- Model routing
- Model cost tracking
- Model latency tracking
- Model quality tracking
- Model fallback
- Model policy enforcement
- Embedding model support
- Vision model support
- Speech model support
- Reranking model support
- Local model support
- Provider abstraction

Key objects:

```text
Model
ModelProvider
ModelPolicy
ModelRoute
ModelUsage
ModelEvaluation
```

---

## 6. Memory & Knowledge

Purpose:

> Preserve context, evidence, history, and reusable knowledge without confusing facts, claims, predictions, and simulations.

Capabilities:

- Agent memory
- User memory
- Organization memory
- Knowledge base
- Document ingestion
- Semantic search
- Full-text search
- Source attribution
- Evidence tracking
- Memory versioning
- Knowledge graph
- Fact / claim / inference classification
- Real vs surreal separation

Key objects:

```text
Memory
KnowledgeItem
Document
Source
Evidence
Fact
Claim
Inference
Prediction
Simulation
```

Reference technologies:

- SurrealDB
- Vector search
- Full-text search
- JSON-LD / schema.org

---

## 7. Data & Graph

Purpose:

> Represent business reality as identities, relationships, actions, events, memory, and time.

Capabilities:

- Entity graph
- Relationship graph
- Event graph
- Temporal records
- Data modeling
- Schema validation
- JSON-LD support
- Data lineage
- Data classification
- Data retention
- Data residency
- Data access policy
- Data export
- Data backup
- Data restore

Key objects:

```text
Entity
Relationship
Action
Event
State
Timeline
Schema
Lineage
```

Reference technologies:

- SurrealDB
- JSON-LD
- schema.org
- CloudEvents

---

## 8. Workflow & Automation

Purpose:

> Convert intent, decisions, approvals, and events into reliable operational flows.

Capabilities:

- Workflow creation
- Workflow execution
- Workflow scheduling
- Event triggers
- Approval steps
- Human-in-the-loop
- Retry
- Replay
- Failure handling
- Dead-letter handling
- Workflow audit
- Workflow metrics
- Workflow templates
- Agent workflow execution

Key objects:

```text
Workflow
Step
Trigger
Schedule
ApprovalStep
Run
Event
Retry
```

Reference technologies:

- Temporal
- Argo Workflows
- Dapr Workflows
- CloudEvents

---

## 9. Application Runtime

Purpose:

> Run customer applications, APIs, websites, automations, and internal tools reliably.

Capabilities:

- App deployment
- API deployment
- Website deployment
- Background worker deployment
- Scheduled job deployment
- Service discovery
- Ingress
- TLS
- Scaling
- Rollout
- Rollback
- Health checks
- Resource limits
- Runtime configuration

Key objects:

```text
Application
Service
API
Job
Worker
Ingress
Deployment
Release
```

Reference technologies:

- Kubernetes / K3s
- Traefik
- cert-manager
- containerd

---

## 10. Cloud & Infrastructure

Purpose:

> Support one platform foundation across many clouds, VPS, edge, on-prem, and sovereign environments.

Capabilities:

- Infrastructure provisioning
- Network provisioning
- Storage provisioning
- Cluster provisioning
- Node management
- Environment management
- Multi-cloud support
- Edge deployment
- On-prem deployment
- Air-gapped deployment
- Infrastructure drift detection
- Infrastructure state management

Key objects:

```text
CloudAccount
Cluster
Node
Network
Storage
Region
Environment
```

Reference technologies:

- OpenTofu
- Kubernetes
- K3s
- Terraform-compatible providers

---

## 11. Deployment & GitOps

Purpose:

> Make desired state declarative, versioned, reviewed, and continuously reconciled.

Capabilities:

- GitOps repository management
- Environment promotion
- Deployment pipeline
- Config management
- Secret references
- Rollback
- Drift detection
- Pull request deployment workflows
- Change review
- Release tracking
- Deployment audit

Key objects:

```text
Repository
Environment
Release
Deployment
ChangeRequest
DesiredState
ActualState
Drift
```

Reference technologies:

- Flux
- ArgoCD
- Forgejo
- GitHub Actions / Forgejo Actions

---

## 12. OCI Artifact Management

Purpose:

> Package, store, distribute, verify, and govern deployable artifacts.

Capabilities:

- Container registry
- OCI artifact registry
- Helm chart registry
- WASM artifact support
- Agent artifact support
- SBOM storage
- Signature verification
- Vulnerability scanning
- Artifact promotion
- Artifact retention
- Artifact provenance

Key objects:

```text
Artifact
Image
Chart
SBOM
Signature
Provenance
Version
Release
```

Reference technologies:

- Harbor
- OCI
- Cosign-ready signing
- SBOM-ready metadata

---

## 13. Observability & Operations

Purpose:

> Keep systems visible, measurable, supportable, and reliable.

Capabilities:

- Metrics
- Logs
- Traces
- Dashboards
- Alerts
- SLOs
- SLAs
- Incident tracking
- Runbooks
- Health checks
- Backup monitoring
- Uptime monitoring
- Agent run observability
- Model usage observability
- Workflow observability

Key objects:

```text
Metric
Log
Trace
Alert
Dashboard
Incident
Runbook
SLO
SLA
```

Reference technologies:

- OpenTelemetry
- Prometheus
- Grafana
- Loki / Tempo-ready

---

## 14. Security & Secrets

Purpose:

> Protect credentials, workloads, data, agents, identities, and operations.

Capabilities:

- Secret storage
- Secret rotation
- Secret injection
- Access policy
- Vulnerability scanning
- Security posture review
- Network policy
- TLS management
- Artifact signing
- Supply-chain security
- Audit logs
- Incident response
- Security runbooks

Key objects:

```text
Secret
Credential
Token
Certificate
Vulnerability
SecurityPolicy
Incident
```

Reference technologies:

- Infisical
- cert-manager
- OPA
- Harbor scanning
- Kubernetes RBAC

---

## 15. Cost & Usage Management

Purpose:

> Make usage visible, measurable, explainable, and billable.

Capabilities:

- Usage metering
- Compute metering
- Storage metering
- Bandwidth metering
- Agent run metering
- Workflow execution metering
- Observability retention metering
- Cost allocation
- Showback
- Chargeback
- Budgeting
- Alerts
- Cost reports
- Optimization recommendations

Key objects:

```text
UsageRecord
Meter
CostCenter
Budget
InvoiceLine
Report
```

Reference technologies:

- OpenCost
- Kubernetes metrics
- Custom usage meters

---

## 16. Billing & Pricing

Purpose:

> Support no-upfront-credit-card onboarding, usage-based pricing, managed operations fees, and support tiers.

Capabilities:

- Free start
- Usage-based billing
- Base platform access
- Managed operations pricing
- Support tiers
- Governance tiers
- Invoicing
- Plan management
- Subscription management
- Credits
- Discounts
- Enterprise contracts

Key objects:

```text
Plan
Subscription
Invoice
UsageLine
Credit
SupportTier
GovernanceTier
```

---

## 17. Support & Service Management

Purpose:

> Provide accountable human support and managed operations around the platform.

Capabilities:

- Support tickets
- Incident response
- Customer communication
- Escalation
- SLA tracking
- Support tier management
- Monthly reviews
- Quarterly reviews
- Knowledge base
- Runbook access
- Change requests
- Service reports

Key objects:

```text
Ticket
Incident
Escalation
SLA
Review
Runbook
ServiceReport
```

---

## 18. Channel & Integration

Purpose:

> Allow users, developers, operators, partners, and agents to interact through many channels.

Capabilities:

- Web console
- Admin console
- Developer portal
- CLI
- API
- SDK
- Git channel
- Chat channel
- Email channel
- WhatsApp channel
- Slack / Teams integration
- Webhooks
- Event API
- MCP tools
- A2A endpoints
- Partner portal

Key objects:

```text
Channel
Integration
Connector
Webhook
EventSubscription
APIKey
Tool
```

---

## 19. Industry Solution Composition

Purpose:

> Compose reusable platform blocks into sector-ready solutions.

Capabilities:

- Industry templates
- Solution blueprints
- Reusable modules
- Dashboard templates
- Workflow templates
- Agent templates
- Integration templates
- Compliance templates
- Pricing templates
- Deployment templates

Industry packages:

- Retail Cloud
- Restaurant Cloud
- Clinic Cloud
- Learning Cloud
- Services Cloud
- Realty Cloud
- Operations Cloud
- Logistics Cloud
- Community Cloud
- Developer Cloud
- Enterprise Cloud
- Sovereign Cloud

Key objects:

```text
Solution
Blueprint
Module
Template
Package
Industry
```

---

## 20. Marketplace & Partner Ecosystem

Purpose:

> Enable partners, developers, and ecosystem participants to distribute, compose, and operate reusable solutions.

Capabilities:

- Partner onboarding
- Partner profiles
- Solution listings
- Template listings
- Integration catalog
- Agent catalog
- Tool catalog
- Certification
- Review and approval
- Revenue sharing
- Support routing

Key objects:

```text
Partner
Listing
CatalogItem
Certification
Approval
RevenueShare
```

---

## 21. Compliance & Audit

Purpose:

> Provide evidence for security, operational, data, agent, and business governance.

Capabilities:

- Audit log collection
- Audit evidence export
- Access review reports
- Policy decision logs
- Agent action logs
- Data access logs
- Change history
- Backup verification evidence
- Incident reports
- Compliance mapping
- Control catalog
- Evidence retention

Key objects:

```text
AuditLog
Evidence
Control
ComplianceReport
AccessReview
PolicyDecision
```

---

## 22. Learning & Improvement Loop

Purpose:

> Continuously improve the platform through evidence, usage, incidents, feedback, and business outcomes.

Capabilities:

- Feedback collection
- Usage analysis
- Incident learning
- Cost optimization
- Agent evaluation
- Workflow improvement
- Customer health scoring
- Product telemetry
- Improvement backlog
- Roadmap planning
- Experiment tracking

Loop:

```text
Reality
  ↓
Observe
  ↓
Learn
  ↓
Decide
  ↓
Act
  ↓
Reality
```

Key objects:

```text
Feedback
Insight
Experiment
Improvement
BacklogItem
RoadmapItem
```

---

## Capability Map Summary

| Domain | Anchor |
|---|---|
| Tenant & Workspace | Tenant isolation |
| Identity & Access | Actor identity |
| Governance & Policy | Policy decision evidence |
| Agent Runtime | Governed agent action |
| Model Management | Approved model usage |
| Memory & Knowledge | Evidence-backed memory |
| Data & Graph | Reality graph |
| Workflow & Automation | Reliable execution |
| Application Runtime | Running workload |
| Cloud & Infrastructure | Compute reality |
| Deployment & GitOps | Desired state |
| OCI Artifacts | Versioned artifact |
| Observability & Operations | Operational evidence |
| Security & Secrets | Protected credentials |
| Cost & Usage | Metered consumption |
| Billing & Pricing | Usage-based value capture |
| Support & Service | Accountable support |
| Channel & Integration | User and agent interface |
| Industry Composition | Reusable solution blocks |
| Marketplace & Partner | Ecosystem distribution |
| Compliance & Audit | Evidence record |
| Learning Loop | Continuous adaptation |

---

## Final Statement

> The Unboxd Cloud capability model turns the constitution into executable platform domains. Each capability must remain reality-anchored, governed, observable, composable, and usable across tenants, clouds, models, frameworks, delivery modes, channels, industries, and governance needs.
