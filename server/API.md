# Edge Metrics Server API Specification

## Overview

Edge Metrics Server는 edge-metrics-exporter 클라이언트를 위한 중앙 설정 관리 서버입니다.

- **Base URL**: `http://localhost:8081`
- **Content-Type**: `application/json`

---

## Endpoints

### GET /config

모든 디바이스의 설정 목록을 조회합니다.

**Request**
```
GET /config
```

**Response (200 OK)**
```json
{
  "configs": [
    {
      "device_id": "edge-01",
      "device_type": "jetson_orin",
      "port": 9100,
      "reload_port": 9101,
      "enabled_metrics": ["jetson_power_vdd_gpu_soc_watts"],
      "jetson": {"use_tegrastats": true}
    },
    {
      "device_id": "edge-02",
      "device_type": "raspberry_pi",
      "port": 9100,
      "reload_port": 9101
    }
  ],
  "total": 2
}
```

**Example**
```bash
curl http://localhost:8081/config
```

---

### GET /config/{device_id}

디바이스별 설정을 조회합니다.

**Request**
```
GET /config/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname (예: `edge-01`, `orin-desktop`) |

**Response (200 OK)**
```json
{
  "device_type": "jetson_orin",
  "port": 9100,
  "reload_port": 9101,
  "enabled_metrics": [
    "jetson_power_vdd_gpu_soc_watts",
    "jetson_temp_cpu_celsius"
  ],
  "jetson": {
    "use_tegrastats": true
  }
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device",
  "message": "No configuration available for this device"
}
```

**Example**
```bash
curl http://localhost:8081/config/edge-01
```

---

### PUT /config/{device_id}

디바이스 설정을 생성하거나 업데이트합니다 (Upsert).

**Request**
```
PUT /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "155.230.34.203",
  "port": 9100,
  "reload_port": 9101,
  "enabled_metrics": [
    "jetson_power_vdd_gpu_soc_watts"
  ],
  "jetson": {
    "use_tegrastats": true
  }
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| device_type | string | **Yes** | - | 디바이스 타입 |
| ip_address | string | No | 기존 IP 유지 | 디바이스 IP 주소 (미제공 시 기존 IP 유지) |
| port | integer | No | 9100 | Prometheus 메트릭 서버 포트 |
| reload_port | integer | No | 9101 | 설정 리로드 트리거 포트 |
| enabled_metrics | array | No | null | 수집할 메트릭 목록 (null=전체) |
| * | object | No | - | 디바이스별 추가 설정 (shelly, jetson 등) |

**Response (200 OK) - 새 디바이스 등록**
```json
{
  "status": "registered",
  "device_id": "orin-desktop",
  "reload_triggered": false
}
```

**Response (200 OK) - 기존 디바이스 업데이트**
```json
{
  "status": "updated",
  "device_id": "edge-01",
  "reload_triggered": true
}
```

> **Note**: `reload_triggered`가 `true`이면 exporter의 `/reload` 엔드포인트가 호출되어 설정이 즉시 적용됩니다.

**Response (400 Bad Request)**
```json
{
  "error": "Missing required field",
  "message": "device_type is required"
}
```

```json
{
  "error": "invalid_ip_address",
  "message": "Invalid IP address format: not_an_ip"
}
```

**Example - 새 디바이스 등록**
```bash
curl -X PUT http://localhost:8081/config/orin-desktop \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "jetson_orin",
    "ip_address": "155.230.34.203",
    "jetson": {"use_tegrastats": true}
  }'
```

---

### POST /config/{device_id}

새 디바이스를 등록합니다. 이미 존재하는 경우 409 Conflict를 반환합니다.

**Request**
```
POST /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "155.230.34.203",
  "port": 9100,
  "reload_port": 9101
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| device_type | string | **Yes** | - | 디바이스 타입 |
| ip_address | string | **Yes** | - | 디바이스 IP 주소 |
| port | integer | No | 9100 | Prometheus 메트릭 서버 포트 |
| reload_port | integer | No | 9101 | 설정 리로드 트리거 포트 |
| enabled_metrics | array | No | null | 수집할 메트릭 목록 |
| * | object | No | - | 디바이스별 추가 설정 (shelly, jetson 등) |

**Response (201 Created)**
```json
{
  "status": "created",
  "device_id": "new-device"
}
```

**Response (400 Bad Request)**
```json
{
  "error": "ip_address_required",
  "message": "Device IP address must be specified in configuration"
}
```

```json
{
  "error": "invalid_ip_address",
  "message": "Invalid IP address format: not_an_ip"
}
```

**Response (409 Conflict)**
```json
{
  "error": "Device already exists",
  "device_id": "edge-01",
  "message": "Use PUT to update existing device"
}
```

**Example**
```bash
curl -X POST http://localhost:8081/config/new-device \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "raspberry_pi",
    "ip_address": "155.230.34.205"
  }'
