# Category-wise Registries (CNCF landscape as services)

The catalog is the **composable, full-stack registry**: each entry exposes a
[CNCF landscape](https://landscape.cncf.io/) project (or AI-native project) as a
provisionable, metered service, organized by **category**. Pick one per category
to compose a full stack.

## Registry index
`GET /v1/categories` returns the categories; `GET /v1/catalog?category=<c>`
returns that category's registry; `GET /v1/catalog?profile=<p>` filters by
persona.

## Categories → example offerings

| Category | Offerings (seed) | Underlying projects |
|----------|------------------|---------------------|
| `compute` | Managed Kubernetes | vCluster, Cluster API |
| `build` | App Builder | Cloud Native Buildpacks, Tekton |
| `data` | Object Storage | Rook, CSI |
| `messaging` | Managed NATS | NATS |
| `observability` | Managed Prometheus | Prometheus, OpenTelemetry |
| `ai` | Model Inference (CPU LLMs), ML Pipelines, Vector DB, AI Customer Support | KServe, llama.cpp/Ollama, Kubeflow, Milvus, Chatwoot |

These map to broader CNCF landscape categories (orchestration, app definition &
build, observability & analysis, runtime, provisioning, serverless, security &
compliance, database, streaming & messaging). Adding a category or offering is a
**dataset** edit (`deploy/datasets/offerings.json`) — composable, no code change.

## Composability
- Each offering binds: a project (what), a Crossplane composition (how), meters
  (billing), certifications (compliance), and personas (visibility).
- A "stack" is a selection across categories; provisioning fans out to the
  active provider via Crossplane.
- Third parties **publish** offerings with a revenue share (marketplace model,
  see `docs/operating-models.md`).
