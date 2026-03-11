# Edge Metrics Server API Specification

## Overview

Edge Metrics Server is a centralized configuration management server for edge-metrics-exporter clients.

- **Base URL**: `http://localhost:8081`
- **Content-Type**: `application/json`

## Authentication

When the `API_KEY` environment variable is set, all endpoints require the `X-API-Key` header.

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8081/config
```

If `API_KEY` is not set (default), authentication is disabled and no header is required.

**Response (401 Unauthorized)**
```json
{
  "error": "Unauthorized"
}
```

---

## Endpoints

### GET /config

List all device configurations.

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
      "port": 9102,
      "reload_port": 9101,
      "enabled_metrics": ["jetson_power_vdd_gpu_soc_watts"],
      "jetson": {"use_tegrastats": true}
    },
    {
      "device_id": "edge-02",
      "device_type": "raspberry_pi",
      "port": 9102,
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

Get configuration for a specific device.

**Request**
```
GET /config/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname (e.g., `edge-01`, `orin-desktop`) |

**Response (200 OK)**
```json
{
  "device_type": "jetson_orin",
  "port": 9102,
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

Create or update a device configuration (upsert).

**Request**
```
PUT /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "155.230.34.203",
  "port": 9102,
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
| device_type | string | **Yes** | - | Device type |
| ip_address | string | No | keep existing | Device IP address (keeps existing IP if omitted) |
| port | integer | No | 9102 | Prometheus metrics server port |
| reload_port | integer | No | 9101 | Config reload trigger port |
| enabled_metrics | array | No | null | Metrics to collect (null = all) |
| * | object | No | - | Device-specific extra config (shelly, jetson, etc.) |

**Response (200 OK) — new device registered**
```json
{
  "status": "registered",
  "device_id": "orin-desktop",
  "reload_triggered": false
}
```

**Response (200 OK) — existing device updated**
```json
{
  "status": "updated",
  "device_id": "edge-01",
  "reload_triggered": true
}
```

> **Note**: When `reload_triggered` is `true`, the exporter's `/reload` endpoint was called and the configuration is applied immediately.

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

**Example — register new device**
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

Register a new device. Returns 409 Conflict if the device already exists.

**Request**
```
POST /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "155.230.34.203",
  "port": 9102,
  "reload_port": 9101
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| device_type | string | **Yes** | - | Device type |
| ip_address | string | **Yes** | - | Device IP address |
| port | integer | No | 9102 | Prometheus metrics server port |
| reload_port | integer | No | 9101 | Config reload trigger port |
| enabled_metrics | array | No | null | Metrics to collect |
| * | object | No | - | Device-specific extra config (shelly, jetson, etc.) |

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

Partially update a device configuration. Only provided fields are changed.

**Request**
```
PATCH /config/{device_id}
Content-Type: application/json
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Request Body**
```json
{
  "port": 9200
}
```

Or to change IP address:
```json
{
  "ip_address": "155.230.34.210"
}
```

> Only include the fields you want to change. Passing `null` resets a field to its default or removes it. (Exception: `ip_address` with `null` keeps the existing IP.)

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
# Change port
curl -X PATCH http://localhost:8081/config/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"port": 9200}'

# Change IP address
curl -X PATCH http://localhost:8081/config/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "155.230.34.210"}'
```

---

### DELETE /config/{device_id}

Delete a device configuration.

**Request**
```
DELETE /config/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

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

Check server health status.

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

