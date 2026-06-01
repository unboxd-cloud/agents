# Frontend & Admin Control Panel

## Framework choice: htmx
The admin panel uses **htmx** (~14kb) over the server's stdlib `html/template`.

Why htmx (vs React/Vue/Svelte):
- **Lightweight & framework-agnostic:** no Node toolchain, no build step, no SPA
  bundle — matches the platform's stdlib-only, "lightweight" tenets.
- **Server-rendered:** HTML fragments come from the same Go services; one
  language, no API duplication for a separate FE.
- **Progressive:** plain HTML works; htmx adds partial updates (chat, live trace
  panel) with attributes only.

The UI is **optional and separable** — all services are headless/API-first; the
panel only consumes the SDK.

## What the panel provides
- **Chat interface:** drive the control plane in natural commands (`catalog`,
  `rate`, `frameworks`, …); each command runs through the SDK.
- **Flow traces (APM):** every chat action is traced span-by-span and shown in an
  auto-refreshing panel — debuggable to a granular level.
- **OTLP export:** `/debug/traces.json` for analysis in Jaeger/Tempo/Prometheus.
- **Embedded BI:** links to Grafana/Jaeger/Superset (via env) for insights.
- **Overview:** providers, operating models, billing meters, catalog,
  compliance frameworks.

## Run it
```bash
./bin/admin   # :8080 ; set CATALOG_URL/BILLING_URL/COMPLIANCE_URL/METERING_URL
```
Env: `GRAFANA_URL`, `JAEGER_URL`, `SUPERSET_URL`, `OTEL_COLLECTOR_URL` to embed
BI links.
