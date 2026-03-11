# Edge Metrics Deployment Guide

> **DEPRECATED**: Shell script deployment is deprecated. Use the Helm chart instead.

## Helm Installation (Recommended)

```bash
# Install from OCI registry
helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics \
  --namespace monitoring --create-namespace

# Development environment
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-dev.yaml \
  --namespace monitoring --create-namespace

# Production environment
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-prod.yaml \
  --namespace monitoring --create-namespace

# With site-specific overrides
cp charts/edge-metrics/values-cluster.example.yaml charts/edge-metrics/values-cluster.yaml
# Edit values-cluster.yaml with your device IPs, registry, etc.
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-prod.yaml \
  -f charts/edge-metrics/values-cluster.yaml \
  --namespace monitoring --create-namespace

# Uninstall
helm uninstall edge-metrics -n monitoring
```

### Key Options

```bash
# Stub mode (test without hardware)
helm install edge-metrics charts/edge-metrics \
  --set exporter.profiles.jetson.env.STUB_MODE=true \
  --set exporter.profiles.generic.env.STUB_MODE=true

# Enable ServiceMonitor (requires Prometheus Operator)
helm install edge-metrics charts/edge-metrics \
  --set server.serviceMonitor.enabled=true \
  --set exporter.serviceMonitor.enabled=true

# Disable PVC (use emptyDir)
helm install edge-metrics charts/edge-metrics \
  --set server.persistence.enabled=false
```

### Port Forwarding

```bash
# Backend API
kubectl port-forward -n monitoring svc/edge-metrics-server 8081:8081

# Frontend Dashboard
kubectl port-forward -n monitoring svc/edge-metrics-frontend 3000:3000
```

## Prerequisites

- Kubernetes cluster (1.26+)
- Helm 3.x
- kubectl configured
- Node labels: `device-family=jetson` (Jetson nodes), `device-family=generic` (generic nodes)

---

## Legacy: Shell Script Deployment (Deprecated)

> These scripts will be removed after 2026-06-01.

### Basic Usage

```bash
# Deploy full stack (local images)
./deploy/deploy-all.sh

# With specific version tag
./deploy/deploy-all.sh v1.0.0
```

### Environment Variables

```bash
# Use registry
REGISTRY=daclab ./deploy/deploy-all.sh v1.0.0

# Enable PVC for persistent storage
USE_PVC=true ./deploy/deploy-all.sh v1.0.0

# Mount local config.db file (development only!)
USE_LOCAL_DB=true ./deploy/deploy-all.sh v1.0.0

# Include ServiceMonitor (Prometheus Operator)
DEPLOY_SERVICEMONITOR=true ./deploy/deploy-all.sh v1.0.0

# All options
NAMESPACE=monitoring \
REGISTRY=daclab \
USE_PVC=true \
DEPLOY_SERVICEMONITOR=true \
./deploy/deploy-all.sh v1.0.0
```

### Deployment Order

1. **Backend**
   - Build Docker image
   - Push image (when using registry)
   - Create RBAC, PVC, Deployment, Service
   - Create ServiceMonitor (optional)

2. **Exporter DaemonSet**
   - Jetson DaemonSet (privileged, tegrastats hostPath)
   - Generic DaemonSet (non-privileged, security hardened)
   - Headless Service + ServiceMonitor (5s interval) per profile

3. **Frontend**
   - Build Docker image
   - Push image (when using registry)
   - Create Deployment, Service
   - Inject Backend Service DNS via environment variable

4. **Status Check**
   - All Pod statuses
   - All Service statuses
   - Print access information

## Uninstall (Full Stack)

### Basic Usage

```bash
# Uninstall full stack (confirmation prompt)
./deploy/undeploy-all.sh

# Force delete without confirmation
FORCE=true ./deploy/undeploy-all.sh
```

### Environment Variables

```bash
# Delete PVC (permanent data loss!)
DELETE_PVC=true ./deploy/undeploy-all.sh

# Delete Docker images
DELETE_IMAGE=true ./deploy/undeploy-all.sh

# All options
NAMESPACE=monitoring \
DELETE_PVC=true \
DELETE_IMAGE=true \
FORCE=true \
./deploy/undeploy-all.sh
```

### Deletion Order

1. **Frontend** — Service -> Deployment -> wait for Pod termination
2. **Exporter DaemonSet** — ServiceMonitor -> Service -> DaemonSet (Jetson + Generic)
3. **Backend** — ServiceMonitor -> Service -> Deployment -> PVC (optional) -> RBAC
4. **Status Check** — verify remaining resources, PVC preservation notice

