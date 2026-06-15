# Unboxd Cloud Solutions Model

## Purpose

This document defines the Unboxd Cloud solutions model: how one open-source platform foundation is packaged into business-ready, sector-ready, enterprise-ready, MSP-ready, and marketplace-ready solutions.

Related artifact:

- [Business Model](./business-model.md)

---

## Core Principle

> One open-source platform foundation, many packaged solutions.

Unboxd Cloud does not sell isolated tools. It delivers composed solutions that combine infrastructure, applications, automations, agents, integrations, governance, support, and managed operations.

---

## Solutions Model

```text
Solutions Model
=
Solution Categories
+ Solution Packages
+ Buyer Segments
+ Delivery Modes
+ Revenue Attachment
+ Expansion Path
```

---

## Solution Categories

```text
1. Small Business Solutions
2. Developer / Startup Solutions
3. Mid-Market Platform Solutions
4. Enterprise Governance Solutions
5. Public / Nonprofit / Sovereign Solutions
6. MSP / Partner Delivered Solutions
7. Marketplace Solutions
```

---

## 1. Small Business Solutions

Small businesses buy outcomes, not platform terminology.

Positioning:

> Done-for-you managed cloud solutions for running business operations, websites, automations, and simple agents without hiring a platform team.

Revenue attachment:

```text
setup fee + usage revenue + managed support
```

### Retail Cloud

For:

- Shops
- Boutiques
- Local stores
- D2C sellers

Core outcome:

> Sell, manage stock, track orders, and retain customers.

Components:

- Website / catalog
- Orders
- Inventory
- Billing
- Customer records
- Offers
- Notifications
- Sales dashboard
- Support agent

Expansion path:

```text
Retail Cloud → Marketing automation → Customer support agent → Inventory intelligence → Managed operations
```

---

### Restaurant Cloud

For:

- Cafes
- Restaurants
- Bakeries
- Cloud kitchens

Core outcome:

> Manage menus, orders, bookings, delivery, reviews, and customer updates.

Components:

- Digital menu
- Online ordering
- Table booking
- Kitchen workflow
- Delivery workflow
- Reviews
- Customer notifications
- Daily sales summary

Expansion path:

```text
Restaurant Cloud → Delivery automation → Review management → Customer loyalty → Managed operations
```

---

### Clinic Cloud

For:

- Doctors
- Clinics
- Diagnostics
- Wellness centers

Core outcome:

> Manage appointments, follow-ups, patient records, reminders, and reports.

Components:

- Appointment booking
- Patient records
- Prescription notes
- Follow-up reminders
- Reports
- Access control
- Audit trail
- Knowledge agent

Expansion path:

```text
Clinic Cloud → Patient reminders → Records search → Compliance evidence → Managed support
```

---

### Learning Cloud

For:

- Coaching centers
- Trainers
- Schools
- Online educators

Core outcome:

> Run courses, track students, collect payments, and issue certificates.

Components:

- Course pages
- Student registration
- Schedule
- Assignments
- Payments
- Certificates
- Progress tracking
- Notifications

Expansion path:

```text
Learning Cloud → Course marketplace → Student support agent → Certificate verification → Managed operations
```

---

### Services Cloud

For:

- Consultants
- Agencies
- Accountants
- Lawyers
- Freelancers

Core outcome:

> Manage leads, clients, documents, tasks, proposals, and billing.

Components:

- Lead capture
- Client portal
- Proposal workflow
- Documents
- Tasks
- Invoices
- Knowledge base
- Meeting notes

Expansion path:

```text
Services Cloud → Client portal → Knowledge agent → Billing automation → Managed support
```

---

### Operations Cloud

For:

- Workshops
- Repair businesses
- Small manufacturers
- Fabrication shops

Core outcome:

> Track jobs, inventory, suppliers, assets, tasks, reports, and costs.

Components:

- Job cards
- Work orders
- Inventory
- Supplier records
- Asset records
- Staff tasks
- Daily reports
- Cost tracking

Expansion path:

```text
Operations Cloud → Inventory automation → Work order agent → Cost analytics → Managed operations
```

---

## 2. Developer / Startup Solutions

Startups and developers buy speed, reliability, and lower platform burden.

Positioning:

> Developer cloud and runtime foundation for building, deploying, monitoring, and operating applications and agents without building a platform team first.

Revenue attachment:

