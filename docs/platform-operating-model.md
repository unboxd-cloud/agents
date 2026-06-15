# Unboxd Cloud Platform Operating Model

## Purpose

This document defines how Unboxd Cloud operates: how stakeholders interact, how solutions move from request to production, how governance works, how evidence is created, and how the platform improves over time.

Related artifacts:

- [Platform Constitution](./platform-constitution.md)
- [Platform Principles](./platform-principles.md)
- [Business Model](./business-model.md)
- [Platform Capability Model](./platform-capability-model.md)
- [Platform Reference Architecture](./platform-reference-architecture.md)
- [Platform Service Catalog](./platform-service-catalog.md)

---

## Operating Model Summary

```text
Assess → Compose → Deploy → Operate → Govern → Improve
```

This is the core operating loop for every customer, tenant, solution, workload, agent, and managed service.

---

## Operating Principles

1. Reality is the anchor.
2. Every action must create or update evidence.
3. Every customer context must be mapped to stakeholders, outcomes, risks, and constraints.
4. Every solution must be composed from reusable platform services where possible.
5. Every deployment must be observable, supportable, and recoverable.
6. Every agent must operate under identity, policy, trust, and audit boundaries.
7. Every managed service must have an owner, scope, support tier, and operating cadence.
8. Every improvement must be grounded in usage, feedback, incidents, cost, risk, or customer outcomes.

---

## Stakeholder Roles

| Role | Responsibility |
|---|---|
| Customer Sponsor | Owns business outcome and commercial relationship |
| Business Owner | Defines sector or business-specific requirements |
| IT Admin | Owns access, users, systems, data, and operational controls |
| Developer | Builds or maintains apps, integrations, automations, and agents |
| Security Owner | Reviews risk, access, secrets, compliance, and audit evidence |
| Platform Operator | Operates infrastructure, deployments, backups, observability, and incidents |
| Agent Owner | Owns agent purpose, permissions, tools, memory, and evaluation expectations |
| Support Owner | Owns support relationship, tickets, escalation, and service reporting |
| Unboxd Architect | Designs solution composition and reference architecture |
| Unboxd Operator | Runs managed operations and platform support |
| Unboxd Governance Owner | Reviews policies, approvals, risks, and audit requirements |
| Partner | Implements or supports approved solutions under platform standards |

---

## Customer Journey

```text
Lead
  ↓
Discovery
  ↓
Assessment
  ↓
Solution Design
  ↓
Proposal
  ↓
Onboarding
  ↓
Deployment
  ↓
Go-Live
  ↓
Managed Operations
  ↓
Review & Improvement
  ↓
Expansion
```

### 1. Lead

Goal:

> Understand whether the customer fits a packaged solution, platform service, or enterprise engagement.

Evidence created:

- Lead source
- Segment
- Use case
- Initial stakeholder
- Basic requirements

---

### 2. Discovery

Goal:

> Understand business goals, pain points, constraints, and urgency.

Inputs:

- Business problem
- Current tools
- Current infrastructure
- Current data sources
- Current workflows
- Budget expectation
- Timeline

Outputs:

- Discovery notes
- Stakeholder map
- Initial solution fit
- Risk assumptions

---

### 3. Assessment

Goal:

> Map current reality before proposing the target solution.

Assessment areas:

- Infrastructure
- Applications
- Data
- Identity
- Security
- Backups
- Monitoring
- Cost
- AI / agent maturity
- Compliance needs

Outputs:

- Current-state map
- Gap analysis
- Risk profile
- Recommended delivery mode

---

### 4. Solution Design

Goal:

> Compose the right platform services into a customer-ready solution.

Design outputs:

- Solution architecture
- Selected service catalog items
- Deployment mode
- Tenancy model
- Support tier
- Governance requirements
- Pricing estimate
- Success metrics

---

### 5. Proposal

Goal:

> Confirm scope, responsibilities, pricing model, support model, and timeline.

Proposal sections:

- Outcome
- Scope
- Exclusions
- Delivery model
- Operating model
- Pricing model
- Support tier
- Governance expectations
- Acceptance criteria

---

### 6. Onboarding

Goal:

> Prepare tenant, access, billing, support, and deployment context.

Onboarding checklist:

