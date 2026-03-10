# Edge Metrics Exporter

Prometheus exporter for edge device power consumption monitoring. Deployed as K8s DaemonSet across KubeEdge edge nodes.

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
 ├─ Config Loader (central API + local fallback + bidirectional sync)
 ├─ Shelly Server (WebSocket + HTTP, background process)
 ├─ Collector (device-specific power reading with auto-discovery)
 └─ Exporter
     ├─ :9102/metrics (Prometheus)
     └─ :9101 Management API
         ├─ GET  /health (Health check)
         ├─ POST /reload (Config reload trigger)
         ├─ GET  /metrics/list (List all metrics and status)
         └─ POST /metrics/enable (Enable/disable metrics)

[Central Server]
 ├─ Config Server (GET /config/{device_id}, PUT /config/{device_id})
 └─ Prometheus (ServiceMonitor, 5s interval)
```

## Deployment (K8s DaemonSet)

### Prerequisites

- K8s cluster with KubeEdge
- 노드 라벨: `device-family=jetson` 또는 `device-family=generic`
- Docker Hub 이미지: `daclab/edge-metrics-exporter:v1.0.0`
- EdgeMesh: `edge-metrics-server` 서비스에 EdgeMesh 라벨 필요

### Quick Deploy

```bash
# 전체 배포 (Jetson + Generic)
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --all

# Jetson만
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --jetson

# Generic만
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --generic
```

### DaemonSet 프로파일

| 프로파일 | 대상 | 특성 |
|---------|------|------|
| **Jetson** (`daemonset.yaml`) | jetsono, jetsonx, jetson-nano | `privileged: true`, tegrastats hostPath 마운트 |
| **Generic** (`daemonset-generic.yaml`) | rasp5, orangepi, lattepanda | Non-privileged, security hardened |

공통: `hostNetwork: true`, `dnsPolicy: ClusterFirstWithHostNet`, tini PID 1

### 이미지 빌드

```bash
# 멀티아키텍처 빌드 + Docker Hub push
REGISTRY=daclab ./scripts/build.sh v1.0.0
```

빌드 플랫폼: `linux/arm64`, `linux/amd64`

### 삭제

```bash
# 전체 삭제
./scripts/undeploy.sh --all

# Jetson만 삭제
./scripts/undeploy.sh --jetson
```

### 통합 배포 (Server + Exporter + Frontend)

```bash
# 전체 스택 배포
REGISTRY=daclab ./deploy/deploy-all.sh v1.0.0

# 전체 스택 삭제
./deploy/undeploy-all.sh
```

## Configuration

### Environment Variables

DaemonSet Pod env로 설정:

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_SERVER_URL` | `http://edge-metrics-server.monitoring.svc.cluster.local:8081` | Config 서버 URL (FQDN 필수) |
| `CONFIG_TIMEOUT` | `5` | Config 서버 요청 타임아웃 (초) |
| `LOCAL_CONFIG_PATH` | `/config/config.yaml` | 로컬 fallback config 경로 |
| `SHELLY_ENABLED` | `true` | Shelly WebSocket 서버 활성화 |
| `NODE_NAME` | (Downward API) | K8s 노드 이름 |

### Config File (`config.yaml`)

`/var/lib/edge-metrics/config.yaml` (hostPath로 마운트):

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

Config 우선순위: Config Server fetch → 로컬 fallback → 에러

## Usage

### Prometheus Configuration

ServiceMonitor가 자동으로 Prometheus 타겟을 등록합니다. 수동 설정이 필요한 경우:

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
# Total power consumption by device
power_total_watts{device_type="jetson_orin"}

# Shelly plug power
shelly_power_total_watts

# Sum of all devices
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
| `shelly_power_total_watts` | Total power consumption | Watts |
| `shelly_power_voltage_volts` | Voltage | Volts |
| `shelly_power_current_amps` | Current | Amps |
| `shelly_power_frequency_hz` | Frequency | Hz |
| `shelly_energy_total_wh` | Total energy consumed | Wh |
| `shelly_temperature_celsius` | Device temperature | Celsius |

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

### Pod가 CrashLoopBackOff

```bash
# Pod 로그 확인
kubectl logs -n monitoring -l app=edge-metrics-exporter-jetson --tail=50

# Health 확인
curl http://<node-ip>:9101/health
```

### Config 서버 연결 실패

EdgeMesh 라벨 확인:
```bash
kubectl get svc -n monitoring edge-metrics-server --show-labels
# 필수 라벨: kubeedge.io/edgemesh-service=true
```

로컬 fallback config 확인:
```bash
kubectl exec -n monitoring <pod-name> -- cat /config/config.yaml
```

### tegrastats permission denied (Jetson)

Jetson DaemonSet은 `privileged: true`로 실행되므로 권한 문제가 없어야 합니다.
tegrastats 바이너리 마운트 확인:
```bash
kubectl exec -n monitoring <pod-name> -- ls -la /usr/bin/tegrastats
```

### Shelly device 연결 안 됨

Shelly plug 펌웨어의 WebSocket 대상 IP가 해당 노드 IP와 일치하는지 확인.
`hostNetwork: true`이므로 Pod IP = 노드 IP.

```bash
# Shelly 서버 상태 확인
curl http://<node-ip>:8766/devices
```

## Development

### 로컬 실행 (개발용)

```bash
pip3 install -r requirements.txt
python3 exporter.py
```

### 디버깅

```bash
# Pod 로그 실시간
kubectl logs -n monitoring -l 'app in (edge-metrics-exporter-jetson,edge-metrics-exporter-generic)' --tail=50 -f

# 특정 노드 Pod
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
- [x] Bidirectional server sync (client → server)
- [x] Shelly plug WebSocket collector
- [x] K8s DaemonSet deployment (Jetson + Generic)
- [x] Multi-arch container image (arm64 + amd64)
- [x] Prometheus ServiceMonitor integration
- [x] tini PID 1 for zombie process prevention

## License

MIT License
