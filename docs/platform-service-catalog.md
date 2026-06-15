# Unboxd Cloud Platform Service Catalog

## Purpose

This document defines the Unboxd Cloud service catalog: the platform services, managed services, solution services, and governance services offered through the Unboxd Cloud platform.

Related artifacts:

- [Platform Constitution](./platform-constitution.md)
- [Platform Principles](./platform-principles.md)
- [Business Model](./business-model.md)
- [Platform Capability Model](./platform-capability-model.md)
- [Platform Reference Architecture](./platform-reference-architecture.md)

---

## Catalog Model

```text
Capabilities
  ↓
Platform Services
  ↓
Managed Services
  ↓
Solution Packages
  ↓
Customer Outcomes
```

The service catalog answers:

> What does Unboxd Cloud offer, operate, govern, and support?

---

## Service Categories

```text
1. Cloud Foundation Services
2. Developer Platform Services
3. Agent-Native Services
4. Data, Memory & Knowledge Services
5. Identity, Policy & Governance Services
6. Observability & Operations Services
7. Security & Compliance Services
8. Cost, Usage & Billing Services
9. Multi-Cloud & Deployment Services
10. Industry Solution Services
11. Support & Managed Operations Services
12. Partner & Marketplace Services
```

---

## 1. Cloud Foundation Services

### 1.1 Kubernetes Foundation Service

Purpose:

> Provide a production-ready Kubernetes or K3s foundation for apps, agents, workflows, and platform services.

Includes:

- K3s / Kubernetes installation
- Node setup
- Namespace strategy
- Storage class setup
- Ingress setup
- TLS readiness
- RBAC baseline
- Resource governance
- Health checks
- Upgrade path

Reference technologies:

- K3s
- Kubernetes
- containerd
- Traefik
- cert-manager

Anchor:

```text
Runtime anchor → Kubernetes / K3s
```

---

### 1.2 Ingress & TLS Service

Purpose:

> Expose services securely with routing, domains, and certificates.

Includes:

- Ingress configuration
- Domain routing
- TLS certificates
- Certificate renewal
- HTTP to HTTPS redirect
- Route governance

Reference technologies:

- Traefik
- cert-manager
- Kubernetes Ingress / Gateway-ready patterns

Anchor:

```text
Access anchor → secure entrypoint
```

---

### 1.3 Cloud Foundation Hardening Service

Purpose:

> Harden the platform foundation for production usage.

Includes:

- OS update baseline
- Firewall review
- Kubernetes RBAC review
- Secret handling review
- Backup readiness
- Monitoring readiness
- Network exposure review
- Runbook creation

Anchor:

```text
Trust anchor → secure baseline
```

---

## 2. Developer Platform Services

### 2.1 Code Forge Service

Purpose:

> Provide a code collaboration and source control foundation.

Includes:

- Git hosting
- Organization and repository setup
- Branching model
- Access controls
- Pull request workflow
- Webhooks
- CI/CD integration readiness

Reference technologies:

- Forgejo
- GitHub where required

Anchor:

```text
Code anchor → Git repository
```

---

### 2.2 OCI Registry Service

Purpose:

> Store and distribute containers, Helm charts, WASM modules, SBOMs, and future agent artifacts.

Includes:

- Container registry
- Helm chart registry
- Artifact metadata
- Vulnerability scanning readiness
- Retention policy
- Promotion workflow
- Access controls

Reference technologies:

- Harbor
- OCI
- Cosign-ready signing
- SBOM-ready metadata

Anchor:

```text
Artifact anchor → OCI registry
```

---

### 2.3 GitOps Service

Purpose:

> Make platform state declarative, versioned, reviewed, and reconciled.

Includes:

- GitOps repository setup
- Environment folders
- Deployment manifests
- Helm / Kustomize setup
- Reconciliation
- Drift detection
- Rollback workflow
- Promotion workflow

Reference technologies:

- Flux
- ArgoCD
- Forgejo / GitHub
- Helm
- Kustomize

Anchor:

```text
Desired state anchor → Git
```

---

### 2.4 CI/CD Service

Purpose:

> Automate build, test, package, scan, and deployment workflows.

Includes:

- Pipeline setup
- Build automation
- Test automation
- Image build
- Registry push
- Deployment trigger
- Environment promotion
- Basic security checks

Reference technologies:

- Forgejo Actions
- GitHub Actions
- Buildpacks where appropriate
- OCI registry

Anchor:

```text
Delivery anchor → pipeline evidence
```

---

## 3. Agent-Native Services

### 3.1 Agent Runtime Service

Purpose:

> Run agents as governed actors that can observe, reason, use tools, execute workflows, and escalate to humans.

Includes:

- Agent lifecycle
- Agent identity
- Tool bindings
- Model bindings
- Memory bindings
- Policy bindings
- Run history
- Traceability
- Suspension / termination controls

Reference technologies:

- MCP
- A2A
- OpenAPI
- CloudEvents
- Kubernetes workloads

Anchor:

```text
Agent anchor → governed actor identity
```

---

### 3.2 Agent Tool Registry Service

Purpose:

> Register, govern, expose, and audit tools used by humans and agents.

Includes:

- Tool catalog
- Tool schemas
- Tool permissions
- Tool approval
- Tool usage logs
- Tool health checks
- MCP endpoint mapping

Reference technologies:

- MCP
- OpenAPI
- Webhooks
- OpenFGA / OPA

Anchor:

```text
Tool anchor → governed capability contract
```

---

### 3.3 Model Router Service

Purpose:

> Route requests to the right model based on task, policy, cost, privacy, latency, and quality.

Includes:

- Provider abstraction
- Model registry
- Model routing
- Fallback
- Usage tracking
- Cost tracking
- Quality tracking
- Policy enforcement

Model classes:

- Cloud models
- Open-source models
- Local models
- Embedding models
- Vision models
- Speech models
- Reranking models

Anchor:

```text
Model anchor → approved model route
```

---

### 3.4 Agent Evaluation Service

Purpose:

> Evaluate agents, outputs, workflows, model decisions, and operational quality.

Includes:

- Agent run evaluation
- Output evaluation
- Tool-use evaluation
- Human feedback
- Regression checks
- Trust score inputs
- Evaluation reports

Anchor:

```text
Reliability anchor → evaluation evidence
```

---

### 3.5 Human Approval Service

Purpose:

> Ensure high-risk actions are approved by accountable humans.

Includes:

- Approval requests
- Approval workflows
- Escalation
- Decision evidence
- Approval audit trail
- Policy-based approval rules

Anchor:

```text
Accountability anchor → human decision record
```

---

## 4. Data, Memory & Knowledge Services

### 4.1 Graph & Reality Model Service

Purpose:

> Represent business reality as actors, assets, relationships, actions, events, memory, and time.

Includes:

- Entity graph
- Relationship graph
- Event graph
- Temporal records
- Schema validation
- JSON-LD metadata
- Reality classification

Reference technologies:

- SurrealDB
- JSON-LD
- schema.org
- CloudEvents

Anchor:

```text
Reality anchor → evidence-backed graph
```

---

### 4.2 Memory Service

Purpose:

> Preserve human, agent, organization, project, and system memory.

Includes:

- Agent memory
- Organization memory
- Project memory
- Conversation memory where permitted
- Memory versioning
- Memory governance
- Memory retrieval
- Memory deletion / retention policy

Anchor:

```text
Continuity anchor → governed memory
```

---

### 4.3 Knowledge Base Service

Purpose:

> Store, search, govern, and retrieve documents, sources, evidence, and knowledge.

Includes:

- Document ingestion
- Source attribution
- Full-text search
- Semantic search
- Knowledge graph
- Evidence tracking
- Access policy
- Retention policy

Reference technologies:

- SurrealDB
- Vector search where required
- JSON-LD / schema.org

Anchor:

```text
Knowledge anchor → source-backed evidence
```

---

### 4.4 Real vs Surreal Classification Service

Purpose:

> Keep facts, claims, predictions, simulations, goals, and recommendations clearly separated.

Includes:

- Fact classification
- Claim classification
- Inference classification
- Prediction classification
- Simulation classification
- Goal classification
- Evidence linking
- Confidence scoring

Anchor:

```text
Truth anchor → typed reality claim
```

---

## 5. Identity, Policy & Governance Services

### 5.1 Identity Fabric Service

Purpose:

> Provide portable identity across humans, agents, organizations, services, tools, and systems.

Includes:

- Human identity
- Agent identity
- Service identity
- Tool identity
- Organization identity
- SSO readiness
- SCIM readiness
- DID / VC readiness
- Credential lifecycle

Anchor:

```text
Identity anchor → addressable actor
```

---

### 5.2 Relationship Authorization Service

Purpose:

> Govern access based on relationships between actors, resources, tenants, teams, and roles.

