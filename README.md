# Edge Metrics

Kubernetes-native edge device metrics collection and monitoring platform. Collects hardware metrics (power, temperature, GPU utilization) from heterogeneous edge devices (Jetson, Raspberry Pi, etc.) and exposes them as Prometheus metrics.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                 Kubernetes Cluster                   │
│                                                     │
│  ┌──────────┐   ┌──────────────┐   ┌────────────┐  │
│  │ Frontend │──▶│    Server    │   │  Prometheus │  │
│  │ (React)  │   │ (Go/SQLite)  │   │             │  │
│  └──────────┘   └──────┬───────┘   └──────┬──────┘  │
│                        │                   │         │
│            ┌───────────┼───────────┐       │         │
│            ▼           ▼           ▼       │         │
│     ┌──────────┐ ┌──────────┐ ┌──────────┐│         │
│     │ Exporter │ │ Exporter │ │ Exporter ││         │
│     │ (Jetson) │ │ (Generic)│ │ (Stub)   │◀─────────┘
│     └──────────┘ └──────────┘ └──────────┘          │
│     DaemonSet     DaemonSet                          │
└─────────────────────────────────────────────────────┘
```

## Quick Start

### Helm Install

```bash
# From OCI registry
helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics \
  --namespace monitoring --create-namespace

# From source (development)
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-dev.yaml \
  --namespace monitoring --create-namespace
```

### Stub Mode (No Hardware Required)

```bash
helm install edge-metrics charts/edge-metrics \
  --set exporter.profiles.jetson.enabled=false \
  --set exporter.profiles.generic.env.STUB_MODE=true \
  --set exporter.profiles.generic.nodeSelector=null \
  --namespace monitoring --create-namespace
```

### Access

```bash
# Frontend Dashboard
kubectl port-forward -n monitoring svc/edge-metrics-frontend 3000:3000

# Backend API
kubectl port-forward -n monitoring svc/edge-metrics-server 8081:8081

# Exporter Metrics
curl http://<edge-node-ip>:9102/metrics
```

## Components

| Component | Description | Image |
|-----------|-------------|-------|
| **Server** | Config management API + K8s sync | `ghcr.io/noflowwater/edge-metrics-server` |
| **Exporter** | Prometheus metrics collector (DaemonSet) | `ghcr.io/noflowwater/edge-metrics-exporter` |
| **Frontend** | React dashboard | `ghcr.io/noflowwater/edge-metrics-front` |

## Configuration

See [charts/edge-metrics/values.yaml](charts/edge-metrics/values.yaml) for all configuration options.

Key configurations:
- `server.persistence.enabled` — PVC for SQLite database (default: true)
- `exporter.profiles` — DaemonSet profiles per device family
- `server.serviceMonitor.enabled` — Prometheus ServiceMonitor
- `server.prometheusRule.enabled` — Alert rules

## Development

```bash
# Run all tests
make test-all

# Individual test suites
make test-server      # Go unit tests
make test-exporter    # Python tests
make test-helm        # Helm chart tests
make kubeconform      # K8s schema validation
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development setup.

## License

[MIT](LICENSE)