- Tenant created
- Workspace created
- Stakeholders assigned
- Access model configured
- Support channel configured
- Billing model configured
- Deployment environment selected
- Governance baseline applied

---

### 7. Deployment

Goal:

> Install, configure, test, and verify the solution.

Deployment steps:

- Provision infrastructure
- Deploy platform services
- Configure domain and TLS
- Configure identity and access
- Configure secrets
- Configure app / agent / workflow
- Configure monitoring
- Configure backups
- Run verification checks

Evidence created:

- Deployment record
- Version record
- Configuration record
- Health check result
- Backup check result
- Access validation

---

### 8. Go-Live

Goal:

> Move the solution into active use with support and governance in place.

Go-live checklist:

- Customer acceptance
- Monitoring active
- Backup active
- Support path active
- Admin handover complete
- Runbook available
- Rollback path known
- Initial usage meter active

---

### 9. Managed Operations

Goal:

> Keep the system reliable, secure, observable, cost-aware, and supportable.

Operations activities:

- Monitoring
- Alert response
- Backups
- Patch support
- Upgrades
- Incident response
- Cost tracking
- Access support
- Change support
- Monthly review

---

### 10. Review & Improvement

Goal:

> Learn from reality and improve continuously.

Review inputs:

- Usage
- Cost
- Incidents
- Feedback
- Support tickets
- Agent evaluations
- Workflow performance
- Security findings
- Business outcome metrics

Outputs:

- Improvement backlog
- Optimization plan
- Expansion recommendation
- Risk remediation plan

---

## Solution Lifecycle

```text
Idea
  ↓
Blueprint
  ↓
Composition
  ↓
Deployment
  ↓
Operation
  ↓
Governance
  ↓
Improvement
  ↓
Versioned Solution
```

### Solution States

| State | Meaning |
|---|---|
| Draft | Idea or early concept |
| Blueprint | Defined reusable solution pattern |
| Composed | Assembled for a customer or segment |
| Deployed | Installed in an environment |
| Live | In active customer use |
| Managed | Operated under support scope |
| Governed | Under policy, audit, risk, and access controls |
| Deprecated | Replaced or no longer recommended |
| Archived | Retained for history and evidence |

---

## Agent Lifecycle

```text
Proposed
  ↓
Designed
  ↓
Approved
  ↓
Provisioned
  ↓
Active
  ↓
Monitored
  ↓
Evaluated
  ↓
Improved / Suspended / Retired
```

### Agent controls

Every agent must have:

- Owner
- Purpose
- Identity
- Model policy
- Tool policy
- Memory policy
- Data access policy
- Evaluation criteria
- Escalation path
- Kill switch
- Audit trail

### Agent states

| State | Meaning |
|---|---|
| Proposed | Agent idea exists |
| Designed | Purpose, tools, model, memory, and policies defined |
| Approved | Governance approval granted |
| Provisioned | Agent identity and runtime configured |
| Active | Agent can perform allowed actions |
| Probation | Agent requires increased oversight |
| Suspended | Agent cannot act |
| Retired | Agent is no longer used |
| Archived | Agent history retained for audit |

---

## Change Management Flow

```text
Change Request
  ↓
Impact Review
  ↓
Policy Check
  ↓
Approval
  ↓
Implementation
  ↓
Verification
  ↓
Evidence Record
```

Change categories:

- Infrastructure change
- Application change
- Agent change
- Model change
- Tool change
- Policy change
- Data change
- Access change
- Pricing change
- Support scope change

Required evidence:

- Who requested the change
- Why the change was needed
- What was changed
- Who approved it
- What policy applied
- What risk was accepted
- What verification passed
- What rollback path exists

---

## Governance Flow

```text
Actor
  ↓
Intent
  ↓
Policy Evaluation
  ↓
Trust / Risk Evaluation
  ↓
Approval if required
  ↓
Action
  ↓
Audit Evidence
```

Governance decision types:

- Allow
- Deny
- Require approval
- Require more evidence
- Escalate
- Suspend
- Revoke

---

## Incident Flow

```text
Signal
  ↓
Alert
  ↓
Triage
  ↓
Incident
  ↓
Response
  ↓
Resolution
  ↓
Postmortem
  ↓
Improvement
```

Incident evidence:

- Trigger
- Time
- Affected tenant
- Affected service
- Severity
- Owner
- Timeline
- Actions taken
- Customer communication
- Root cause
- Corrective actions

