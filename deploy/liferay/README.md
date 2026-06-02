# Liferay on k3s

Liferay (Digital Experience Platform) as the composable-experience / portal
layer, deployed on k3s. The manifest is dev-grade (embedded HSQL DB on a PVC).

## Deploy

```sh
kubectl apply -f deploy/liferay/liferay.yaml
kubectl -n liferay rollout status deploy/liferay --timeout=600s   # ~2-4 min first boot
```

## Access

Via the k3s Traefik ingress (add the host to `/etc/hosts` pointing at a node IP):

```sh
echo "$(kubectl get node -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}') liferay.local" | sudo tee -a /etc/hosts
# open http://liferay.local
```

Or port-forward:

```sh
kubectl -n liferay port-forward svc/liferay 8080:8080
# open http://localhost:8080
```

## Production notes

- Use an external database (Postgres/MySQL) instead of embedded HSQL.
- Pin a specific Liferay GA image tag (e.g. `liferay/portal:7.4.3.132-ga132`).
- Liferay is memory-heavy: size requests/limits and JVM `-Xmx` accordingly.
- Front with TLS (cert-manager) and scale replicas behind a shared DB + clustering.
