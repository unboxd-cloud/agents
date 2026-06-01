# ADR-0005: Compliance & tax as loadable data + one engine

## Status
Accepted

## Context
The platform must support taxes (VAT/GST/sales), law-of-the-land regulations
(GDPR/CCPA), industry frameworks (HIPAA/PCI-DSS/DORA), and security frameworks
(SOC2/ISO-27001/FedRAMP/NIS2). Hard-coding rates, jurisdictions, and control sets
would duplicate logic and rot quickly.

## Decision
- **Tax:** one `ApplyTax` engine over `TaxRule` data (allowing reverse-charge and
  compounding); applied to the customer-facing amount so it composes with rating
  and partner settlement. Rates load at deployment (`tax-rules.json`).
- **Compliance:** a data-free engine (`Evaluate`) over a loadable framework
  `Registry` (`compliance-frameworks.json`). It checks data residency, offering
  certification, and encryption requirements; the framework definitions are
  deployment-time datasets, not code.
- **Enforcement:** decisions are gated by the existing OPA (policy) + OpenFGA
  (ReBAC) seam. Compliance is keyed by `TenantID` (ADR-0002).

## Consequences
- New taxes/regulations are dataset edits, not releases.
- Invoices are reproducible (price-book version pinned) and tax/compliance are
  auditable.
- The engines stay small, pure, and testable; data lineage is explicit.
