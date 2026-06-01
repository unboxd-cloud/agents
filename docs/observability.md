# Observability, APM & Embedded BI

Everything the platform does is observable and **publishable** to CNCF tooling.

## Three signals
| Signal | How | Export to |
|--------|-----|-----------|
| **Metrics** | `/metrics` (Prometheus text) on every service | Prometheus → Grafana |
| **Traces (APM)** | `internal/observe` records spans per action | OTLP/JSON → OTel Collector → Jaeger / Grafana Tempo |
| **Usage/cost** | `metering` + OpenCost | Grafana / Apache Superset (BI) |

## Metrics
`server.New` wires `/metrics` automatically, publishing `platform_up`,
`platform_uptime_seconds`, `platform_http_requests_total`,
`platform_http_errors_total`. Scrape with a standard Prometheus
`ServiceMonitor`/annotation; no app changes needed.

## APM / flow traces
The admin chat and SDK actions create spans (operation, attributes, duration,
status) in a ring buffer:
- in-UI: the **Flow Traces** panel (auto-refreshing) shows granular, per-step
  timing for debugging,
- export: `GET /debug/traces.json` returns an **OTLP/JSON** payload ready to POST
  to an OpenTelemetry Collector, which fans out to Jaeger, Grafana Tempo, etc.

This keeps the platform debuggable to a granular level while remaining
vendor-neutral — you analyze in whichever CNCF tool you run.

## Embedded BI / insights
The admin panel embeds links (configurable via `GRAFANA_URL`, `JAEGER_URL`,
`SUPERSET_URL`, `OTEL_COLLECTOR_URL`) so operators reach dashboards in context.
Because traces and metrics export in open formats, **insights live in your BI
tool**, not locked in the platform. Pair revenue (billing) with cost (OpenCost)
for the unit-economics dashboards in `docs/unit-economics.md`.
