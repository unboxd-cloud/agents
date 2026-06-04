# Unboxd Cloud Platform Principles

## Purpose

This document defines the operating principles behind Unboxd Cloud.

The business model explains how value is created and captured. The platform principles explain what must remain true as the platform grows.

---

## Core Principle

> Everything needs an anchor.

For Unboxd Cloud, the anchor is reality: running systems, observable evidence, customer outcomes, operational health, identity, policy, usage, and trust.

The platform must not drift into abstraction. Every model, agent, workflow, deployment, and solution must remain connected to measurable reality.

---

## Principle 1: Reality Anchored

Unboxd Cloud must always reconcile against reality.

Reality means:

- Running infrastructure
- Actual usage
- Observable metrics
- Logs and traces
- Customer outcomes
- Incidents
- Costs
- Access decisions
- Audit evidence

Every claim must be grounded in evidence.

```text
Model → Evidence → Reality
```

Not:

```text
Model → Model → Model
```

---

## Principle 2: Open Source First

Unboxd Cloud prefers open-source primitives for infrastructure, runtime, governance, observability, and agent-native systems.

Open source provides:

- Transparency
- Portability
- Auditability
- Community validation
- Lower vendor lock-in
- Long-term customer control

Commercial value comes from composition, operation, integration, governance, support, and reliability.

---

## Principle 3: Agent-Native by Design

Agent-native does not mean adding chatbots to infrastructure.

Agent-native means the platform treats humans, agents, tools, workflows, models, data, and systems as governed actors in the same operational environment.

Agent-native systems require:

- Identity
- Memory
- Policy
- Tools
- Permissions
- Audit trails
- Evaluation
- Human approval where needed
- Trust and risk controls

Agents may reason in possibility space, but they must act in reality space.

---

## Principle 4: Kubernetes-Mature

Every production solution should be Kubernetes-mature, not merely Kubernetes-compatible.

Kubernetes-mature means:

- Containerized workloads
- Declarative configuration
- Health checks
- Resource requests and limits
- Ingress and TLS
- Externalized secrets
- Persistent storage where required
- Backup and restore plan
- Metrics, logs, and traces
- RBAC
- Network policies where needed
- GitOps compatibility
- Upgrade and rollback path
- Runbooks

---

## Principle 5: CNCF-Aligned and Cloud-Native Mature

Unboxd Cloud should prefer CNCF graduated and incubating projects for the production foundation.

Cloud-native maturity means the platform is:

- Scalable
- Resilient
- Observable
- Secure
- Automated
- Portable
- Cost-visible
- Recoverable
- Upgradeable

Sandbox, experimental, or non-CNCF tools may be used only when they fill a clear gap and are wrapped with governance.

---

## Principle 6: Multi-Tenant by Default, Dedicated When Needed

Unboxd Cloud should support multiple tenancy models:

- Shared multi-tenant
- Dedicated tenant
- Single-tenant private cloud
- On-prem
- Edge
- Air-gapped

Every tenant must have isolation for:

- Identity
- Data
- Secrets
- Agents
- Applications
- Logs
- Billing
- Policies
- Backups
- Support context

---

## Principle 7: Multi-Model

The platform must support the right model for the right task.

Model choice should be governed by:

- Cost
- Accuracy
- Latency
- Privacy
- Compliance
- Data sensitivity
- Task complexity
- Customer preference

Supported model classes include:

- Cloud AI models
- Open-source models
- Local models
- Domain-specific models
- Vision models
- Speech models
- Embedding models
- Reranking models
- Small task models
- High-accuracy reasoning models

---

## Principle 8: Multi-Cloud and Portable

The platform must not assume a single cloud provider.

Supported targets include:

- VPS
- Bare metal
- K3s / Kubernetes
- AWS
- Azure
- GCP
- OCI
- Hetzner
- OVH
- On-prem
- Edge
- Air-gapped environments

Cloud is an execution environment, not the platform itself.

---

## Principle 9: Multi-Framework

The platform should support the frameworks customers already use.

Unboxd Cloud should support multiple classes of frameworks:

- Web frameworks
- Backend frameworks
- Agent frameworks
- Workflow frameworks
- Data frameworks
- Integration frameworks
- UI frameworks

