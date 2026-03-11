# Edge Metrics Exporter

Prometheus exporter for edge device power consumption monitoring. Deployed as K8s DaemonSet across edge nodes.

## Features

- **Multiple Device Support**: Jetson Orin/Xavier/Nano, Raspberry Pi, Orange Pi, LattePanda, Shelly smart plugs
- **Dynamic Config Reload**: Update configuration without restarting the pod
- **Central Config Server**: Manage all device configurations from one place (with bidirectional sync)
- **Local Fallback**: Continue operating with local config if central server is unavailable
- **Prometheus Integration**: Standard Prometheus metrics format with ServiceMonitor
- **Selective Metrics Collection**: Choose which metrics to collect via API or config file
- **Auto-Discovery**: Automatically discover and register new metrics
- **Server Sync**: Changes sync automatically to central server (optional)

## Current Implementation Status

| Device | Status | Method |
|--------|--------|--------|
| Jetson Orin/Xavier/Nano | Implemented | tegrastats |
| Shelly Plug | Implemented | WebSocket + HTTP API |
| Raspberry Pi | TODO | INA260 I2C |
| Orange Pi | TODO | sysfs |
| LattePanda | TODO | RAPL |

## Architecture

```
[Edge Node - DaemonSet Pod]
 +-- Config Loader (central API + local fallback + bidirectional sync)
 +-- Shelly Server (WebSocket + HTTP, background process)
 +-- Collector (device-specific power reading with auto-discovery)
 +-- Exporter
     +-- :9102/metrics (Prometheus)
     +-- :9101 Management API
         +-- GET  /health (Health check)
         +-- POST /reload (Config reload trigger)
         +-- GET  /metrics/list (List all metrics and status)
         +-- POST /metrics/enable (Enable/disable metrics)

[Central Server]
 +-- Config Server (GET /config/{device_id}, PUT /config/{device_id})
 +-- Prometheus (ServiceMonitor, 5s interval)
```

## Deployment (K8s DaemonSet)

### Prerequisites

- Kubernetes cluster (1.26+)
- Node labels: `device-family=jetson` or `device-family=generic`
- Container images available (via Helm install or manual push)

### Helm (Recommended)

See the [Helm chart README](../charts/edge-metrics/README.md) for full installation instructions.

```bash
# Install full stack
helm install edge-metrics charts/edge-metrics \
  --namespace monitoring --create-namespace

# Verify exporter pods
kubectl get pods -n monitoring -l app.kubernetes.io/name=exporter
```

### DaemonSet Profiles

| Profile | Target Nodes | Characteristics |
|---------|-------------|-----------------|
| **Jetson** | jetsono, jetsonx, jetson-nano | `privileged: true`, tegrastats hostPath mount |
| **Generic** | rasp5, orangepi, lattepanda | Non-privileged, security hardened |

Common: `hostNetwork: true`, `dnsPolicy: ClusterFirstWithHostNet`, tini PID 1

### Legacy: Script-Based Deployment (Deprecated)

> Deprecated in favor of Helm. Scripts will be removed after 2026-06-01.

<details>
<summary>Click to expand legacy deployment instructions</summary>

```bash
# Full deployment (Jetson + Generic)
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --all

# Jetson only
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --jetson

# Generic only
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --generic

# Undeploy
./scripts/undeploy.sh --all

# Build multi-arch image
REGISTRY=daclab ./scripts/build.sh v1.0.0
```

Build platforms: `linux/arm64`, `linux/amd64`

</details>

## Configuration

### Environment Variables

Set via DaemonSet Pod env:

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_SERVER_URL` | `http://edge-metrics-server.monitoring.svc.cluster.local:8081` | Config server URL (FQDN required) |
| `CONFIG_TIMEOUT` | `5` | Config server request timeout (seconds) |
| `LOCAL_CONFIG_PATH` | `/config/config.yaml` | Local fallback config path |
| `SHELLY_ENABLED` | `true` | Enable Shelly WebSocket server |
| `NODE_NAME` | (Downward API) | K8s node name |

### Config File (`config.yaml`)

`/var/lib/edge-metrics/config.yaml` (mounted via hostPath):

```yaml
device_type: "jetson_orin"  # Device type
interval: 1                  # Collection interval (seconds)
port: 9102                   # Prometheus metrics port
reload_port: 9101            # Management API port

metrics:
  jetson_power_vdd_gpu_soc_watts: true
  jetson_power_vdd_cpu_cv_watts: true
  jetson_temp_cpu_celsius: true
  jetson_temp_gpu_celsius: false
  jetson_ram_used_percent: true
```

Config priority: Config Server fetch -> local fallback -> error

## Usage

### Prometheus Configuration

ServiceMonitor automatically registers Prometheus targets. For manual configuration:

```yaml
scrape_configs:
  - job_name: 'edge-power'
    static_configs:
      - targets:
          - 'edge-01:9102'
          - 'edge-02:9102'
    scrape_interval: 5s
```

### Example Queries

```promql
# Total power consumption by device (Shelly external measurement)
power_total_watts{device_type="jetson_orin"}

# Internal power rails
jetson_power_vdd_gpu_soc_watts{device_type="jetson_orin"}

# Sum of all devices (Shelly)
sum(power_total_watts)

# Average power over 5 minutes
avg_over_time(power_total_watts[5m])
```

