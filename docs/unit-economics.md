# Unit Economics

Unit economics = revenue and cost measured **per billable unit (meter)**. The
platform makes both sides observable: **revenue** from the billing engine and
**cost** from OpenCost, joined on the meter and `TenantID`.

## Definitions

| Term | Definition | Source |
|------|------------|--------|
| Unit | one meter unit (e.g. 1 vCPU-hour, 1M tokens, 1 build-minute) | `metering` |
| Unit price | tier-resolved price for the unit | `billing.PriceBook` |
| Unit cost (COGS) | underlying infra cost for the unit | OpenCost |
| Unit margin | `unit price − unit cost` | derived |
| Contribution margin % | `(price − cost) / price` | derived |
| Effective price | price after free allowance over a period | `billing.Rate` |
| Blended margin | margin after partner settlement + tax-exclusive | `billing.Settle` |

## Worked example (per meter)

| Meter | Unit price | ~Unit cost | Unit margin | Margin % |
|-------|-----------:|-----------:|------------:|---------:|
| compute.vcpu.hour (tier 1) | 0.040 | 0.022 | 0.018 | 45% |
| ai.cpu.hour (open-source CPU LLM) | 0.080 | 0.045 | 0.035 | 44% |
| ai.gpu.hour (tier 1) | 2.500 | 1.700 | 0.800 | 32% |
| ai.tokens.million | 0.600 | 0.300 | 0.300 | 50% |
| build.minute | 0.008 | 0.004 | 0.004 | 50% |

(Costs are illustrative; real costs come from OpenCost per cluster/provider.)

## Why CPU-based LLMs matter
Defaulting AI inference to **open-source, CPU-based LLMs** (llama.cpp/Ollama)
keeps unit cost low and predictable versus scarce GPUs — better contribution
margin and broader runnability (incl. edge/k3s). GPU stays available as a
higher-priced tier for latency-sensitive workloads.

## Effect of operating models
- **Direct:** platform keeps full unit margin.
- **Reseller / service-provider:** partner adds markup on top — platform keeps
  base margin; partner keeps the markup.
- **Agency / marketplace:** commission is taken from base — platform margin =
  base margin − commission.
- **Marketplace publishing:** `revShare` splits revenue with the publisher; the
  platform's margin is the retained share minus unit cost.

## Making it observable
- Revenue per meter/tenant: `billing` invoices/line items.
- Cost per meter/tenant: OpenCost (a `metering.Source`).
- Health/throughput: `/metrics` (Prometheus) on every service.
- Join revenue ↔ cost in Grafana/Superset (see `docs/observability.md`) for live
  per-unit margin dashboards.
