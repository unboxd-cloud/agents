# ADR-0003: Pay-as-you-go rating with one tiered engine

## Status
Accepted

## Context
Pricing needs flat per-unit rates, free allowances, and graduated tiers. We do
not want one code path per pricing style (that duplicates logic and drifts).

## Decision
- Usage is captured as immutable `metering.UsageEvent`s (tenant, meter,
  quantity, timestamp).
- Pricing lives in a **versioned `PriceBook`**: each meter has an optional free
  allowance and an ordered list of graduated tiers (`upTo`, `unitPrice`).
- A single `Rater` aggregates usage per meter per period and walks the tiers
  once. Flat pricing is just "one tier with no upper bound"; free allowance is a
  leading zero-priced tier. One engine, every pricing shape.

## Consequences
- Adding a pricing model is data in a `PriceBook`, not new code.
- Price changes are versioned and auditable; historical invoices stay
  reproducible.
- The rater is pure (usage + pricebook → invoice), making it easy to test.