### Management API

#### Health Check

```bash
curl http://<node-ip>:9101/health
```

```json
{
  "status": "healthy",
  "device_id": "jetson-orin",
  "device_type": "jetson_orin",
  "uptime_seconds": 3600,
  "metrics_count": 21,
  "enabled_metrics": 21
}
```

#### List All Metrics

```bash
curl http://<node-ip>:9101/metrics/list
```

```json
{
  "metrics": {
    "jetson_power_vdd_gpu_soc_watts": true,
    "jetson_power_vdd_cpu_cv_watts": true,
    "jetson_temp_cpu_celsius": true
  },
  "device_type": "jetson_orin",
  "source": "local"
}
```

#### Enable/Disable Metrics

```bash
curl -X POST http://<node-ip>:9101/metrics/enable \
  -H "Content-Type: application/json" \
  -d '{
    "jetson_gpu_usage_percent": true,
    "jetson_cpu_avg_usage_percent": true
  }'
```

#### Config Reload

```bash
curl -X POST http://<node-ip>:9101/reload
```

## Metrics

### Auto-Discovery

The exporter automatically discovers all available metrics from the collector. When a new metric is detected:
1. It's automatically added to `config.yaml` with `enabled: false`
2. Changes are synced to Config Server (if available)
3. You can enable it via the API or config file

### Jetson Orin/Xavier

Example available metrics (dynamically discovered from tegrastats):

| Metric | Description | Unit |
|--------|-------------|------|
| `jetson_power_vdd_gpu_soc_watts` | GPU/SoC power consumption | Watts |
| `jetson_power_vdd_cpu_cv_watts` | CPU power consumption | Watts |
| `jetson_temp_cpu_celsius` | CPU temperature | Celsius |
| `jetson_temp_gpu_celsius` | GPU temperature | Celsius |
| `jetson_ram_used_mb` | RAM usage | MB |
| `jetson_ram_used_percent` | RAM usage | Percent |
| `jetson_cpu_avg_usage_percent` | Average CPU usage | Percent |
| `jetson_gpu_usage_percent` | GPU usage | Percent |

### Shelly Plug

| Metric | Description | Unit |
|--------|-------------|------|
| `power_total_watts` | Total power consumption | Watts |
| `power_voltage_volts` | Voltage | Volts |
| `power_current_amps` | Current | Amps |

All metrics include labels:
- `device_type`: Device type (e.g., `jetson_orin`, `raspberry_pi`)
- `hostname`: Device hostname

## Adding New Devices

1. Create new collector in `collectors/`:

```python
# collectors/new_device.py
from .base import BaseCollector

class NewDeviceCollector(BaseCollector):
    @classmethod
    def metric_names(cls):
        return ["power_total_watts"]

    def get_metrics(self):
        # Implement power reading logic
        return {"power_total_watts": 10.5}
```

2. Register in `collectors/__init__.py`:

```python
elif device_type == "new_device":
    from .new_device import NewDeviceCollector
    return NewDeviceCollector(config)
```

3. Update config:

```yaml
device_type: "new_device"
```

## Troubleshooting

### Pod in CrashLoopBackOff

```bash
# Check Pod logs
kubectl logs -n monitoring -l app.kubernetes.io/name=exporter --tail=50

# Health check
curl http://<node-ip>:9101/health
```

### Config Server Connection Failure

Verify server service is running:
```bash
kubectl get svc -n monitoring edge-metrics-server
```

Check local fallback config:
```bash
kubectl exec -n monitoring <pod-name> -- cat /config/config.yaml
```

### tegrastats Permission Denied (Jetson)

Jetson DaemonSet runs with `privileged: true`, so there should be no permission issues.
Check tegrastats binary mount:
```bash
kubectl exec -n monitoring <pod-name> -- ls -la /usr/bin/tegrastats
```

### Shelly Device Not Connecting

Check that the Shelly plug firmware's WebSocket target IP matches the node IP.
Since `hostNetwork: true`, Pod IP = node IP.

```bash
# Check Shelly server status
curl http://<node-ip>:8766/devices
```

## Development

### Local Execution (Development)

```bash
pip3 install -r requirements.txt
python3 exporter.py
```

### Debugging

```bash
# Real-time Pod logs
kubectl logs -n monitoring -l app.kubernetes.io/name=exporter --tail=50 -f

# Specific node Pod
kubectl logs -n monitoring <pod-name> -f
```

## TODO

### Collectors
- [ ] Implement Raspberry Pi collector (INA260)
- [ ] Implement Orange Pi collector (sysfs)
- [ ] Implement LattePanda collector (RAPL)

### Completed
- [x] Selective metrics collection (metrics dict format)
- [x] Auto-discovery of new metrics
- [x] Management API (GET /metrics/list, POST /metrics/enable)
- [x] Bidirectional server sync (client -> server)
- [x] Shelly plug WebSocket collector
- [x] K8s DaemonSet deployment (Jetson + Generic)
- [x] Multi-arch container image (arm64 + amd64)
- [x] Prometheus ServiceMonitor integration
- [x] tini PID 1 for zombie process prevention

## License

MIT License