```text
usage revenue + managed platform fee + support tier
```

### Developer Cloud

Core outcome:

> Build, deploy, monitor, and operate apps with a production-ready open-source developer platform.

Components:

- Git
- CI/CD
- OCI registry
- Kubernetes / K3s runtime
- Secrets
- Monitoring
- Backups
- Cost visibility
- Deployment support

Expansion path:

```text
Developer Cloud → GitOps Platform → Observability Stack → Agent Runtime → Managed Operations
```

---

### GitOps Platform

Core outcome:

> Make deployments declarative, versioned, reviewed, and recoverable.

Components:

- GitOps repository
- Environment promotion
- Deployment manifests
- Drift detection
- Rollback workflow
- Release history

Expansion path:

```text
GitOps Platform → Policy gates → Multi-environment deployment → Enterprise governance
```

---

### Private Agent Runtime

Core outcome:

> Run agents, tools, workflows, and model calls in a governed private runtime.

Components:

- Agent runtime
- Model routing
- Tool registry
- Workflow execution
- Memory / knowledge base
- Agent traces
- Human approval

Expansion path:

```text
Private Agent Runtime → Agent governance → Enterprise support → Private cloud
```

---

## 3. Mid-Market Platform Solutions

Mid-market customers buy standardization, integration, managed operations, and governance maturity.

Positioning:

> Managed open-source platform and integration foundation for growing companies that need reliability without heavy vendor lock-in.

Revenue attachment:

```text
implementation + monthly managed operations + support + governance add-ons
```

### Managed Open Source Platform

Core outcome:

> Standardize core infrastructure, deployments, monitoring, backups, and support.

Components:

- Kubernetes foundation
- GitOps
- OCI registry
- Secrets
- Monitoring
- Backups
- Cost visibility
- Support process

Expansion path:

```text
Managed Platform → Internal Tools Cloud → Agent Runtime → Governance Stack
```

---

### Internal Tools Cloud

Core outcome:

> Build and operate internal tools, dashboards, workflows, and automations.

Components:

- App hosting
- Database
- Workflow engine
- Dashboards
- Access control
- Notifications
- Support

Expansion path:

```text
Internal Tools Cloud → Workflow automation → Agent assistance → Governance
```

---

### Integration Cloud

Core outcome:

> Connect business systems, APIs, events, webhooks, and workflows.

Components:

- API integrations
- Webhooks
- Event flows
- Workflow triggers
- Data sync
- Monitoring
- Error handling

Expansion path:

```text
Integration Cloud → Automation → Agent tools → Audit and governance
```

---

## 4. Enterprise Governance Solutions

Enterprises buy control, auditability, reliability, support, and risk reduction.

Positioning:

> Governed agent-native infrastructure for regulated, multi-cloud, private, and enterprise environments.

Revenue attachment:

```text
enterprise contract + support + governance + managed operations
```

### Governed Agent-Native Infrastructure

Core outcome:

> Run agent-native systems with identity, policy, audit, approval, and governance controls.

Components:

- Agent identity
- Tool permissions
- Policy-as-code
- Audit logs
- Approval workflows
- Model governance
- Agent evaluation
- Trust and risk controls

Expansion path:

```text
Governed Agent Infrastructure → Private Agent Cloud → Compliance evidence → Enterprise support
```

---

### Identity + Policy Stack

Core outcome:

> Control who can do what, under which conditions, with what evidence.

Components:

- SSO / SCIM readiness
- RBAC / ABAC / ReBAC
- OpenFGA-style relationship authorization
- OPA-style policy evaluation
- Access reviews
- Policy decision logs

Expansion path:

```text
Identity + Policy Stack → Agent governance → Compliance reports → Enterprise contract
```

---

### Observability + Audit Stack

Core outcome:

> Make workloads, agents, workflows, cost, incidents, and governance evidence visible.

Components:

- Metrics
- Logs
- Traces
- Dashboards
- Alerts
- Audit records
- Usage reports
- Incident evidence

Expansion path:

```text
Observability Stack → SLO/SLA → Governance evidence → Managed operations
```

---

### Private / Dedicated Cloud

Core outcome:

> Give customers stronger isolation, control, and deployment ownership.

Components:

- Dedicated tenant
- Dedicated cluster where needed
- Private deployment
- Customer-owned environment option
- Governance integration
- Support model

Expansion path:

```text
Dedicated Cloud → Private Cloud → Multi-cloud → Co-managed operations
```

---

## 5. Public / Nonprofit / Sovereign Solutions

Public, nonprofit, and sovereign customers buy data control, transparency, affordability, and trust.

Positioning:

> Sovereign open-source infrastructure for organizations that need control over data, systems, knowledge, and public-interest workflows.

Revenue attachment:

```text
project delivery + support + governance
```

### Sovereign Cloud Foundation

Core outcome:

> Run open-source systems in a controlled, portable, auditable environment.

Components:

- Kubernetes / K3s
- GitOps
- Open-source stack
- Backups
- Monitoring
- Identity
- Governance
- Documentation

Expansion path:

```text
Sovereign Cloud → Knowledge platform → Public data cloud → Governance services
```

---

### Community Cloud

Core outcome:

> Manage members, donors, events, volunteers, documents, and reports.

Components:

- Website
- Member database
- Donations
- Events
- Volunteer coordination
- Newsletter
- Document archive
- Reports

Expansion path:

```text
Community Cloud → Donor intelligence → Reporting automation → Knowledge cloud
```

---

## 6. MSP / Partner Delivered Solutions

MSPs and partners buy repeatable packages they can sell, deploy, operate, and support.

Positioning:

> Partner-ready managed solution packages powered by Unboxd Cloud templates, automation, governance, metering, and escalation support.

Revenue attachment:

```text
implementation margin + managed operations share + support share + usage revenue share
```

### MSP Managed Cloud

Core outcome:

> Allow MSPs to deliver managed open-source cloud solutions to their own customers.

Components:

- Partner tenant
- Customer tenant templates
- Usage metering
- Managed operations playbooks
- Support escalation
- Governance templates
- Billing support

Expansion path:

```text
MSP Managed Cloud → Sector packages → Marketplace listings → Revenue share
```

---

### Partner Solution Package

Core outcome:

> Enable partners to package industry knowledge into reusable managed solutions.

Components:

- Solution blueprint
- Deployment template
- Documentation
- Support model
- Pricing model
- Certification readiness

Expansion path:

```text
Partner Package → Certified package → Marketplace distribution → Managed revenue share
```

---

## 7. Marketplace Solutions

Marketplace solutions create ecosystem supply and scalable distribution.

Positioning:

> Reusable templates, agents, tools, workflows, connectors, and managed packages that customers can discover, deploy, subscribe to, and pay for by usage.

Revenue attachment:

```text
usage revenue share + listing fees + certification fees + managed package fees + partner support share
```

### Marketplace Asset Types

- Templates
- Agents
- Tools
- Workflows
- Connectors
- Industry blueprints
- Managed solution packages
- Governance packs
- Observability packs
- Integration packs

### Marketplace Participants

```text
Unboxd Cloud
→ owns platform, governance, billing, certification

Vendors
→ publish integrations and components

MSPs / partners
→ publish managed solution packages

Developers
→ publish templates, tools, workflows, agents

Customers
→ discover, deploy, subscribe, and pay by usage
```

Expansion path:

```text
Template → Certified asset → Marketplace listing → Usage revenue → Managed package
```

---

## Solution Revenue Attachment Summary

| Solution Type | Primary Revenue |
|---|---|
| Small Business Solutions | setup + usage + managed support |
| Developer / Startup Solutions | usage + managed platform + support |
| Mid-Market Platform Solutions | implementation + operations + support |
| Enterprise Governance Solutions | enterprise contract + governance + support |
| Public / Nonprofit / Sovereign Solutions | project delivery + support + governance |
| MSP / Partner Delivered Solutions | revenue share + support + usage |
| Marketplace Solutions | usage share + listing + certification + managed package fees |

---

## Solution Expansion Flywheel

```text
Packaged Solution
   ↓
Customer Adoption
   ↓
Usage Revenue
   ↓
Managed Operations
   ↓
Governance / Support Expansion
   ↓
Reusable Blueprint
   ↓
Partner / Marketplace Distribution
   ↓
More Packaged Solutions
```

---

## Final Solutions Model Definition

> Unboxd Cloud packages one open-source agent-native cloud foundation into reusable, sector-ready, developer-ready, enterprise-ready, MSP-ready, and marketplace-ready solutions that create usage revenue, managed operations revenue, governance revenue, partner revenue, and enterprise contract expansion.
