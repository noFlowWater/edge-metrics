# Edge Metrics Server

Centralized edge device configuration management server with Kubernetes integration.

## Features

- Edge device configuration management (CRUD)
- Device status monitoring
- Automatic reload trigger on config changes
- **Kubernetes integration**: Virtualizes external edge devices as Service/Endpoints so Prometheus can scrape them like in-cluster Pods
- **API Key authentication** (optional, via `API_KEY` environment variable)

## Requirements

- Go 1.25+
- SQLite3
- Kubernetes cluster (for K8s integration features)

## Quick Start

### Local Execution

```bash
# Build
go build -o edge-metrics-server

# Run
./edge-metrics-server

# With environment variables
PORT=8080 DB_PATH=/data/config.db ./edge-metrics-server
```

## Docker

### Build Docker Image

```bash
# Basic build
docker build -t edge-metrics-server:latest .

# With specific version tag
docker build -t edge-metrics-server:v1.0.0 .

# Push to registry
docker tag edge-metrics-server:v1.0.0 myregistry.com/edge-metrics-server:v1.0.0
docker push myregistry.com/edge-metrics-server:v1.0.0
```

### Run Docker Image

```bash
# Local run
docker run -d \
  --name edge-metrics-server \
  -p 8081:8081 \
  -v $(pwd)/data:/data \
  edge-metrics-server:latest

# View logs
docker logs -f edge-metrics-server

# Stop and remove
docker stop edge-metrics-server
docker rm edge-metrics-server
```

## Kubernetes Deployment

### Helm (Recommended)

See the [Helm chart README](../charts/edge-metrics/README.md) for installation instructions.

```bash
# Quick install
helm install edge-metrics charts/edge-metrics \
  --namespace monitoring --create-namespace

# Verify
kubectl get pods -n monitoring -l app.kubernetes.io/name=server
kubectl logs -n monitoring -l app.kubernetes.io/name=server --tail=50 -f
```

### Legacy: Script-Based Deployment (Deprecated)

> Deprecated in favor of Helm. Scripts will be removed after 2026-06-01.

<details>
<summary>Click to expand legacy deployment instructions</summary>

#### Build Script

```bash
# Basic build
./scripts/build.sh v1.0.0

# Push to registry
REGISTRY=myregistry.com PUSH=true ./scripts/build.sh v1.0.0

# Multi-platform build (AMD64 + ARM64)
PLATFORM=linux/amd64,linux/arm64 PUSH=true REGISTRY=myregistry.com ./scripts/build.sh v1.0.0
```

| Variable | Default | Description |
|----------|---------|-------------|
| `REGISTRY` | (none) | Docker registry address |
| `PUSH` | false | Push to registry when true |
| `PLATFORM` | (none) | Multi-platform build targets (requires buildx) |

#### Deploy Script

```bash
# Basic deployment (local image, emptyDir)
./scripts/deploy.sh

# With version tag
./scripts/deploy.sh v1.0.0

# Full options
NAMESPACE=monitoring \
REGISTRY=myregistry.com \
USE_PVC=true \
DEPLOY_SERVICEMONITOR=true \
./scripts/deploy.sh v1.0.0
```

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | Kubernetes namespace |
| `REGISTRY` | (none) | Docker registry address |
| `USE_PVC` | false | Enable PVC (emptyDir otherwise) |
| `USE_LOCAL_DB` | false | Mount local config.db (development only) |
| `DEPLOY_SERVICEMONITOR` | false | Deploy ServiceMonitor |

#### Undeploy Script

```bash
./scripts/undeploy.sh
DELETE_PVC=true FORCE=true ./scripts/undeploy.sh
```

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | Target namespace |
| `DELETE_PVC` | false | Delete PVC (permanent data loss!) |
| `DELETE_IMAGE` | false | Delete local Docker image |
| `FORCE` | false | Skip confirmation prompt |

</details>

## Kubernetes Integration

### Overview

edge-metrics-server maps external edge devices (Jetson, Raspberry Pi, etc.) to Kubernetes Service/Endpoints resources, allowing Prometheus to scrape them like in-cluster Pods.

### How It Works

```
+--------------------------------------------------+
| Kubernetes Cluster (monitoring namespace)         |
|                                                   |
|  +--------------------------------------------+  |
|  | Prometheus                                  |  |
|  | - Auto-discovers via ServiceMonitor         |  |
|  | - Scrapes edge-device-* Services            |  |
|  +--------------------------------------------+  |
|              | (scrapes)                          |
|  +--------------------------------------------+  |
|  | Service: edge-device-edge-01                |  |
|  | Endpoints: 192.168.1.10:9102               |  |
|  +--------------------------------------------+  |
|              | (points to)                        |
+--------------+------------------------------------+
               | (external network)
     +---------------------+
     | Edge Device          |
     | IP: 192.168.1.10     |
     | Exporter: :9102      |
     +---------------------+
```

### API Endpoints

#### POST /kubernetes/sync