```

---

### PATCH /config/{device_id}

디바이스 설정을 부분 업데이트합니다. 전달된 필드만 변경됩니다.

**Request**
```
PATCH /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Request Body**
```json
{
  "port": 9200
}
```

또는 IP 주소 변경:
```json
{
  "ip_address": "155.230.34.210"
}
```

> 변경하고자 하는 필드만 포함하면 됩니다. `null`을 전달하면 필드를 기본값으로 리셋하거나 삭제합니다. (단, `ip_address`는 `null`이어도 기존 IP 유지)

**Response (200 OK)**
```json
{
  "status": "patched",
  "device_id": "edge-01",
  "reload_triggered": true
}
```

**Response (400 Bad Request)**
```json
{
  "error": "invalid_ip_address",
  "message": "Invalid IP address format: not_an_ip"
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device",
  "message": "Use POST or PUT to create new device"
}
```

**Example**
```bash
# 포트 변경
curl -X PATCH http://localhost:8081/config/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"port": 9200}'

# IP 주소 변경
curl -X PATCH http://localhost:8081/config/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "155.230.34.210"}'
```

---

### DELETE /config/{device_id}

디바이스 설정을 삭제합니다.

**Request**
```
DELETE /config/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Response (200 OK)**
```json
{
  "status": "deleted",
  "device_id": "edge-01"
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device"
}
```

**Example**
```bash
curl -X DELETE http://localhost:8081/config/edge-01
```

---

### GET /health

서버 상태를 확인합니다.

**Request**
```
GET /health
```

**Response (200 OK)**
```json
{
  "status": "healthy",
  "service": "config-server",
  "version": "1.0.0"
}
```

**Example**
```bash
curl http://localhost:8081/health
```

---

### GET /devices

등록된 모든 디바이스와 상태를 조회합니다.

**Request**
```
GET /devices
```

**Response (200 OK)**
```json
{
  "devices": [
    {
      "device_id": "edge-01",
      "device_type": "jetson_orin",
      "ip_address": "192.168.1.10",
      "port": 9100,
      "reload_port": 9101,
      "status": "healthy",
      "last_seen": "2024-01-15T10:30:00Z"
    },
    {
      "device_id": "edge-02",
      "device_type": "jetson_xavier",
      "ip_address": "192.168.1.11",
      "port": 9100,
      "reload_port": 9101,
      "status": "unreachable",
      "error": "connection refused"
    }
  ],
  "total": 2,
  "healthy": 1,
  "unhealthy": 1
}
```

| Field | Type | Description |
|-------|------|-------------|
| devices | array | 디바이스 상태 목록 |
| total | integer | 전체 디바이스 수 |
| healthy | integer | 정상 디바이스 수 |
| unhealthy | integer | 비정상 디바이스 수 |

**Device Status Fields**

| Field | Type | Description |
|-------|------|-------------|
| device_id | string | 디바이스 ID |
| device_type | string | 디바이스 타입 |
| ip_address | string | 디바이스 IP 주소 |
| port | integer | 메트릭 서버 포트 |
| reload_port | integer | 리로드 트리거 포트 |
| status | string | healthy, unhealthy, unreachable, unknown |
| last_seen | string | 마지막 응답 시간 (healthy인 경우) |
| error | string | 에러 메시지 (비정상인 경우) |

**Example**
```bash
curl http://localhost:8081/devices
```

---

### GET /devices/{device_id}/status

특정 디바이스의 상태를 조회합니다.

**Request**
```
GET /devices/{device_id}/status
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "device_type": "jetson_orin",
  "ip_address": "192.168.1.10",
  "port": 9100,
  "reload_port": 9101,
  "status": "healthy",
  "last_seen": "2024-01-15T10:30:00Z"
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device"
}
```

**Example**
```bash
curl http://localhost:8081/devices/edge-01/status
```

---

### PATCH /devices/{device_id}

디바이스의 기본 정보만 수정합니다 (device_type, ip_address, port, reload_port).
이 API는 데이터베이스만 업데이트하고 디바이스 reload는 트리거하지 않습니다.

**수정 가능한 필드**: device_type, ip_address, port, reload_port
**수정 불가능한 필드**: enabled_metrics, extra_config (jetson, shelly 등)

**Request**
```
PATCH /devices/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "192.168.1.20",
  "port": 9100,
  "reload_port": 9101
}
```

모든 필드는 선택적입니다. 제공된 필드만 업데이트됩니다.

**Response (200 OK)**
```json
{
  "status": "updated",
  "device_id": "edge-01",
  "message": "Device basic information updated (reload not triggered)"
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device"
}
```

**Response (400 Bad Request - Invalid IP)**
```json
{
  "error": "invalid_ip_address",
  "message": "Invalid IP address format: 192.168.1.999"
}
```

**Example**
```bash
# IP 주소만 변경
curl -X PATCH http://localhost:8081/devices/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "192.168.1.20"}'

