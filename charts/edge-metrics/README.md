# edge-metrics

Kubernetes-native edge device metrics collection and monitoring platform.

## Prerequisites

- Kubernetes 1.26+
- Helm 3.x
- (Optional) Prometheus Operator for ServiceMonitor/PrometheusRule
- (Optional) Grafana with sidecar for dashboard auto-loading

## Installation

### From OCI Registry

```bash
helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics \
  --namespace monitoring --create-namespace
```

### From Source

```bash
helm install edge-metrics charts/edge-metrics \
  --namespace monitoring --create-namespace
```

### Development

```bash
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-dev.yaml \
  --namespace monitoring --create-namespace
```

### Production

```bash
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-prod.yaml \
  --namespace monitoring --create-namespace
```

## Upgrading

```bash
helm upgrade edge-metrics charts/edge-metrics --namespace monitoring
```

## Uninstalling

```bash
helm uninstall edge-metrics --namespace monitoring
```

## Configuration

### Global

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `global.namespace` | string | `monitoring` | Target namespace |

### Server

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `server.enabled` | bool | `true` | Enable server component |
| `server.replicaCount` | int | `1` | Replica count (keep 1 for SQLite) |
| `server.image.repository` | string | `ghcr.io/noflowwater/edge-metrics-server` | Image repository |
| `server.image.tag` | string | `""` | Image tag (defaults to appVersion) |
| `server.persistence.enabled` | bool | `true` | Enable PVC for SQLite |
| `server.persistence.size` | string | `256Mi` | PVC size |
| `server.env.DB_PATH` | string | `/data/config.db` | Database path |
| `server.env.SYNC_INTERVAL` | string | `"60"` | Auto-sync interval (seconds) |
| `server.serviceMonitor.enabled` | bool | `false` | Create ServiceMonitor |
| `server.prometheusRule.enabled` | bool | `false` | Create PrometheusRule |

### Exporter

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `exporter.enabled` | bool | `true` | Enable exporter component |
| `exporter.image.repository` | string | `ghcr.io/noflowwater/edge-metrics-exporter` | Image repository |
| `exporter.profiles.jetson.enabled` | bool | `true` | Enable Jetson DaemonSet |
| `exporter.profiles.jetson.nodeSelector` | object | `{device-family: jetson}` | Jetson node selector |
| `exporter.profiles.generic.enabled` | bool | `true` | Enable generic DaemonSet |
| `exporter.profiles.generic.nodeSelector` | object | `{device-family: generic}` | Generic node selector |
| `exporter.commonEnv.CONFIG_SERVER_URL` | string | `http://edge-metrics-server:8081` | Server URL |
| `exporter.hostNetwork` | bool | `true` | Enable host networking |
| `exporter.serviceMonitor.enabled` | bool | `false` | Create ServiceMonitor |

### Frontend

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `frontend.enabled` | bool | `true` | Enable frontend component |
| `frontend.image.repository` | string | `ghcr.io/noflowwater/edge-metrics-front` | Image repository |
| `frontend.service.type` | string | `NodePort` | Service type |
| `frontend.service.port` | int | `3000` | Service port |

## DaemonSet Profiles

The exporter uses profiles to deploy different DaemonSets per device family:

- **jetson**: Privileged access for tegrastats, SYS_RAWIO capability
- **generic**: Restricted PSS-compliant, for Raspberry Pi / OrangePi

Each profile can be independently enabled/disabled and customized with its own nodeSelector, securityContext, and volumes.

## Observability

When ServiceMonitor and PrometheusRule are enabled:

- 5 alert rules: ExporterDown, ServerCrashLoop, DeviceUnreachable, HighGPUTemp, PowerAnomaly
- Grafana dashboards auto-loaded via sidecar (fleet overview + device detail)
