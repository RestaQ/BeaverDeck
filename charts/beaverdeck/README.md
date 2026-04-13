# BeaverDeck Helm Chart

`beaverdeck` installs BeaverDeck, a lightweight Kubernetes operations panel for day-2 cluster work.

## What This Chart Deploys

- `Deployment`
- `Service`
- `ServiceAccount`
- cluster-scoped `ClusterRole` and `ClusterRoleBinding`
- optional `PersistentVolumeClaim`
- optional `Ingress` resource

## Install

Create the target namespace with Helm and install the chart.
Recommendation: enable persistence from the start to keep your custom configuration safe.

```bash
helm upgrade --install beaverdeck oci://ghcr.io/arequs/charts/beaverdeck \
  --version 2.0.0 \
  --set persistence.enabled=true \
  --set persistence.size=1Gi \
  --set persistence.storageClass=standard \
  --set clusterName=your-cluster-name
```

## First Start and Common Configuration

BeaverDeck does not use a pre-created admin secret.
On first start, the application writes a bootstrap token to the pod log. Open the UI, enter that token, and set the admin password.
Example:

```bash
kubectl -n beaverdeck logs deployment/beaverdeck
```

### Important Notes

- The chart does not create a `Namespace` object. Use `--namespace` and `--create-namespace` during install if needed.
- The RBAC installed by this chart is cluster-scoped because BeaverDeck needs access to cluster-wide resources such as nodes, PVs, CRDs, storage classes, and metrics endpoints.
- `clusterName` is displayed in the UI header. Set it explicitly to a human-readable cluster name.

### Enable Ingress

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
  host: beaverdeck.example.com
  path: /
  pathType: Prefix
  tls:
    - hosts:
        - beaverdeck.example.com
      secretName: beaverdeck-tls
```

## Values

| Value | Default | Description |
| --- | --- | --- |
| `nameOverride` | `""` | Override for the chart name portion of generated resource names. |
| `fullnameOverride` | `""` | Full override for generated resource names. |
| `namespaceOverride` | `""` | Override for the namespace used by rendered resources. Defaults to the Helm release namespace. |
| `image.repository` | `arequs/beaverdeck` | Container image repository. |
| `image.tag` | `1.2.1` | Container image tag. |
| `image.pullPolicy` | `IfNotPresent` | Image pull policy. |
| `listenAddr` | `:8080` | HTTP listen address passed to the container. |
| `dataDir` | `/data` | Directory used by BeaverDeck to store SQLite data. |
| `clusterName` | `Cluster name not set` | Cluster name shown in the UI header. |
| `managedNamespace` | `""` | Namespace BeaverDeck treats as its managed namespace. If empty, the pod namespace is used. |
| `allowAllNamespaces` | `true` | If `true`, BeaverDeck can operate across all namespaces allowed by Kubernetes RBAC. |
| `podAnnotations` | `{}` | Extra pod annotations. |
| `podLabels` | `{}` | Extra pod labels. |
| `serviceAccount.create` | `true` | Create a dedicated ServiceAccount. |
| `serviceAccount.name` | `""` | Override ServiceAccount name. If empty, the chart fullname is used. |
| `serviceAccount.annotations` | `{}` | Extra ServiceAccount annotations. |
| `rbac.create` | `true` | Create ClusterRole and ClusterRoleBinding for BeaverDeck. |
| `rbac.clusterRoleName` | `""` | Override ClusterRole name. If empty, the chart fullname is used. |
| `persistence.enabled` | `false` | Use a PersistentVolumeClaim instead of `emptyDir`. Strongly recommended for non-demo installations. |
| `persistence.accessModes` | `["ReadWriteOnce"]` | PVC access modes. |
| `persistence.size` | `1Gi` | PVC size request. |
| `persistence.storageClass` | `default` | StorageClass name for the PVC. Replace with the class available in your cluster. |
| `service.type` | `ClusterIP` | Kubernetes Service type. |
| `service.port` | `80` | Service port. |
| `service.targetPort` | `8080` | Container port exposed by BeaverDeck. |
| `service.annotations` | `{}` | Extra Service annotations. |
| `resources` | `{}` | Container resource requests and limits. |
| `nodeSelector` | `{}` | Node selector for the pod. |
| `tolerations` | `[]` | Pod tolerations. |
| `affinity` | `{}` | Pod affinity rules. |
| `livenessProbe.enabled` | `true` | Enable the liveness probe. |
| `livenessProbe.path` | `/healthz` | Liveness probe HTTP path. |
| `livenessProbe.initialDelaySeconds` | `10` | Initial delay before the liveness probe starts. |
| `livenessProbe.periodSeconds` | `20` | Liveness probe period. |
| `readinessProbe.enabled` | `true` | Enable the readiness probe. |
| `readinessProbe.path` | `/healthz` | Readiness probe HTTP path. |
| `readinessProbe.initialDelaySeconds` | `5` | Initial delay before the readiness probe starts. |
| `readinessProbe.periodSeconds` | `10` | Readiness probe period. |
| `ingress.enabled` | `false` | Render a single Ingress resource for BeaverDeck. |
| `ingress.className` | `""` | Ingress class name. |
| `ingress.annotations` | `{}` | Ingress annotations. |
| `ingress.host` | `""` | Ingress host. Leave empty to omit host matching. |
| `ingress.path` | `/` | Ingress path. |
| `ingress.pathType` | `Prefix` | Ingress path type. |
| `ingress.tls` | `[]` | Ingress TLS configuration. |
| `extraEnv` | `[]` | Extra environment variables appended to the BeaverDeck container. |