# 여러 필드 동시 변경
curl -X PATCH http://localhost:8081/devices/edge-01 \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "jetson_orin",
    "ip_address": "192.168.1.25",
    "port": 9100,
    "reload_port": 9101
  }'
```

**기존 API와의 차이점**
- `PATCH /config/{device_id}`: 모든 필드 수정 가능, reload 트리거 O
- `PATCH /devices/{device_id}`: 기본 필드만 수정 가능, reload 트리거 X

---

### GET /devices/{device_id}/local-config

디바이스의 로컬 config.yaml 파일 내용을 조회합니다.
서버가 디바이스의 `GET :9101/config` 엔드포인트를 프록시하여 CORS 이슈를 해결합니다.

**Request**
```
GET /devices/{device_id}/local-config
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Response (200 OK)**
```json
{
  "device_type": "jetson_orin",
  "port": 9100,
  "reload_port": 9101,
  "interval": 10,
  "metrics": {
    "jetson_power_vdd_gpu_soc_watts": true,
    "jetson_power_vdd_cpu_cv_watts": true
  },
  "jetson": {
    "model": "NVIDIA Jetson AGX Orin"
  }
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device"
}
```

**Response (400 Bad Request)**
```json
{
  "error": "No IP address",
  "device_id": "edge-01",
  "message": "Device has no IP address configured"
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Device unreachable",
  "device_id": "edge-01",
  "message": "Failed to connect to device: connection refused"
}
```

**Response (502 Bad Gateway)**
```json
{
  "error": "Device error",
  "device_id": "edge-01",
  "message": "Device returned HTTP 500"
}
```

**Example**
```bash
curl http://localhost:8081/devices/edge-01/local-config
```

---

### POST /devices/{device_id}/reload

특정 디바이스에 수동으로 reload를 트리거합니다.

**Request**
```
POST /devices/{device_id}/reload
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | 디바이스 hostname |

**Response (200 OK)**
```json
{
  "status": "reloaded",
  "device_id": "edge-01"
}
```

**Response (404 Not Found)**
```json
{
  "error": "Device not found",
  "device_id": "unknown-device"
}
```

**Response (503 Service Unavailable)**
```json
{
  "status": "failed",
  "device_id": "edge-01",
  "error": "connection refused"
}
```

**Example**
```bash
curl -X POST http://localhost:8081/devices/edge-01/reload
```

---

### POST /devices/reload

모든 디바이스에 일괄 reload를 트리거합니다.

**Request**
```
POST /devices/reload
```

**Response (200 OK)**
```json
{
  "results": [
    {
      "device_id": "edge-01",
      "status": "reloaded"
    },
    {
      "device_id": "edge-02",
      "status": "failed",
      "error": "connection refused"
    }
  ],
  "total": 2,
  "success": 1,
  "failed": 1
}
```

**Example**
```bash
curl -X POST http://localhost:8081/devices/reload
```

---

### GET /metrics/summary

전체 시스템 요약 통계를 조회합니다.

**Request**
```
GET /metrics/summary
```

**Response (200 OK)**
```json
{
  "total": 5,
  "healthy": 3,
  "unhealthy": 2,
  "by_device_type": {
    "jetson_orin": 2,
    "raspberry_pi": 2,
    "shelly": 1
  }
}
```

**Example**
```bash
curl http://localhost:8081/metrics/summary
```

---

## Kubernetes Integration

### GET /kubernetes/status

전체 Kubernetes 동기화 상태를 조회합니다.

**Request**
```
GET /kubernetes/status?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | 조회할 네임스페이스 |

