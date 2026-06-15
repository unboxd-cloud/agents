# Theoretical foundations

The platform agent rests on four explicit theories. Each is encoded in a
metamodel (validated by `make agents`) and summarized here.

## 1. Theory of completeness

An agent acts only on complete context. A `Brief` is complete iff every
requirement is satisfied, no clarification is open, and every cited source is
verified:

```
completeness(b) = satisfied(b) / required(b)
complete(b)     = (completeness(b) == 1) AND openClarifications(b) == 0
                  AND unverifiedSources(b) == 0
```

Until `complete(b)`, the agent reports "complete context missing" and asks.
Completeness brings accountability: a complete, confirmed brief has a named,
signed-off owner. Model: `metamodels/solution.agent`.

## 2. Control theory

The platform is a control system. The agent is a controller driving observed
state toward a desired state (the approved agent definition), with a
level-triggered feedback loop:

```
Observe -> Compare -> Classify -> Score -> Explain -> Recommend -> Verify
```

- **Setpoint:** the approved `DesiredAgent` definition (a certified blueprint).
- **Feedback:** runtime `Snapshot`s observed continuously.
- **Error:** `Drift` (capability/tool/prompt/memory/knowledge/workflow/policy/
  objective/model/configuration), risk-scored.
- **Actuation:** `Remediation` — reconcile toward setpoint.
- **Stability:** level-triggered reconciliation, self-healing, no single point of
  failure — "as reliable as Kubernetes." Memory uses Ebbinghaus decay
  (`effectiveConfidence = baseConfidence * e^(-lambda * days)`).

Models: `metamodels/control.agent`, `metamodels/memory.agent`.

## 3. Decision-making theory

Decision intelligence is the north star. A decision chooses among options to
maximize expected utility over weighted criteria, grounded in verified evidence,
using a chosen framework (OODA, ReAct, Plan-Execute, MCDA, Bayesian):

```
expectedUtility(o) = SUM_c weight_c * score(o, c)
chosen             = argmax_o expectedUtility(o)
```

Scores are computed with quantitative **models the agent wields as tools, skills,
and knowledge**: mathematical, statistical, probabilistic, optimization, machine
learning, causal, simulation, and operations research. The north-star metric is
the decision-intelligence score:

```
diScore = w_d*decisionQuality + w_e*experience + w_t*trust
```

Model: `metamodels/decision.agent`.

## 4. Trust theory

Trust is social, propagative, and composable (cf. TrustGNN, arXiv:2205.12784):
never one human's opinion. Every interacting agent or human rates a subject;
trust then propagates over the trust graph and composes across paths.

```
localTrust(s) = clamp[0,1]( SUM_i (w_i * r_i) / SUM_i w_i )
trust(a, b)   = clamp[0,1]( localTrust(b) blended with AGG_{p: a~>b} pathTrust(p) )
```

Performance enters as one input, never the end in itself. Model:
`metamodels/trust.agent`.