Includes:

- ReBAC
- RBAC
- ABAC-ready patterns
- Relationship tuples
- Permission checks
- Access review
- Delegation

Reference technologies:

- OpenFGA

Anchor:

```text
Authorization anchor → relationship decision
```

---

### 5.3 Policy-as-Code Service

Purpose:

> Evaluate policy decisions across infrastructure, applications, agents, tools, workflows, and governance.

Includes:

- Policy definition
- Policy testing
- Policy simulation
- Policy enforcement
- Policy audit
- Exception handling

Reference technologies:

- OPA
- AuthZEN-ready patterns

Anchor:

```text
Policy anchor → explainable decision
```

---

### 5.4 Trust & Risk Service

Purpose:

> Calculate and continuously update trust and risk for actors, agents, tools, data, workflows, and actions.

Includes:

- Trust scoring
- Risk scoring
- Provenance checks
- Reputation history
- Drift detection
- Trust-based approvals
- Agent kill switch inputs

Anchor:

```text
Trust anchor → evidence + history + policy + time
```

---

### 5.5 Audit & Evidence Service

Purpose:

> Preserve evidence for governance, compliance, operations, and accountability.

Includes:

- Audit logs
- Policy decision logs
- Access logs
- Agent action logs
- Change logs
- Evidence export
- Retention policy

Anchor:

```text
Audit anchor → immutable evidence trail
```

---

## 6. Observability & Operations Services

### 6.1 Telemetry Service

Purpose:

> Collect metrics, logs, traces, and events from platform components, workloads, agents, models, and workflows.

Includes:

- Metrics collection
- Logs collection
- Traces collection
- Event collection
- Dashboards
- Alerting
- Retention policy

Reference technologies:

- OpenTelemetry
- Prometheus
- Grafana
- Loki / Tempo-ready

Anchor:

```text
Operational anchor → telemetry evidence
```

---

### 6.2 Incident Management Service

Purpose:

> Track, respond to, resolve, and learn from incidents.

Includes:

- Incident records
- Alert routing
- Escalation
- Status updates
- Postmortems
- Corrective actions
- Runbook links

Anchor:

```text
Reliability anchor → incident evidence
```

---

### 6.3 Backup & Recovery Service

Purpose:

> Protect data, configuration, artifacts, and operational state.

Includes:

- Backup schedules
- Backup verification
- Restore testing
- Recovery runbooks
- Retention policy
- Evidence logs

Anchor:

```text
Recovery anchor → verified restore path
```

---

### 6.4 Managed Operations Service

Purpose:

> Operate customer systems continuously under defined scope and support tier.

Includes:

- Monitoring
- Patching
- Upgrades
- Backup checks
- Incident support
- Cost review
- Monthly service review
- Improvement backlog

Anchor:

```text
Service anchor → accountable operations
```

---

## 7. Security & Compliance Services

### 7.1 Secrets Management Service

Purpose:

> Store, rotate, inject, and govern secrets safely.

Includes:

- Secret storage
- Secret access policy
- Secret injection
- Rotation workflow
- Secret audit
- Environment-specific secrets

Reference technologies:

- Infisical
- Kubernetes secrets integration

Anchor:

```text
Secret anchor → governed credential
```

---

### 7.2 Security Posture Service

Purpose:

> Assess platform, workload, dependency, artifact, and configuration security.

Includes:

- Vulnerability review
- Configuration review
- Registry scanning readiness
- Policy review
- Exposure review
- Supply chain review

Anchor:

```text
Security anchor → risk evidence
```

---

### 7.3 Compliance Evidence Service

Purpose:

> Provide evidence packages for customer governance and compliance needs.

Includes:

- Audit exports
- Access review reports
- Backup verification reports
- Incident reports
- Policy decision reports
- Change records
- Control mapping

Anchor:

```text
Compliance anchor → evidence package
```

---

## 8. Cost, Usage & Billing Services

### 8.1 Usage Metering Service

Purpose:

> Measure usage for infrastructure, applications, agents, workflows, support, and governance.

Includes:

- Compute usage
- Storage usage
- Bandwidth usage
- Backup usage
- App usage
- Environment usage
- Agent run usage
- Workflow execution usage
- Observability retention usage
- Support hour usage

Anchor:

```text
Usage anchor → metered consumption
```

---

### 8.2 Cost Visibility Service

Purpose:

> Make costs visible, explainable, and optimizable.

Includes:

- Cost dashboards
- Budget alerts
- Showback
- Chargeback-ready reports
- Optimization suggestions

Reference technologies:

- OpenCost
- Kubernetes metrics

Anchor:

```text
Cost anchor → measurable spend
```

---

### 8.3 Billing Service

Purpose:

> Support usage-based pricing, base access, managed operations, support tiers, and enterprise contracts.

Includes:

- Plans
- Subscriptions
- Invoices
- Usage lines
- Credits
- Discounts
- Support tiers
- Governance tiers

Customer promise:

> Start free. No upfront credit card. Scale when ready. Pay only for what you use.

Anchor:

```text
Revenue anchor → usage-backed invoice
```

---

## 9. Multi-Cloud & Deployment Services

### 9.1 Infrastructure Provisioning Service

Purpose:

> Provision infrastructure across VPS, cloud, on-prem, edge, and sovereign environments.

Includes:

- Infrastructure plans
- Environment provisioning
- Cluster provisioning
- Storage provisioning
- Network provisioning
- Drift detection
- State management

Reference technologies:

- OpenTofu
- Kubernetes providers
- Cloud providers

Anchor:

```text
Infrastructure anchor → desired state + actual resources
```

---

### 9.2 Dedicated Tenant Service

Purpose:

> Provide stronger isolation for growing, regulated, or enterprise customers.

Includes:

- Dedicated namespace or cluster
- Dedicated configuration
- Dedicated data layer where needed
- Separate policies
- Stronger support boundaries

Anchor:

```text
Tenant anchor → isolated customer context
```

---

### 9.3 Private Cloud / On-Prem Service

Purpose:

> Deploy and operate Unboxd Cloud foundation inside customer-owned environments.

Includes:

- Customer environment assessment
- Private deployment architecture
- Installation
- Governance integration
- Support model
- Upgrade model

Anchor:

```text
Sovereignty anchor → customer-controlled environment
```

---

### 9.4 Edge / Air-Gapped Service

Purpose:

> Run selected platform capabilities near operations or without public internet dependency.

Includes:

- Edge K3s cluster
- Local registry option
- Local observability
- Local backup plan
- Controlled sync
- Offline artifact distribution

Anchor:

```text
Resilience anchor → local operational continuity
```

---

## 10. Industry Solution Services

### 10.1 Retail Cloud

For:

- Shops
- Boutiques
- D2C sellers
- Local stores

Includes:

- Catalog
- Orders
- Inventory
- Billing
- Customer records
- Offers
- Notifications
- Sales dashboard
- Support agent

Outcome:

> Sell, manage stock, track orders, and retain customers.

---

### 10.2 Restaurant Cloud

For:

- Cafes
- Restaurants
- Bakeries
- Cloud kitchens

Includes:

- Digital menu
- Online orders
- Table booking
- Kitchen view
- Delivery flow
- Review tracking
- Daily sales summary
- Customer notifications

Outcome:

> Manage menu, orders, bookings, delivery, and reviews.

---

### 10.3 Clinic Cloud

For:

- Doctors
- Clinics
- Diagnostics
- Wellness centers

Includes:

- Appointment booking
- Patient records
- Prescription notes
- Follow-up reminders
- Reports
- Access control
- Audit trail
- Knowledge agent

Outcome:

> Manage appointments, follow-ups, records, and patient communication.

---

### 10.4 Learning Cloud

For:

- Coaching centers
- Trainers
- Schools
- Online educators

Includes:

- Course pages
- Student registration
- Schedules
- Assignments
- Payments
- Certificates
- Progress tracking
- Notifications

Outcome:

> Run courses, track students, collect payments, and issue certificates.

---

### 10.5 Services Cloud

For:

- Consultants
- Agencies
- Accountants
- Lawyers
- Freelancers

Includes:

- Lead capture
- Client portal
- Proposal workflow
- Documents
- Tasks
- Invoices
- Knowledge base
- Meeting notes

Outcome:

> Manage leads, clients, documents, tasks, and billing.

---

### 10.6 Realty Cloud

For:

- Brokers
- Property consultants
- Real estate agencies

Includes:

- Property listings
- Lead CRM
- Site visit tracking
- Buyer and seller records
- Document workflow
- Follow-up reminders
- Listing analytics

Outcome:

> Manage listings, leads, visits, documents, and follow-ups.

---

### 10.7 Operations Cloud

For:

- Workshops
- Small manufacturers
- Repair businesses
- Fabrication shops

