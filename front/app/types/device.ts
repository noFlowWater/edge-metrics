export interface DeviceConfig {
  device_id: string;
  device_type: DeviceType;
  port: number;
  reload_port: number;
  [key: string]: any; // Allow dynamic fields
}

// Device Local Config (exporter에서 반환하는 형식)
export interface DeviceLocalConfig {
  device_type: string;
  interval?: number;
  port: number;
  reload_port: number;
  metrics?: Record<string, boolean>;
  shelly?: any;
  jetson?: any;
  ina260?: any;
  [key: string]: any; // Allow additional fields
}

export type DeviceType =
  | 'jetson_orin'
  | 'jetson_xavier'
  | 'jetson_nano'  
  | 'jetson'
  | 'raspberry_pi'
  | 'orange_pi'
  | 'lattepanda'
  | 'shelly';

export type DeviceStatus = 'healthy' | 'unhealthy' | 'unreachable' | 'unknown';

export interface DeviceState {
  device_id: string;
  device_type: DeviceType;
  ip_address: string;
  port: number;
  reload_port: number;
  status: DeviceStatus;
  last_seen?: string;
  error?: string;
}

export interface DevicesResponse {
  devices: DeviceState[];
  total: number;
  healthy: number;
  unhealthy: number;
}

export interface ConfigsResponse {
  configs: DeviceConfig[];
  total: number;
}

export interface MetricsSummary {
  total: number;
  healthy: number;
  unhealthy: number;
  by_device_type: Record<string, number>;
}

export interface HealthResponse {
  status: string;
  service: string;
  version: string;
}

export interface ReloadResult {
  device_id: string;
  status: string;
  error?: string;
}

export interface BulkReloadResponse {
  results: ReloadResult[];
  total: number;
  success: number;
  failed: number;
}

// Kubernetes types
export interface KubernetesResourceStatus {
  device_id: string;
  service_exists: boolean;
  endpoints_exists: boolean;
}

export interface KubernetesStatus {
  kubernetes_enabled: boolean;
  namespace: string;
  total_k8s_resources: number;
  total_registered_devices: number;
  synced: number;
  unsynced: number;
  resources: KubernetesResourceStatus[];
}

export interface KubernetesHealth {
  kubernetes_available: boolean;
  client_initialized: boolean;
  namespace_accessible: boolean;
  rbac_permissions: {
    namespace?: string;
    services?: string;
    endpoints?: string;
  };
}

export interface SyncResult {
  device_id: string;
  service: string;
  status: string;
  error?: string;
}

export interface KubernetesSyncResponse {
  status: string;
  created: SyncResult[];
  updated: SyncResult[];
  deleted: SyncResult[];
  failed: SyncResult[];
  total_healthy: number;
}

export interface KubernetesServiceInfo {
  name: string;
  exists: boolean;
  cluster_ip?: string;
  ports?: Array<{ name: string; port: number }>;
}

export interface KubernetesEndpointsInfo {
  name: string;
  exists: boolean;
  ready_addresses?: string[];
  not_ready_addresses?: string[];
}

export interface KubernetesDeviceResources {
  device_id: string;
  service: KubernetesServiceInfo;
  endpoints: KubernetesEndpointsInfo;
  prometheus_target?: string;
}

export interface KubernetesCleanupResponse {
  status: string;
  deleted_services: string[];
  deleted_endpoints: string[];
  namespace: string;
}