---

## Support Flow

```text
Request
  ↓
Ticket
  ↓
Classification
  ↓
Assignment
  ↓
Resolution
  ↓
Customer Confirmation
  ↓
Knowledge Update
```

Support categories:

- Access
- Billing
- Deployment
- Incident
- Configuration
- Agent behavior
- Workflow issue
- Data issue
- Security issue
- Change request

---

## Operating Cadence

### Daily

- Monitor platform health
- Review critical alerts
- Review failed backups or jobs
- Review high-priority support tickets
- Review critical security signals

### Weekly

- Review incidents
- Review support trends
- Review deployment changes
- Review cost anomalies
- Review agent failures or risky actions
- Update improvement backlog

### Monthly

- Customer service review
- Usage and cost review
- Backup verification review
- Security posture review
- Access review where applicable
- Roadmap and improvement review

### Quarterly

- Architecture review
- Governance review
- Support tier review
- Business outcome review
- Expansion or optimization planning
- Compliance evidence review where applicable

---

## Evidence Model

Every important process must create evidence.

Evidence types:

- Discovery evidence
- Assessment evidence
- Deployment evidence
- Change evidence
- Policy decision evidence
- Access evidence
- Agent action evidence
- Tool execution evidence
- Model usage evidence
- Incident evidence
- Backup evidence
- Cost evidence
- Support evidence
- Compliance evidence

Evidence must include where possible:

- Actor
- Time
- Source
- Action
- Target
- Policy
- Result
- Confidence
- Link to logs, traces, metrics, or records

---

## Ownership Model

Every platform object should have an owner.

Objects requiring ownership:

- Tenant
- Workspace
- Environment
- Application
- Agent
- Tool
- Workflow
- Dataset
- Model route
- Policy
- Secret
- Integration
- Dashboard
- Backup plan
- Support contract

Ownership types:

- Business owner
- Technical owner
- Security owner
- Operations owner
- Governance owner
- Support owner

---

## Decision Rights

| Decision | Primary owner | Required reviewers |
|---|---|---|
| Business scope | Customer sponsor | Unboxd architect |
| Solution design | Unboxd architect | Customer IT / business owner |
| Production deployment | Platform operator | Security owner where required |
| Agent activation | Agent owner | Governance owner |
| High-risk tool access | Governance owner | Security owner, business owner |
| Model approval | AI owner | Governance owner, security owner |
| Access policy | IT admin | Security owner |
| Incident severity | Platform operator | Support owner |
| Pricing exception | Commercial owner | Business owner |
| Compliance exception | Governance owner | Security owner |

---

## Real vs Surreal Operating Rule

The operating model must preserve the distinction between actual reality and possible reality.

Actual reality:

- Incidents happened
- Deployments happened
- Costs were incurred
- Agents acted
- Users accessed data
- Backups completed or failed

Possible reality:

- Plans
- Forecasts
- Simulations
- Recommendations
- Proposed changes
- Desired states

Rule:

> Possibilities may guide action, but evidence must confirm reality.

---

## Operating Metrics

### Reliability

- Uptime
- Incident count
- Mean time to detect
- Mean time to respond
- Mean time to recover
- Backup success rate
- Restore success rate

### Usage

- Active tenants
- Active users
- Active apps
- Agent runs
- Workflow executions
- Storage used
- Compute used
- Bandwidth used

### Governance

- Policy decisions
- Approval requests
- Access review completion
- Audit evidence completeness
- Agent actions reviewed
- Risk exceptions

### Support

- Ticket volume
- Response time
- Resolution time
- Escalations
- Customer satisfaction

### Business

- Monthly recurring revenue
- Usage revenue
- Managed operations revenue
- Support revenue
- Customer retention
- Expansion revenue

---

## Operating Model Summary

```text
Stakeholder
  ↓
Intent
  ↓
Assessment
  ↓
Solution Composition
  ↓
Deployment
  ↓
Operation
  ↓
Governance
  ↓
Evidence
  ↓
Improvement
```

---

## Final Statement

> The Unboxd Cloud operating model turns platform principles into repeatable action. Every customer journey, solution, deployment, agent, change, incident, and support interaction must produce evidence, remain governed, and feed the continuous improvement loop.