Syncs all healthy devices to Kubernetes as Service + Endpoints.

```bash
curl -X POST http://edge-metrics-server:8081/kubernetes/sync \
  -H "Content-Type: application/json" \
  -d '{"namespace": "monitoring"}'
```

#### GET /kubernetes/manifests

Generates Kubernetes YAML manifests for healthy devices (for manual apply).

```bash
curl http://edge-metrics-server:8081/kubernetes/manifests?namespace=monitoring > edge-devices.yaml
kubectl apply -f edge-devices.yaml
```

#### DELETE /kubernetes/cleanup

Deletes all edge-device-* resources in the namespace.

```bash
curl -X DELETE http://edge-metrics-server:8081/kubernetes/cleanup?namespace=monitoring
```

### Prometheus Integration

#### ServiceMonitor (with Prometheus Operator)

Enable via Helm:

```bash
helm install edge-metrics charts/edge-metrics \
  --set server.serviceMonitor.enabled=true
```

Or apply manually:

```bash
kubectl apply -f manifests/servicemonitor.yaml
```

Verify:

```bash
kubectl get servicemonitor -n monitoring
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090
# Open http://localhost:9090/targets
```

### Usage Scenarios

#### Manual Sync

```bash
# 1. Register edge device
curl -X PUT http://edge-metrics-server:8081/config/edge-01 \
  -d '{"device_type": "jetson_orin", "ip_address": "192.168.1.10", "port": 9102}'

# 2. Check device status
curl http://edge-metrics-server:8081/devices

# 3. Sync to Kubernetes
curl -X POST http://localhost:8081/kubernetes/sync \
  -H "Content-Type: application/json" \
  -d '{"namespace": "monitoring"}'

# 4. Verify created resources
kubectl get svc,endpoints -n monitoring -l managed_by=edge-metrics-server
```

#### CronJob for Periodic Sync

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: edge-device-sync
  namespace: monitoring
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: sync
            image: curlimages/curl:latest
            args:
            - sh
            - -c
            - |
              curl -X POST http://edge-metrics-server:8081/kubernetes/sync \
                -H "Content-Type: application/json" \
                -d '{"namespace": "monitoring"}'
          restartPolicy: OnFailure
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8081 | Server port |
| `DB_PATH` | ./config.db | SQLite database path |
| `SERVER_URL` | http://localhost:8081 | Self URL (used for K8s sync) |
| `API_KEY` | (empty) | API key for authentication (disabled when empty) |

## API Documentation

See [API.md](./API.md) for the complete API specification.

## Architecture

```
server/
+-- main.go                     # Entry point
+-- database/                   # SQLite database
+-- models/                     # Data models
+-- repository/                 # Database CRUD
+-- handlers/                   # HTTP handlers
|   +-- handlers.go            # Device management API
|   +-- kubernetes_handler.go  # Kubernetes integration API
|   +-- health.go              # Health check utility
+-- middleware/                 # HTTP middleware
|   +-- auth.go                # API Key authentication
+-- router/                     # Route configuration
+-- kubernetes/                 # Kubernetes client
|   +-- client.go              # K8s client initialization
|   +-- service.go             # Service resource management
|   +-- endpoints.go           # Endpoints resource management
|   +-- sync.go                # Sync logic
+-- manifests/                  # Kubernetes manifests (legacy)
+-- scripts/                    # Deployment scripts (legacy)
+-- Dockerfile                  # Multi-stage build
+-- .dockerignore
```

## Security

### RBAC Permissions

edge-metrics-server requires the following Kubernetes permissions:

- **services**: get, list, create, update, patch, delete
- **endpoints**: get, list, create, update, patch, delete
- **servicemonitors** (optional): get, list, create, update, patch, delete

### API Key Authentication

When the `API_KEY` environment variable is set, all requests must include the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8081/config
```

If `API_KEY` is empty, authentication is disabled (default).

### Network Requirements

The Kubernetes Pod must be able to reach external edge devices:
- Allow device port (default 9102) through firewall
- Configure network routing for private networks (VPN/Tailscale)

## Troubleshooting

### Kubernetes client not initialized

```
Kubernetes client not initialized: failed to create Kubernetes config
```

**Cause**: Pod is not using a ServiceAccount, or kubeconfig is missing.

**Solution**:
1. Verify `serviceAccountName: edge-metrics-server` in Deployment
2. Check RBAC resources: `kubectl get sa,role,rolebinding -n monitoring`

### Service created but endpoints empty

```bash
kubectl get endpoints -n monitoring edge-device-edge-01
# No endpoints available
```

**Cause**: Device IP address is not registered.

**Solution**:
1. Check device status: `curl http://server/devices`
2. If IP is empty, register the device with its IP address

### Prometheus not scraping edge devices

**Checklist**:
1. Verify ServiceMonitor has correct label selector
2. Verify Prometheus Operator watches the target namespace
3. Check Prometheus logs for target discovery

## License

MIT