**Response (200 OK)**
```json
{
  "kubernetes_enabled": true,
  "namespace": "monitoring",
  "total_k8s_resources": 5,
  "total_registered_devices": 7,
  "synced": 5,
  "unsynced": 2,
  "resources": [
    {
      "device_id": "edge-01",
      "service_exists": true,
      "endpoints_exists": true
    },
    {
      "device_id": "edge-02",
      "service_exists": false,
      "endpoints_exists": false
    }
  ]
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl http://localhost:8081/kubernetes/status?namespace=monitoring
```

---

### GET /kubernetes/health

Kubernetes 연결 상태 및 RBAC 권한을 확인합니다.

**Request**
```
GET /kubernetes/health?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | 확인할 네임스페이스 |

**Response (200 OK)**
```json
{
  "kubernetes_available": true,
  "client_initialized": true,
  "namespace_accessible": true,
  "rbac_permissions": {
    "namespace": "ok",
    "services": "ok",
    "endpoints": "ok"
  }
}
```

**Response (503 Service Unavailable)**
```json
{
  "kubernetes_available": false,
  "client_initialized": false,
  "namespace_accessible": false,
  "rbac_permissions": {}
}
```

**Example**
```bash
curl http://localhost:8081/kubernetes/health
```

---

### POST /kubernetes/sync

현재 healthy 상태인 모든 디바이스를 Kubernetes Service + Endpoints로 동기화합니다.

**Request**
```
POST /kubernetes/sync
Content-Type: application/json
```

**Request Body**
```json
{
  "namespace": "monitoring"
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| namespace | string | No | monitoring | 동기화 대상 Kubernetes 네임스페이스 |

**Response (200 OK)**
```json
{
  "status": "synced",
  "created": [
    {
      "device_id": "edge-01",
      "service": "edge-device-edge-01",
      "status": "created"
    }
  ],
  "updated": [
    {
      "device_id": "edge-02",
      "service": "edge-device-edge-02",
      "status": "updated"
    }
  ],
  "deleted": [],
  "failed": [],
  "total_healthy": 2
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl -X POST http://localhost:8081/kubernetes/sync \
  -H "Content-Type: application/json" \
  -d '{"namespace": "monitoring"}'
```

**동작:**
1. GET /devices API를 호출하여 healthy 디바이스 목록 조회
2. 각 디바이스마다 Service + Endpoints 리소스 생성/업데이트
   - Service 이름: `edge-device-{device_id}`
   - Endpoints IP: 디바이스의 `ip_address`
   - 포트: 디바이스의 `port` (기본 9100)
   - 레이블: `app=edge-exporter`, `device_id`, `device_type`, `managed_by=edge-metrics-server`
3. DB에는 있지만 unhealthy하거나 삭제된 디바이스의 리소스는 삭제
4. 결과 반환

---

### POST /kubernetes/sync/{device_id}

특정 디바이스만 Kubernetes에 동기화합니다.

**Request**
```
POST /kubernetes/sync/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | 동기화할 디바이스 ID |
| namespace | string | query | monitoring | 동기화 대상 네임스페이스 |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "created"
}
```

**Response (200 OK - Failed)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "failed",
  "error": "device is not healthy"
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl -X POST http://localhost:8081/kubernetes/sync/edge-01?namespace=monitoring
```

---

### GET /kubernetes/manifests

Healthy 디바이스들의 Kubernetes YAML 매니페스트를 생성합니다 (수동 적용용).

**Request**
```
GET /kubernetes/manifests?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | 매니페스트 생성 대상 네임스페이스 |

**Response (200 OK)**
```yaml
# Kubernetes manifests for edge devices
# Generated for namespace: monitoring

---
apiVersion: v1
kind: Service
metadata:
  name: edge-device-edge-01
  namespace: monitoring
  labels:
    app: edge-exporter
    device_id: edge-01
    device_type: jetson_orin
    managed_by: edge-metrics-server
spec:
  clusterIP: None
  ports:
  - name: metrics
    port: 9100
    targetPort: 9100
    protocol: TCP
---
apiVersion: v1
kind: Endpoints
metadata:
  name: edge-device-edge-01
  namespace: monitoring
  labels:
    app: edge-exporter
    device_id: edge-01
    managed_by: edge-metrics-server
subsets:
- addresses:
  - ip: 192.168.1.10
  ports:
  - name: metrics
    port: 9100
    protocol: TCP

# ... (추가 디바이스들)
```

**Example**
```bash
# YAML 생성 및 저장
curl http://localhost:8081/kubernetes/manifests?namespace=monitoring > edge-devices.yaml

# Kubernetes에 적용
kubectl apply -f edge-devices.yaml
```

**동작:**
1. 모든 디바이스 설정 조회
2. 각 디바이스의 health 체크
3. Healthy 디바이스들만 YAML 매니페스트 생성
4. text/plain으로 반환

---

### GET /kubernetes/resources/{device_id}

특정 디바이스의 Kubernetes 리소스 상세 정보를 조회합니다.

**Request**
```
GET /kubernetes/resources/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | 조회할 디바이스 ID |
| namespace | string | query | monitoring | 조회할 네임스페이스 |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "service": {
    "name": "edge-device-edge-01",
    "exists": true,
    "cluster_ip": "None",
    "ports": [
      {
        "name": "metrics",
        "port": 9100
      }
    ]
  },
  "endpoints": {
    "name": "edge-device-edge-01",
    "exists": true,
    "ready_addresses": ["192.168.1.10:9100"],
    "not_ready_addresses": []
  },
  "prometheus_target": "http://edge-device-edge-01.monitoring.svc:9100/metrics"
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl http://localhost:8081/kubernetes/resources/edge-01?namespace=monitoring
```

---

### DELETE /kubernetes/resources/{device_id}

특정 디바이스의 Kubernetes 리소스를 삭제합니다.

**Request**
```
DELETE /kubernetes/resources/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | 삭제할 디바이스 ID |
| namespace | string | query | monitoring | 삭제할 네임스페이스 |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "deleted"
}
```

**Response (200 OK - Failed)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "failed",
  "error": "delete service: services \"edge-device-edge-01\" not found"
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl -X DELETE http://localhost:8081/kubernetes/resources/edge-01?namespace=monitoring
```

---

### DELETE /kubernetes/cleanup

monitoring 네임스페이스의 모든 edge-device-* 리소스를 삭제합니다.

**Request**
```
DELETE /kubernetes/cleanup?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | 정리할 네임스페이스 |

**Response (200 OK)**
```json
{
  "status": "cleaned",
  "deleted_services": [
    "edge-device-edge-01",
    "edge-device-edge-02"
  ],
  "deleted_endpoints": [
    "edge-device-edge-01",
    "edge-device-edge-02"
  ],
  "namespace": "monitoring"
}
```

**Response (503 Service Unavailable)**
```json
{
  "error": "Kubernetes client not initialized",
  "message": "Server not running in Kubernetes environment or kubeconfig not found"
}
```

**Example**
```bash
curl -X DELETE http://localhost:8081/kubernetes/cleanup?namespace=monitoring
```

**동작:**
1. `managed_by=edge-metrics-server` 레이블을 가진 모든 Service 조회
2. 모든 Service 삭제
3. `managed_by=edge-metrics-server` 레이블을 가진 모든 Endpoints 조회
4. 모든 Endpoints 삭제
5. 삭제된 리소스 목록 반환

---

## Device Types

지원되는 디바이스 타입:

| device_type | Description | Extra Config |
|-------------|-------------|--------------|
| `jetson_orin` | NVIDIA Jetson Orin | `jetson` |
| `jetson_xavier` | NVIDIA Jetson Xavier | `jetson` |
| `jetson_nano` | NVIDIA Jetson Nano | `jetson` |
| `jetson` | Generic NVIDIA Jetson | `jetson` |
| `raspberry_pi` | Raspberry Pi | - |
| `orange_pi` | Orange Pi | - |
| `lattepanda` | LattePanda | - |
| `shelly` | Shelly smart plug | `shelly` |

---

## Extra Config Examples

### Jetson
```json
{
  "device_type": "jetson_orin",
  "jetson": {
    "use_tegrastats": true
  }
}
```

### Shelly
```json
{
  "device_type": "shelly",
  "shelly": {
    "host": "192.168.1.100",
    "switch_id": 0
  }
}
```

### INA260
```json
{
  "device_type": "jetson_orin",
  "ina260": {
    "i2c_address": "0x40"
  }
}
```

---

## Error Responses

모든 에러 응답은 다음 형식을 따릅니다:

```json
{
  "error": "Error type",
  "device_id": "device-id",
  "message": "Detailed error message"
}
```

| Status Code | Description |
|-------------|-------------|
| 200 | 성공 |
| 201 | 생성됨 (POST) |
| 400 | 잘못된 요청 (필수 필드 누락, 잘못된 JSON, 잘못된 IP 주소) |
| 404 | 디바이스를 찾을 수 없음 |
| 409 | 충돌 (이미 존재하는 디바이스) |
| 500 | 서버 내부 오류 |

**주요 에러 타입:**

| Error Code | Description | HTTP Status |
|------------|-------------|-------------|
| `Missing required field` | 필수 필드 누락 (device_type) | 400 |
| `ip_address_required` | IP 주소 필수 (POST 요청 시) | 400 |
| `invalid_ip_address` | 잘못된 IP 주소 형식 | 400 |
| `Device already exists` | 이미 존재하는 디바이스 (POST) | 409 |
| `Device not found` | 디바이스를 찾을 수 없음 | 404 |
| `Internal server error` | 서버 내부 오류 | 500 |

---

## Database Schema

```sql
CREATE TABLE devices (
    device_id TEXT PRIMARY KEY,
    device_type TEXT NOT NULL,
    port INTEGER DEFAULT 9100,
    reload_port INTEGER DEFAULT 9101,
    enabled_metrics TEXT,    -- JSON array
    extra_config TEXT,       -- JSON object
    ip_address TEXT,         -- User-provided device IP address
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8081 | 서버 포트 |
| DB_PATH | ./config.db | SQLite 데이터베이스 경로 |

---

## Running the Server

```bash
# 기본 실행
./edge-metrics-server

# 환경변수 지정
PORT=8080 DB_PATH=/data/config.db ./edge-metrics-server
```

---

## Grafana Dashboards

Edge Metrics Server는 수집된 메트릭을 시각화하기 위한 Grafana 대시보드를 제공합니다.

### 1. Edge Devices Power & Energy Monitoring
**파일**: `manifests/grafana-dashboard.json`

전체 edge 디바이스의 전력 및 에너지 모니터링 대시보드입니다.

**주요 패널:**
- 실시간 전력 메트릭 (모든 전력 메트릭)
- 디바이스별 전력 순위 (Boxplot)
- 디바이스별 전력 사용 세부 내역 (Pie Chart)

### 2. Jetson Power Analysis (이기종 디바이스)
**파일**: `manifests/grafana-dashboard-jetson-power.json`
**UID**: `jetson-power-analysis`

Jetson Nano, Xavier, Orin 이기종 디바이스의 전력 분석 전용 대시보드입니다.

**주요 패널:**
1. **디바이스 간 전력 비교** (외부 플러그 기준)
   - Shelly 플러그로 측정된 보드 전체 전력을 기준으로 공정하게 비교
   - 모든 선택된 디바이스를 한 그래프에 표시

2. **전력 비교 (내부 vs 외부)** (Repeat Panel by hostname)
   - 내부 Total Power: 모델별 자동 선택
     - Nano: `pom_5v_in_watts`
     - Xavier: `vdd_in_watts`
     - Orin: 4개 레일 합산 (`vdd_cpu_cv + vdd_gpu_soc + vddq_vdd2_1v8ao + vin_sys_5v0`)
   - 외부 Shelly: 실제 보드 전체 전력

3. **내부 레일 분해** (Repeat Panel by hostname)
   - Stacked Area 그래프로 각 디바이스의 모든 내부 전력 레일 표시
   - 모델별로 존재하는 레일만 자동 표시

4. **Unaccounted Power** (Repeat Panel by hostname)
   - `(Shelly - 내부 Total) / Shelly * 100` 비율(%) 표시
   - 색상: Green → Yellow → Orange → Red (비율에 따라)
   - 전압변환 손실, 주변기기, 미측정 레일 등의 전력 파악

**Variables:**
- `device_type`: 디바이스 타입 필터 (Multi-select)
- `hostname`: 호스트명 필터 (Multi-select)

**특징:**
- Orin의 Total Power는 4개 레일 합산으로 자동 계산
- PromQL `or` 연산자로 모델별 메트릭 자동 선택
- Repeat Panel로 디바이스별 상세 분석 자동 생성

---