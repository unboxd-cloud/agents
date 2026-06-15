# Sandbox: testing each phase of the deployment path

Validate the platform locally, phase by phase, before any real cluster. The
default container manager is **Podman** (rootless, daemonless); set
`CONTAINER=docker` to use Docker instead.

## Phase 0 — source (no containers)
```bash
make check     # go vet + tests
make build     # static binaries into ./bin
```

## Phase 1 — run binaries directly
```bash
./bin/catalog & ./bin/metering & ./bin/billing & ./bin/compliance &
curl localhost:8083/v1/catalog
curl localhost:8084/v1/frameworks   # empty unless COMPLIANCE_DATASET is set
```
Test dataset loading by pointing env at the repo datasets:
```bash
COMPLIANCE_DATASET=deploy/datasets/compliance-frameworks.json ./bin/compliance &
CATALOG_DATASET=deploy/datasets/offerings.json ./bin/catalog &
```

## Phase 2 — OCI images (Podman)
```bash
make images                 # podman build all four service images
podman images | grep unboxd-cloud
```

## Phase 3 — run the stack as a Pod (Podman)
```bash
make sandbox-up             # podman play kube deploy/sandbox/pod.yaml
make sandbox-smoke          # curl catalog / pricebook / frameworks
make sandbox-down
```
This is the same Pod spec Kubernetes understands, so a green sandbox means the
images and wiring are cluster-ready.

## Phase 4 — Kubernetes + Helm (kind)
```bash
kind create cluster
# load locally-built images into the kind node
for s in catalog metering billing compliance; do \
  kind load docker-image localhost/unboxd-cloud/$s:dev; done
helm install platform deploy/helm/platform \
  --set image.registry=localhost --set image.repository=unboxd-cloud \
  --set image.tag=dev \
  --set-file datasets.files.offerings\.json=deploy/datasets/offerings.json \
  --set-file datasets.files.pricebook\.json=deploy/datasets/pricebook.json \
  --set-file datasets.files.tax-rules\.json=deploy/datasets/tax-rules.json \
  --set-file datasets.files.compliance-frameworks\.json=deploy/datasets/compliance-frameworks.json
kubectl get pods
```
This phase exercises the **deployment-time dataset loading** via the ConfigMap.

## Phase 5 — GitOps
Point Argo CD at `deploy/helm/platform` (see `docs/gitops.md`); the same chart +
datasets reconcile from Git. What you tested in the sandbox is what ships.
