# Meters & Units of Measure

Every billable quantity is a **meter**. A meter key encodes its dimension and
unit so usage, pricing, and invoices are unambiguous. Convention:
`<domain>.<resource>.<unit>`.

| Meter key | Unit of measure | Definition | Source |
|-----------|-----------------|------------|--------|
| `compute.vcpu.hour` | vCPU·hour | one virtual CPU allocated for one hour | OpenCost/Prometheus |
| `compute.mem.gb.hour` | GiB·hour | one gibibyte of memory for one hour | OpenCost/Prometheus |
| `storage.gb.month` | GiB·month | one gibibyte stored for one month (730 h) | OpenCost |
| `network.egress.gb` | GiB | one gibibyte of egress data transferred | Prometheus |
| `metrics.series.hour` | series·hour | one active time series retained for one hour | Prometheus |
| `messaging.msg.million` | 10⁶ messages | one million messages produced/consumed | NATS exporter |
| `ai.cpu.hour` | CPU·hour | one CPU core-hour of (open-source) LLM inference | KServe/Ollama |
| `ai.gpu.hour` | GPU·hour | one GPU allocated for one hour | KServe |
| `ai.tokens.million` | 10⁶ tokens | one million tokens (prompt + completion) | inference gateway |
| `build.minute` | minute | one minute of build execution | Tekton/Buildpacks |
| `function.invocation.million` | 10⁶ invocations | one million function (Lambda) invocations | Knative/OpenFaaS |
| `function.gb.second` | GB·second | memory-time of function execution (Lambda model) | Knative/OpenFaaS |
| `token.issued.million` | 10⁶ tokens | one million STS security tokens issued | Dex/SPIRE |
| `s3.request.million` | 10⁶ requests | one million S3-compatible API requests | Rook/Ceph RGW |
| `agent.run.hour` | run·hour | one agent runtime instance-hour | Dapr Agents |
| `email.sent.thousand` | 10³ emails | one thousand emails sent (SES model) | Postal/Haraka |

## Rules
- **Counters** (egress, messages, tokens, build minutes) accumulate over the
  billing period.
- **Rate-over-time** meters (vCPU·hour, mem GiB·hour, GPU/CPU·hour,
  series·hour) are integrated as quantity × duration.
- **Allocation** meters (storage GiB·month) prorate by time held.
- Binary units (GiB = 1024³ bytes) are used for memory/storage/egress;
  decimal multipliers (million = 10⁶) for counts.
- Quantities are non-negative floats; aggregation is additive per meter per
  tenant per period (see ADR-0003).

Adding a meter = add its key here, price it in the price book dataset, and emit
`UsageEvent`s with that key. No code change.
