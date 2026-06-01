# GitOps

Git is the source of truth; the cluster converges to it. The platform supports
this both as **delivery** (Argo CD/Flux) and as a built-in **agent**.

## Best practices applied
- **Declarative everything:** Helm chart + `deploy/datasets/*.json` fully
  describe desired state; nothing is configured imperatively in-cluster.
- **Separation of code and config:** images are built once (compile-time);
  datasets and values are deploy-time, reconciled from Git.
- **Reconcile, don't push:** an in-cluster agent pulls desired state and
  converges (level-triggered), rather than CI pushing to the cluster.
- **Drift detection:** the GitOps agent re-validates desired state every cycle
  and refuses to apply invalid artifacts.
- **Promotion via Git:** environments differ only by values/datasets overlays in
  Git; promote by merging.
- **No secrets in Git:** use External Secrets Operator; Git holds references.

## GitOps as an agent (built-in)
`internal/gitops` + the `operator` binary run a reconcile loop on the shared
agent runtime:
1. read the desired-state directory (datasets),
2. **validate** each artifact with its loader (catch malformed config early),
3. apply via a pluggable `Applier` (dry-run when none configured).

Run it: `RECONCILE_INTERVAL=30s GITOPS_DIR=deploy/datasets ./bin/operator`.

## GitOps as delivery (Argo CD)
Point an Argo CD `Application` at `deploy/helm/platform`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata: { name: unboxd-platform, namespace: argocd }
spec:
  project: default
  source:
    repoURL: https://github.com/unboxd-cloud/platform
    path: deploy/helm/platform
    targetRevision: main
    helm: { valueFiles: [values.yaml] }
  destination: { server: https://kubernetes.default.svc, namespace: platform }
  syncPolicy: { automated: { prune: true, selfHeal: true } }
```

Both paths reconcile the **same** chart + datasets, so what you tested in the
sandbox is what ships.