List all registered devices and their status.

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
      "port": 9102,
      "reload_port": 9101,
      "status": "healthy",
      "last_seen": "2024-01-15T10:30:00Z"
    },
    {
      "device_id": "edge-02",
      "device_type": "jetson_xavier",
      "ip_address": "192.168.1.11",
      "port": 9102,
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
| devices | array | Device status list |
| total | integer | Total device count |
| healthy | integer | Healthy device count |
| unhealthy | integer | Unhealthy device count |

**Device Status Fields**

| Field | Type | Description |
|-------|------|-------------|
| device_id | string | Device ID |
| device_type | string | Device type |
| ip_address | string | Device IP address |
| port | integer | Metrics server port |
| reload_port | integer | Reload trigger port |
| status | string | healthy, unhealthy, unreachable, unknown |
| last_seen | string | Last response time (when healthy) |
| error | string | Error message (when unhealthy) |

**Example**
```bash
curl http://localhost:8081/devices
```

---

### GET /devices/{device_id}/status

Get status for a specific device.

**Request**
```
GET /devices/{device_id}/status
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "device_type": "jetson_orin",
  "ip_address": "192.168.1.10",
  "port": 9102,
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

Update device basic info only (device_type, ip_address, port, reload_port).
This API updates the database only and does **not** trigger a device reload.

**Modifiable fields**: device_type, ip_address, port, reload_port
**Not modifiable**: enabled_metrics, extra_config (jetson, shelly, etc.)

**Request**
```
PATCH /devices/{device_id}
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Request Body**
```json
{
  "device_type": "jetson_orin",
  "ip_address": "192.168.1.20",
  "port": 9102,
  "reload_port": 9101
}
```

All fields are optional. Only provided fields are updated.

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

**Response (400 Bad Request)**
```json
{
  "error": "invalid_ip_address",
  "message": "Invalid IP address format: 192.168.1.999"
}
```

**Example**
```bash
# Change IP only
curl -X PATCH http://localhost:8081/devices/edge-01 \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "192.168.1.20"}'

# Change multiple fields
curl -X PATCH http://localhost:8081/devices/edge-01 \
  -H "Content-Type: application/json" \
  -d '{
    "device_type": "jetson_orin",
    "ip_address": "192.168.1.25",
    "port": 9102,
    "reload_port": 9101
  }'
```

**Difference from PATCH /config/{device_id}**:
- `PATCH /config/{device_id}`: All fields modifiable, triggers reload
- `PATCH /devices/{device_id}`: Basic fields only, no reload

---

### GET /devices/{device_id}/local-config

Get a device's local config.yaml content.
The server proxies the device's `GET :9101/config` endpoint to resolve CORS issues.

**Request**
```
GET /devices/{device_id}/local-config
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

**Response (200 OK)**
```json
{
  "device_type": "jetson_orin",
  "port": 9102,
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

Manually trigger a reload on a specific device.

**Request**
```
POST /devices/{device_id}/reload
```

| Parameter | Type | Location | Description |
|-----------|------|----------|-------------|
| device_id | string | path | Device hostname |

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

Trigger a reload on all devices.

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

Get system-wide summary statistics.

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

Get overall Kubernetes sync status.

**Request**
```
GET /kubernetes/status?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | Target namespace |

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

Check Kubernetes connectivity and RBAC permissions.

**Request**
```
GET /kubernetes/health?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | Target namespace |

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

Sync all healthy devices to Kubernetes as Service + Endpoints.

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
| namespace | string | No | monitoring | Target Kubernetes namespace |

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

**Behavior:**
1. Queries healthy device list via GET /devices
2. For each device, creates/updates Service + Endpoints
   - Service name: `edge-device-{device_id}`
   - Endpoints IP: device's `ip_address`
   - Port: device's `port` (default 9102)
   - Labels: `app=edge-exporter`, `device_id`, `device_type`, `managed_by=edge-metrics-server`
3. Deletes resources for devices that are unhealthy or removed from DB
4. Returns results

---

### POST /kubernetes/sync/{device_id}

Sync a specific device to Kubernetes.

**Request**
```
POST /kubernetes/sync/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | Device ID to sync |
| namespace | string | query | monitoring | Target namespace |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "created"
}
```

**Response (200 OK — failed)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "failed",
  "error": "device is not healthy"
}
```

**Example**
```bash
curl -X POST http://localhost:8081/kubernetes/sync/edge-01?namespace=monitoring
```

---

### GET /kubernetes/manifests

Generate Kubernetes YAML manifests for healthy devices (for manual apply).

**Request**
```
GET /kubernetes/manifests?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | Target namespace |

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
    port: 9102
    targetPort: 9102
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
    port: 9102
    protocol: TCP
```

**Example**
```bash
# Generate and save
curl http://localhost:8081/kubernetes/manifests?namespace=monitoring > edge-devices.yaml

# Apply to Kubernetes
kubectl apply -f edge-devices.yaml
```

**Behavior:**
1. Queries all device configurations
2. Runs health check for each device
3. Generates YAML manifests for healthy devices only
4. Returns as text/plain

---

### GET /kubernetes/resources/{device_id}

Get detailed Kubernetes resource info for a specific device.

**Request**
```
GET /kubernetes/resources/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | Device ID |
| namespace | string | query | monitoring | Target namespace |

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
        "port": 9102
      }
    ]
  },
  "endpoints": {
    "name": "edge-device-edge-01",
    "exists": true,
    "ready_addresses": ["192.168.1.10:9102"],
    "not_ready_addresses": []
  },
  "prometheus_target": "http://edge-device-edge-01.monitoring.svc:9102/metrics"
}
```

**Example**
```bash
curl http://localhost:8081/kubernetes/resources/edge-01?namespace=monitoring
```

---

### DELETE /kubernetes/resources/{device_id}

Delete Kubernetes resources for a specific device.

**Request**
```
DELETE /kubernetes/resources/{device_id}?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| device_id | string | path | - | Device ID |
| namespace | string | query | monitoring | Target namespace |

