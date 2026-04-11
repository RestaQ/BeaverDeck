# BeaverDeck Helm Chart

`beaverdeck` installs BeaverDeck, a lightweight Kubernetes operations panel for day-2 cluster work.

## What This Chart Deploys

- `Deployment`
- `Service`
- `ServiceAccount`
- cluster-scoped `ClusterRole` and `ClusterRoleBinding`
- optional `PersistentVolumeClaim`
- optional `Ingress` resources

## Install

Create the target namespace with Helm and install the chart:

```bash
helm upgrade --install beaverdeck oci://ghcr.io/arequs/charts/beaverdeck \
  --version 1.1.0 \
  --namespace beaverdeck \
  --create-namespace
```

Or install from a local checkout:

```bash
helm upgrade --install beaverdeck ./charts/beaverdeck \
  --namespace beaverdeck \
  --create-namespace
```

## First Start

BeaverDeck does not use a pre-created admin secret.

On first start, the application writes a bootstrap token to the pod log. Open the UI, enter that token, and set the admin password.

Example:

```bash
kubectl -n beaverdeck logs deployment/beaverdeck
```

## Important Notes

- The chart does not create a `Namespace` object. Use `--namespace` and `--create-namespace` during install.
- The RBAC installed by this chart is cluster-scoped because BeaverDeck needs access to cluster-wide resources such as nodes, PVs, CRDs, storage classes, and metrics endpoints.
- `clusterName` is displayed in the UI header. Set it explicitly to a human-readable cluster name.
- If `managedNamespace` is empty, BeaverDeck defaults to the namespace where it is running.
- If `allowAllNamespaces` is `true`, BeaverDeck can operate across all namespaces allowed by Kubernetes RBAC.

## Common Configuration

### Set Image

```bash
helm upgrade --install beaverdeck oci://ghcr.io/arequs/charts/beaverdeck \
  --namespace beaverdeck \
  --create-namespace \
  --set image.repository=arequs/beaverdeck \
  --set image.tag=1.0.0
```

### Persist Data

```bash
helm upgrade --install beaverdeck oci://ghcr.io/arequs/charts/beaverdeck \
  --namespace beaverdeck \
  --create-namespace \
  --set persistence.enabled=true \
  --set persistence.size=5Gi
```

### Restrict to One Namespace

```bash
helm upgrade --install beaverdeck oci://ghcr.io/arequs/charts/beaverdeck \
  --namespace beaverdeck \
  --create-namespace \
  --set managedNamespace=apps \
  --set allowAllNamespaces=false
```

### Enable Ingress

```yaml
ingresses:
  - nameSuffix: public
    enabled: true
    className: nginx
    annotations:
      nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
      nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    hosts:
      - host: beaverdeck.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - hosts:
          - beaverdeck.example.com
        secretName: beaverdeck-tls
```

## Key Values

- `image.repository`
- `image.tag`
- `clusterName`
- `managedNamespace`
- `allowAllNamespaces`
- `persistence.enabled`
- `persistence.size`
- `service.type`
- `ingresses`
- `extraEnv`
