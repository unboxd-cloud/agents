# Single-node deployment on k3s (+ Headlamp)

The platform runs on a single-node [k3s](https://k3s.io) cluster — ideal for
edge, homelab, or a small production footprint — and is manageable through the
**k3s API** and the **Headlamp** UI.

## 1. Install k3s (single node)
```bash
curl -sfL https://get.k3s.io | sh -
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
kubectl get nodes
```

## 2. Load images
Use published images, or load locally-built ones into k3s' containerd:
```bash
make images
for s in catalog metering billing compliance admin operator; do \
  docker save localhost/unboxd-cloud/$s:dev | sudo k3s ctr images import -; done
```

## 3. Install via Helm
```bash
helm install platform deploy/helm/platform \
  --set image.registry=localhost --set image.repository=unboxd-cloud --set image.tag=dev \
  --set-file datasets.files.offerings\.json=deploy/datasets/offerings.json \
  --set-file datasets.files.pricebook\.json=deploy/datasets/pricebook.json \
  --set-file datasets.files.tax-rules\.json=deploy/datasets/tax-rules.json \
  --set-file datasets.files.compliance-frameworks\.json=deploy/datasets/compliance-frameworks.json
kubectl get pods
```
The chart's small resource requests and read-only/non-root hardening suit a
single node. `operator` runs the GitOps + orchestrator agents in-cluster.

## 4. Manage with Headlamp
[Headlamp](https://headlamp.dev) (CNCF) is a lightweight Kubernetes UI:
```bash
kubectl apply -f https://raw.githubusercontent.com/headlamp-k8s/headlamp/main/kubernetes-headlamp.yaml
kubectl port-forward -n kube-system service/headlamp 8443:80
```
Browse to the admin control panel too:
```bash
kubectl port-forward svc/admin 8080:8080   # platform admin panel
```

## 5. k3s API
Everything is standard Kubernetes, so the k3s API server, `kubectl`, GitOps
(Argo CD/Flux), and any Kubernetes-native tool work unchanged.
