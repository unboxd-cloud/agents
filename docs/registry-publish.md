# Docker Hub: one-click publish, search & deploy

All artifacts are **OCI**, so Docker Hub is just one registry among many — no
special-casing.

## One-click publish
```bash
make publish PUBLISH_REGISTRY=docker.io/<youruser> PUBLISH_TAG=0.1.0
```
Builds and pushes every service image to Docker Hub in one command. The same
target works for GHCR, ECR, GCR, etc. by changing `PUBLISH_REGISTRY`.

## Search & deploy from Docker Hub
Because catalog offerings reference images by OCI coordinates, you can source
any service from Docker Hub:
- **Search:** an offering's image is `docker.io/<ns>/<name>:<tag>`; discovery uses
  the Docker Hub registry API (or `docker search`).
- **Deploy:** point Helm at the published images:
  ```bash
  helm install platform deploy/helm/platform \
    --set image.registry=docker.io --set image.repository=<youruser> \
    --set image.tag=0.1.0
  ```
- A catalog/marketplace offering can carry a Docker Hub image reference so a
  one-click deploy pulls straight from Docker Hub onto the active provider
  (k8s/k3s/edge/cloud) — the same path as any publishing route
  (`docs/publishing-routes.md`).

## Provenance
Publishing attaches SBOM + SLSA provenance and signs images (see
`docs/versioning.md`), so search results can be verified before deploy.