## Environment Variable Reference

### deploy-all.sh

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | Kubernetes namespace |
| `REGISTRY` | (none) | Docker registry address |
| `USE_PVC` | false | Enable backend PVC |
| `USE_LOCAL_DB` | false | Mount local config.db file (development only) |
| `DEPLOY_SERVICEMONITOR` | false | Deploy ServiceMonitor |

### undeploy-all.sh

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | Target namespace |
| `DELETE_PVC` | false | Delete backend PVC (data loss!) |
| `DELETE_IMAGE` | false | Delete Docker images |
| `FORCE` | false | Skip confirmation prompt |

## Per-Component Deployment

### Backend Only

```bash
cd server
./scripts/deploy.sh v1.0.0
```

### Exporter Only

```bash
cd exporter

# All profiles (Jetson + Generic)
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --all

# Jetson only
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --jetson

# Generic only
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --generic
```

### Frontend Only

```bash
cd front
./scripts/deploy.sh v1.0.0
```

See each component's README for detailed options:
- [Server README](../server/README.md)
- [Exporter README](../exporter/README.md)
- [Frontend README](../front/README.md)

## Accessing Services

### Frontend (Dashboard)

- **NodePort**: `http://<NodeIP>:30080`
- **Port Forward**:
  ```bash
  kubectl port-forward -n monitoring svc/edge-metrics-frontend 3000:3000
  # http://localhost:3000
  ```

### Backend (API)

- **NodePort**: `http://<NodeIP>:31716`
- **Port Forward**:
  ```bash
  kubectl port-forward -n monitoring svc/edge-metrics-server 8081:8081
  # http://localhost:8081
  ```

### Exporter (Metrics)

Access directly on each edge node IP (hostNetwork):
```bash
# Metrics
curl http://<edge-node-ip>:9102/metrics

# Health
curl http://<edge-node-ip>:9101/health
```

## Viewing Logs

```bash
# Backend logs
kubectl logs -n monitoring -l app.kubernetes.io/name=server --tail=50 -f

# Exporter logs (Jetson + Generic)
kubectl logs -n monitoring -l app.kubernetes.io/name=exporter --tail=50 -f

# Frontend logs
kubectl logs -n monitoring -l app.kubernetes.io/name=frontend --tail=50 -f
```

## Troubleshooting

### Frontend Cannot Connect to Backend

**Cause**: API URL environment variable not set or Service DNS issue

**Check**:
```bash
# Check frontend Pod environment variables
kubectl get pod -n monitoring -l app.kubernetes.io/name=frontend \
  -o jsonpath='{.items[0].spec.containers[0].env}'

# Verify backend Service exists
kubectl get svc -n monitoring edge-metrics-server
```

### Exporter Pod in CrashLoopBackOff

**Cause**: Config server connection failure + no local config

**Check**:
```bash
# Check Pod logs
kubectl logs -n monitoring -l app.kubernetes.io/name=exporter --tail=50
```

**Solution**:
1. Ensure `edge-metrics-server` service is running and reachable
2. Verify local config fallback exists at `/var/lib/edge-metrics/config.yaml`

### Pod in ImagePullBackOff

**Cause**: Docker image not found

**Solution**:
- When using registry: verify image has been pushed
  ```bash
  docker pull daclab/edge-metrics-exporter:v1.0.0
  ```

### PVC in Pending State

**Cause**: Missing StorageClass or provisioner not configured

**Check**:
```bash
kubectl get storageclass
kubectl describe pvc -n monitoring edge-metrics-data
```

**Solution**:
1. Set `storageClassName` in values file
2. Or disable PVC: `--set server.persistence.enabled=false`

## Architecture

```
Kubernetes Cluster (monitoring namespace)
|
+-- Server (edge-metrics-server)
|   +-- Deployment: edge-metrics-server
|   +-- Service: :8081 (ClusterIP / NodePort)
|   +-- PVC: edge-metrics-data (optional)
|   +-- RBAC, ServiceMonitor
|
+-- Exporter (edge-metrics-exporter)
|   +-- DaemonSet: edge-metrics-exporter-jetson (privileged)
|   +-- DaemonSet: edge-metrics-exporter-generic (non-privileged)
|   +-- Headless Service x2 (ServiceMonitor integration)
|   +-- ServiceMonitor x2 (5s interval)
|
+-- Frontend (edge-metrics-frontend)
    +-- Deployment: edge-metrics-frontend
    +-- Service: :3000 (NodePort 30080)
    +-- API integration: http://edge-metrics-server:8081 (Service DNS)
```

## License

MIT
