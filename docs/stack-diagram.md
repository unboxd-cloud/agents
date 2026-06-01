# Stack, Artifacts & Runtimes

## Where everything fits

```mermaid
flowchart TB
  subgraph EX[Experience]
    ADMIN[Admin Control Panel\nhtmx + chat + APM]
    CLI[platform CLI]
    SDK[Go SDK]
    BACK[Backstage portal]
  end
  subgraph CP[Control plane - owned, this repo]
    TEN[tenant]:::svc
    CAT[catalog / registries]:::svc
    MET[metering]:::svc
    BIL[billing: rating+tax+settlement]:::svc
    COMP[compliance]:::svc
    OPR[operator agent\nGitOps + k8s orchestrator]:::svc
  end
  subgraph PS[Platform services - composed CNCF projects]
    XP[Crossplane] --- ARGO[Argo CD] --- CAP[Capsule/vCluster]
    OC[OpenCost] --- PROM[Prometheus] --- OTEL[OpenTelemetry]
    DEX[Dex/SPIFFE] --- OPA[OPA] --- FGA[OpenFGA] --- KEDA[KEDA]
  end
  subgraph PR[Providers - vendor-neutral seam]
    K8S[Kubernetes / k3s] --- CS[Apache CloudStack] --- EDGE[Edge: KubeEdge/K3s] --- CLOUD[AWS/GCP/Azure]
  end
  EX --> CP --> PS --> PR
  classDef svc fill:#1c2230,stroke:#2f6df6,color:#fff;
```

## Artifacts (all OCI-standard, transportable)

| Artifact | Produced by | Consumed by | Notes |
|----------|-------------|-------------|-------|
| Service OCI images | `Dockerfile` (static/scratch) | any OCI runtime / k8s | runs on any cloud |
| Helm chart | `deploy/helm/platform` | Helm / Argo CD | vanilla, no cloud assumptions |
| **Datasets** (offerings, pricebook, tax, frameworks) | `deploy/datasets/*.json` | services at deploy via ConfigMap | versioned data, not code |
| Plugins/extensions | `plugins/*` | binaries (blank import) | native-runtime, compile-time |
| Sandbox manifest | `deploy/sandbox/pod.yaml` | podman / k8s | phase-by-phase testing |
| OTLP traces | `internal/observe` | Jaeger/Tempo/Prometheus | exported for BI/insights |
| SBOM / provenance | CI (see `docs/versioning.md`) | supply-chain tooling | lineage |

## Runtimes (users can target any)

| Runtime | How | Where it fits |
|---------|-----|---------------|
| Local binaries | `make build` | Phase 1 dev |
| Podman pod | `make sandbox-up` | Phase 3 sandbox |
| Single-node **k3s** | `helm install` (see `docs/deploy-k3s.md`) | edge / small prod, managed via **Headlamp** + k3s API |
| Multi-node Kubernetes | Helm + Argo CD | full prod |
| Apache CloudStack VMs | `provider=cloudstack` | legacy/VM estates (optional, ADR-0004) |
| Edge (KubeEdge/K3s) | `provider=edge` | edge sites |

All services are **headless / API-first**; the admin panel is an optional,
separable UI over the same APIs the SDK and CLI use.
