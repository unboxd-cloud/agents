# Taxes, Compliance & Policy

The platform adds support for taxes, law-of-the-land regulations, industry
frameworks, and security frameworks — all as **data + policy**, evaluated by
small engines, enforced through OPA/OpenFGA. No framework data is baked into
code; it is loaded as a dataset at deployment time.

## Taxes (`internal/billing/tax.go`)
- `TaxRule` describes a tax in a jurisdiction (VAT/GST/sales tax), with support
  for **reverse-charge** (B2B self-accounting) and **compounding** (e.g. QST on
  GST).
- `ApplyTax(net, currency, rules)` taxes the customer-facing amount — i.e. after
  partner settlement — so it composes with `Rate` and `Settle`.
- Rates load at deployment from `deploy/datasets/tax-rules.json` (`TaxTable`);
  `TaxRulesFor` is a built-in dev fallback.

## Compliance frameworks (`internal/compliance`)
Supported as a loadable registry (`deploy/datasets/compliance-frameworks.json`):

| Category | Examples |
|----------|----------|
| Privacy (law of the land) | GDPR, CCPA |
| Healthcare (industry) | HIPAA |
| Finance (industry) | PCI-DSS, DORA |
| Security | SOC2, ISO-27001, NIS2 |
| Government | FedRAMP |

Add a framework = add a dataset entry. The engine carries no framework data.

## Per-tenant posture & evaluation
- `compliance.Profile` (keyed by `TenantID`): required `Frameworks`,
  `DataResidency` regions, legal `Jurisdiction` (also drives tax).
- `compliance.Placement`: where/how a resource will run + the offering's
  `Certifications`.
- `Evaluate(profile, placement, registry)` returns a `Report` with findings for:
  - **data residency** — placement region ∈ allowed regions,
  - **offering certification** — offering certified for each required framework,
  - **encryption-at-rest** — required when a framework demands it (from the
    loaded spec).

## Enforcement
`Evaluate` produces the decision; **OPA** (policy gate) blocks non-compliant
actions and **OpenFGA** answers fine-grained "who may do what" — the same
`authz` seam used everywhere. Compliance is thus a first-class gate on
provisioning, not an afterthought.

## Composition summary
Tenancy (one `TenantID`) → metered usage → rated invoice → partner settlement →
**tax by jurisdiction**, with every provisioning action gated by **compliance +
policy**. Each step reuses one engine; new regulations/taxes are data.