Includes:

- Job cards
- Work orders
- Inventory
- Supplier records
- Machine and asset records
- Staff tasks
- Daily reports
- Cost tracking

Outcome:

> Track work, inventory, suppliers, jobs, assets, and costs.

---

### 10.8 Logistics Cloud

For:

- Local delivery businesses
- Fleet operators
- Dispatch teams
- Field service teams

Includes:

- Dispatch board
- Driver records
- Job assignment
- Delivery status
- Proof of delivery
- Customer updates
- Incident logs
- Fleet records

Outcome:

> Assign jobs, track deliveries, update customers, and manage field operations.

---

### 10.9 Community Cloud

For:

- Nonprofits
- Associations
- Societies
- Local communities

Includes:

- Website
- Member database
- Donations
- Events
- Volunteer coordination
- Newsletter
- Document archive
- Reports

Outcome:

> Manage members, donors, events, volunteers, communication, and reporting.

---

### 10.10 Developer Cloud

For:

- Startups
- Agencies
- Indie builders
- SaaS teams

Includes:

- Git
- CI/CD
- OCI registry
- K3s runtime
- Secrets
- Monitoring
- Backups
- Cost visibility
- Deployment support

Outcome:

> Build, deploy, monitor, and operate apps without hiring a full platform team.

---

## 11. Support & Managed Operations Services

### 11.1 Starter Support

Best for:

- Small businesses
- Early users
- Low-risk workloads

Includes:

- Basic support
- Documentation
- Standard response
- Basic monitoring
- Backup baseline

---

### 11.2 Managed Support

Best for:

- Growing businesses
- Production workloads
- Developer teams

Includes:

- Monitoring
- Incident support
- Patch support
- Backup checks
- Monthly review
- Cost review
- Change support

---

### 11.3 Governed Support

Best for:

- Regulated businesses
- Enterprise pilots
- Security-sensitive workloads

Includes:

- SLA / SLO tracking
- Access reviews
- Audit reports
- Security reviews
- Governance reviews
- Change approval support
- Incident postmortems

---

### 11.4 Enterprise Support

Best for:

- Enterprise customers
- Private cloud
- Multi-cloud
- On-prem
- Air-gapped

Includes:

- Dedicated support model
- Architecture reviews
- Upgrade planning
- Security reviews
- Compliance evidence
- Dedicated escalation
- Co-managed operations

---

## 12. Partner & Marketplace Services

### 12.1 Partner Onboarding Service

Purpose:

> Enable agencies, MSPs, system integrators, developers, and consultants to deliver on Unboxd Cloud foundation.

Includes:

- Partner profile
- Solution approval
- Training
- Support routing
- Delivery standards
- Certification readiness

---

### 12.2 Solution Marketplace Service

Purpose:

> Publish reusable solutions, templates, agents, tools, integrations, and blueprints.

Includes:

- Solution listings
- Agent listings
- Tool listings
- Integration listings
- Template listings
- Review and approval
- Governance metadata
- Usage tracking

---

### 12.3 Certification Service

Purpose:

> Certify solutions, integrations, agents, tools, and partners against Unboxd Cloud standards.

Includes:

- Technical review
- Security review
- Governance review
- Documentation review
- Operational readiness review
- Certification badge

Anchor:

```text
Ecosystem anchor → certified reusable capability
```

---

## Service Catalog Summary

| Service Category | Customer Outcome | Anchor |
|---|---|---|
| Cloud Foundation | Reliable runtime | Kubernetes / K3s |
| Developer Platform | Faster build and deploy | Git + OCI |
| Agent-Native | Governed agent operation | Agent identity |
| Data, Memory & Knowledge | Evidence-backed intelligence | Graph + memory |
| Identity, Policy & Governance | Controlled action | Identity + policy |
| Observability & Operations | Operational reliability | Telemetry |
| Security & Compliance | Reduced risk | Evidence + controls |
| Cost, Usage & Billing | Pay for actual use | Metered usage |
| Multi-Cloud & Deployment | Portability | Desired state |
| Industry Solutions | Business outcomes | Composable blocks |
| Support & Managed Operations | Accountable service | Support contract |
| Partner & Marketplace | Ecosystem scale | Certified listings |

---

## Final Statement

> The Unboxd Cloud service catalog turns platform capabilities into customer-facing services, managed operations, governance controls, and industry-ready solution packages. Every service must remain reality-anchored, observable, governed, composable, and usage-aware.