**Response (200 OK)**
```json
{
  "device_id": "edge-01",
  "service": "edge-device-edge-01",
  "status": "deleted"
}
```

**Example**
```bash
curl -X DELETE http://localhost:8081/kubernetes/resources/edge-01?namespace=monitoring
```

---

### DELETE /kubernetes/cleanup

Delete all edge-device-* resources in the namespace.

**Request**
```
DELETE /kubernetes/cleanup?namespace=monitoring
```

| Parameter | Type | Location | Default | Description |
|-----------|------|----------|---------|-------------|
| namespace | string | query | monitoring | Target namespace |

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

**Example**
```bash
curl -X DELETE http://localhost:8081/kubernetes/cleanup?namespace=monitoring
```

**Behavior:**
1. Lists all Services with `managed_by=edge-metrics-server` label
2. Deletes all matching Services
3. Lists all Endpoints with `managed_by=edge-metrics-server` label
4. Deletes all matching Endpoints
5. Returns deleted resource list

---

## Device Types

Supported device types:

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

All error responses follow this format:

```json
{
  "error": "Error type",
  "device_id": "device-id",
  "message": "Detailed error message"
}
```

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created (POST) |
| 400 | Bad request (missing required field, invalid JSON, invalid IP address) |
| 401 | Unauthorized (invalid or missing API key) |
| 404 | Device not found |
| 409 | Conflict (device already exists) |
| 500 | Internal server error |
| 503 | Service unavailable (Kubernetes client not initialized) |

**Common Error Types:**

| Error Code | Description | HTTP Status |
|------------|-------------|-------------|
| `Unauthorized` | Invalid or missing API key | 401 |
| `Missing required field` | Required field missing (device_type) | 400 |
| `ip_address_required` | IP address required (POST request) | 400 |
| `invalid_ip_address` | Invalid IP address format | 400 |
| `Device already exists` | Device already exists (POST) | 409 |
| `Device not found` | Device not found | 404 |
| `Internal server error` | Internal server error | 500 |

---

## Database Schema

```sql
CREATE TABLE devices (
    device_id TEXT PRIMARY KEY,
    device_type TEXT NOT NULL,
    port INTEGER DEFAULT 9102,
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
| PORT | 8081 | Server port |
| DB_PATH | ./config.db | SQLite database path |
| API_KEY | (empty) | API key for X-API-Key authentication (disabled when empty) |

---

## Running the Server

```bash
# Default
./edge-metrics-server

# With environment variables
PORT=8080 DB_PATH=/data/config.db ./edge-metrics-server

# With API key authentication
API_KEY=my-secret-key ./edge-metrics-server
```

---

## Grafana Dashboards

Edge Metrics Server provides Grafana dashboards for visualizing collected metrics.

### 1. Edge Devices Power & Energy Monitoring
**File**: `manifests/grafana-dashboard.json`

Dashboard for power and energy monitoring across all edge devices.

**Key Panels:**
- Real-time power metrics (all power metrics)
- Device power ranking (Boxplot)
- Per-device power breakdown (Pie Chart)

### 2. Jetson Power Analysis (Heterogeneous Devices)
**File**: `manifests/grafana-dashboard-jetson-power.json`
**UID**: `jetson-power-analysis`

Dashboard for power analysis across heterogeneous Jetson devices (Nano, Xavier, Orin).

**Key Panels:**
1. **Cross-Device Power Comparison** (external plug baseline)
   - Compares total board power measured via Shelly plugs
   - All selected devices on a single graph

2. **Power Comparison (Internal vs External)** (repeat panel by hostname)
   - Internal Total Power: auto-selected per model
     - Nano: `pom_5v_in_watts`
     - Xavier: `vdd_in_watts`
     - Orin: sum of 4 rails (`vdd_cpu_cv + vdd_gpu_soc + vddq_vdd2_1v8ao + vin_sys_5v0`)
   - External Shelly: actual total board power

3. **Internal Rail Breakdown** (repeat panel by hostname)
   - Stacked Area graph showing all internal power rails per device
   - Auto-displays only available rails per model

4. **Unaccounted Power** (repeat panel by hostname)
   - `(Shelly - Internal Total) / Shelly * 100` as percentage
   - Color: Green -> Yellow -> Orange -> Red (by ratio)
   - Identifies voltage conversion losses, peripherals, unmeasured rails

**Variables:**
- `device_type`: Device type filter (multi-select)
- `hostname`: Hostname filter (multi-select)

**Features:**
- Orin Total Power is auto-calculated from 4-rail sum
- PromQL `or` operator for automatic per-model metric selection
- Repeat Panel for auto-generated per-device detail views

---