The platform should not force one application framework, one agent framework, one workflow engine, or one data model.

---

## Principle 10: Multi-Delivery

Customers should consume the platform in the way that fits their maturity and constraints.

Delivery modes include:

- Self-serve
- Managed
- Dedicated
- Private cloud
- On-prem
- Edge
- Air-gapped
- Partner-led
- Enterprise co-managed

One foundation should support many delivery realities.

---

## Principle 11: Multi-Channel

Users, developers, operators, partners, and agents should interact with the platform through the right channel.

Channels include:

- Web
- Mobile
- CLI
- API
- SDK
- Git
- Chat
- Email
- WhatsApp
- Slack
- Teams
- Voice
- Marketplace
- Partner portal
- MCP
- A2A
- Webhooks
- Events

Channels are interfaces. Governance must remain consistent across all of them.

---

## Principle 12: Multi-Governance

Governance must be built into the platform, not added afterward.

Governance domains include:

- Identity governance
- Access governance
- Data governance
- AI and agent governance
- Security governance
- Cost governance
- Operational governance
- Compliance governance
- Open-source governance

The platform must answer:

- Who acted?
- What changed?
- Why was it allowed?
- What evidence exists?
- What policy applied?
- What was the cost?
- What risk was introduced?

---

## Principle 13: Multi-Industry and Composable

Unboxd Cloud should not build one rigid product per industry.

It should build reusable components that compose into industry solutions.

Composable blocks include:

- Identity
- Data
- Workflow
- Agent
- Policy
- Integration
- Channel
- Dashboard
- Automation
- Billing
- Observability

Industry solutions are compositions of reusable blocks.

---

## Principle 14: Usage-Based and Low-Risk to Start

Customer-facing promise:

> Start free. No upfront credit card. Scale when ready. Pay only for what you use.

Usage-based pricing should apply to:

- Compute
- Storage
- Bandwidth
- Backups
- Databases
- Apps
- Environments
- Users
- Agent runs
- Workflow executions
- Observability retention
- Support hours
- Governance reports

Usage-based pricing must remain tied to operational reality.

---

## Principle 15: Enterprise-Ready Underneath

Even when serving small businesses, the foundation should be enterprise-ready underneath.

Enterprise-ready means:

- Secure
- Governed
- Auditable
- Observable
- Supportable
- Upgradeable
- Recoverable
- Cost-visible
- Policy-controlled
- Identity-aware

Small-business accessible does not mean fragile.

---

## Principle 16: Human and Agent Collaboration

Humans and agents are both actors in the operational graph.

The platform should model:

- Human to human collaboration
- Human to agent delegation
- Agent to human escalation
- Agent to agent coordination
- Agent to tool execution
- Agent to data access

Agents do not replace accountability. They must operate within identity, policy, trust, evidence, and human approval boundaries.

---

## Principle 17: Real vs Surreal Separation

The platform must distinguish between what is real and what is possible.

Real includes:

- Facts
- Observations
- Events
- Evidence
- Current state
- History

Surreal includes:

- Plans
- Predictions
- Simulations
- Goals
- Hypotheses
- Recommendations
- Desired states

Never store reality, interpretation, prediction, and simulation as the same thing.

Agents may explore unrealized edges, but actions must create evidence in reality.

---

## Principle 18: Continuous Learning Loop

The platform must learn in a loop.

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

Learning is not accumulation. Learning is adaptation.

The loop must remain connected to evidence.

---

## Principle 19: One Platform, Many Forms

The platform should support many business realities without fragmenting into many disconnected products.

```text
One platform.
Many tenants.
Many models.
Many clouds.
Many frameworks.
Many delivery modes.
Many channels.
Many governance needs.
Many industries.
Composable by design.
```

---

## Principle 20: Keep the Anchor Visible

Every artifact, solution, agent, workflow, deployment, and business claim should be able to answer:

- What is the anchor?
- What evidence supports it?
- What reality does it affect?
- Who owns it?
- Who governs it?
- What happens if it fails?

If the anchor is unclear, the system is drifting.

---

## Final Statement

> Unboxd Cloud exists to turn open-source, cloud-native, and agent-native primitives into reality-anchored, governed, composable, production-ready solutions.
