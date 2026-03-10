import type {
  DeviceConfig,
  DeviceLocalConfig,
  DevicesResponse,
  ConfigsResponse,
  MetricsSummary,
  HealthResponse,
  DeviceState,
  BulkReloadResponse,
  KubernetesStatus,
  KubernetesHealth,
  KubernetesSyncResponse,
  KubernetesDeviceResources,
  KubernetesCleanupResponse,
  SyncResult,
} from '~/types/device';

// 환경변수에서 API URL 가져오기 (Kubernetes 배포 시 설정)
// 개발 환경: vite.config.ts의 proxy 사용 (/api -> localhost:8081)
// 프로덕션: 환경변수 API_URL 또는 기본값 /api 사용
const API_BASE = typeof process !== 'undefined' && process.env.API_URL
  ? process.env.API_URL
  : '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `Request failed: ${response.status}`);
  }

  return response.json();
}

export const api = {
  // Health
  health: () => request<HealthResponse>('/health'),

  // Configs
  getConfigs: () => request<ConfigsResponse>('/config'),
  getConfig: (deviceId: string) => request<DeviceConfig>(`/config/${deviceId}`),
  createConfig: (deviceId: string, config: Omit<DeviceConfig, 'device_id'>) =>
    request<{ status: string; device_id: string }>(`/config/${deviceId}`, {
      method: 'POST',
      body: JSON.stringify(config),
    }),
  updateConfig: (deviceId: string, config: Omit<DeviceConfig, 'device_id'>) =>
    request<{ status: string; device_id: string; reload_triggered: boolean }>(
      `/config/${deviceId}`,
      {
        method: 'PUT',
        body: JSON.stringify(config),
      }
    ),
  patchConfig: (deviceId: string, config: Partial<DeviceConfig>) =>
    request<{ status: string; device_id: string; reload_triggered: boolean }>(
      `/config/${deviceId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(config),
      }
    ),
  deleteConfig: (deviceId: string) =>
    request<{ status: string; device_id: string }>(`/config/${deviceId}`, {
      method: 'DELETE',
    }),

  // Devices
  getDevices: () => request<DevicesResponse>('/devices'),
  getDeviceStatus: (deviceId: string) =>
    request<DeviceState>(`/devices/${deviceId}/status`),
  patchDevice: (
    deviceId: string,
    data: {
      device_type?: string;
      ip_address?: string;
      port?: number;
      reload_port?: number;
    }
  ) =>
    request<{ status: string; device_id: string; message: string }>(
      `/devices/${deviceId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      }
    ),
  reloadDevice: (deviceId: string) =>
    request<{ status: string; device_id: string; error?: string }>(
      `/devices/${deviceId}/reload`,
      { method: 'POST' }
    ),
  reloadAllDevices: () =>
    request<BulkReloadResponse>('/devices/reload', { method: 'POST' }),

  // Device Local Config (서버를 통해 프록시로 가져오기 - CORS 회피)
  getDeviceLocalConfig: (deviceId: string) =>
    request<DeviceLocalConfig>(`/devices/${deviceId}/local-config`),

  // Metrics
  getMetricsSummary: () => request<MetricsSummary>('/metrics/summary'),

  // Kubernetes
  getKubernetesStatus: (namespace: string = 'monitoring') =>
    request<KubernetesStatus>(`/kubernetes/status?namespace=${namespace}`),

  getKubernetesHealth: (namespace: string = 'monitoring') =>
    request<KubernetesHealth>(`/kubernetes/health?namespace=${namespace}`),

  syncAllDevices: (namespace: string = 'monitoring') =>
    request<KubernetesSyncResponse>('/kubernetes/sync', {
      method: 'POST',
      body: JSON.stringify({ namespace }),
    }),

  syncDevice: (deviceId: string, namespace: string = 'monitoring') =>
    request<SyncResult>(`/kubernetes/sync/${deviceId}?namespace=${namespace}`, {
      method: 'POST',
    }),

  getKubernetesManifests: (namespace: string = 'monitoring') =>
    fetch(`${API_BASE}/kubernetes/manifests?namespace=${namespace}`).then(res => {
      if (!res.ok) throw new Error(`Request failed: ${res.status}`);
      return res.text();
    }),

  getDeviceKubernetesResources: (deviceId: string, namespace: string = 'monitoring') =>
    request<KubernetesDeviceResources>(`/kubernetes/resources/${deviceId}?namespace=${namespace}`),

  deleteDeviceKubernetesResources: (deviceId: string, namespace: string = 'monitoring') =>
    request<SyncResult>(`/kubernetes/resources/${deviceId}?namespace=${namespace}`, {
      method: 'DELETE',
    }),

  cleanupKubernetesResources: (namespace: string = 'monitoring') =>
    request<KubernetesCleanupResponse>(`/kubernetes/cleanup?namespace=${namespace}`, {
      method: 'DELETE',
    }),
};
