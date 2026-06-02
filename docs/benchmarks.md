# Benchmarking report

## Status — read this first

- **Measured here:** engine micro-benchmarks of *our own code* (`go test -bench`).
  These measure the platform's internal speed only. Reproduce with `make bench`.
- **NOT yet run:** any comparison against another agent platform. No external
  platform, competitor agent, LLM, or agent-bench task suite has been executed in
  this repository. The cross-platform "scorecard" further down is a **design
  claim list, not a benchmark result** — it is explicitly unverified until
  agent-bench is run live (see Methodology).

## Engine micro-benchmarks (measured)

Environment: Go 1.24, linux/amd64, GOMAXPROCS=4. Machine-dependent; rerun with
`make bench`. These benchmark the ADL runtime and the agentdb store in isolation
— not end-to-end agent task performance, and not a comparison to anything.

ADL runtime (`pkg/adl`), compiling `testdata/sample.agent` (16 declarations):

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `Parse` (lex + parse) | ~36,900 | 38,488 | 89 |
| `Compile` (parse + validate) | ~34,600 | 38,592 | 96 |
| `LoadAgent` (compile + project) | ~40,600 | 39,584 | 116 |

agentdb governed store (`pkg/agentdb`):

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `MemStorePutGet` | ~325 | 13 | 1 |
| `KernelProposeApply` | ~11,600 | 2,818 | 5 |
| `KernelGoverned` (under a policy gate) | ~5,100 | 2,064 | 5 |

## Methodology — agent-bench (planned, not yet executed)

To produce a *real* cross-platform comparison, the platform agent runs
**agent-bench**: it instantiates two participants from the **same certified
blueprint** (the `.agent` definition) — ours and another platform's — runs the
suite's tasks against both, and emits a JSON-LD (schema.org) report. Model:
`metamodels/benchmark.agent`.

**This has not been run.** Doing so requires: a certified blueprint, a running
agent-bench task suite, live model endpoints, and an actual competitor
participant — none of which exist in this repo today. Until then there are no
cross-platform scores to report.

## Capability comparison (design claims — UNVERIFIED, not measured)

The following reflects the platform's *intended* design, not measured results. It
is here to define what agent-bench should test, not to assert outcomes.

| Dimension | Our design intent | Notes |
| --- | --- | --- |
| Runtime | Kubernetes-native (k3s) kernel agent | claim |
| Determinism | Deterministic core; LLMs pluggable, out of hot path | claim |
| Governance | propose → govern → approve → apply, audited | claim |
| A2A safety | every A2A call gated by tool + policy | claim |
| Context | complete-context gate before execution | claim |
| Grounding | verified sources + DBpedia/schema.org KG | claim |
| Trust | social, propagative, composable (TrustGNN) | claim |
| Accountability | named, signed-off owner per brief/solution | claim |

## Reproducing

```sh
make bench      # engine micro-benchmarks (the only measured numbers here)
make agents     # validate platform.agent and all metamodels
```
