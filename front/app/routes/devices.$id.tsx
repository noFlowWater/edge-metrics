import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router';
import { api } from '~/lib/api';
import type { DeviceState, DeviceLocalConfig, KubernetesDeviceResources } from '~/types/device';
import { StatusBadge } from '~/components/StatusBadge';
import EditDeviceModal from '~/components/EditDeviceModal';

export function meta() {
  return [
    { title: 'Device Detail - Edge Metrics' },
    { name: 'description', content: 'Device configuration' },
  ];
}

export default function DeviceDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [configJson, setConfigJson] = useState('');
  const [status, setStatus] = useState<DeviceState | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [k8sResources, setK8sResources] = useState<KubernetesDeviceResources | null>(null);
  const [k8sLoading, setK8sLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);

  // Device Local Config 관련 state
  const [deviceConfig, setDeviceConfig] = useState<DeviceLocalConfig | null>(null);
  const [deviceConfigJson, setDeviceConfigJson] = useState('');
  const [loadingDeviceConfig, setLoadingDeviceConfig] = useState(false);
  const [deviceConfigError, setDeviceConfigError] = useState<string | null>(null);

  // Edit Modal state
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  useEffect(() => {
    async function fetchData() {
      if (!id) return;
      try {
        const [configData, statusData] = await Promise.all([
          api.getConfig(id),
          api.getDeviceStatus(id).catch(() => null),
        ]);
        setConfigJson(JSON.stringify(configData, null, 2));
        setStatus(statusData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load device');
      } finally {
        setLoading(false);
      }
    }

    async function fetchK8sResources() {
      if (!id) return;
      try {
        const k8sData = await api.getDeviceKubernetesResources(id);
        setK8sResources(k8sData);
      } catch (err) {
        // K8s may not be available (503), this is acceptable
        setK8sResources(null);
      } finally {
        setK8sLoading(false);
      }
    }

    fetchData();
    fetchK8sResources();
  }, [id]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!id || !deviceConfigJson) {
      alert('No device config loaded');
      return;
    }
    if (jsonError) return;

    setSaving(true);
    try {
      const config = JSON.parse(deviceConfigJson);
      await api.updateConfig(id, config);
      // 서버 config 다시 로드
      const updatedConfig = await api.getConfig(id);
      setConfigJson(JSON.stringify(updatedConfig, null, 2));
      alert('Configuration saved successfully');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save configuration');
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!id || !confirm('Are you sure you want to delete this device?')) return;

    try {
      await api.deleteConfig(id);
      navigate('/devices');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete');
    }
  }

  async function handleReload() {
    if (!id) return;
    try {
      await api.reloadDevice(id);
      alert('Reload triggered');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Reload failed');
    }
  }

  async function handleK8sSync() {
    if (!id) return;
    setSyncing(true);
    try {
      await api.syncDevice(id);
      const k8sData = await api.getDeviceKubernetesResources(id);
      setK8sResources(k8sData);
      alert('Kubernetes sync completed');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Kubernetes sync failed');
    } finally {
      setSyncing(false);
    }
  }

  async function handleK8sDelete() {
    if (!id || !confirm('Are you sure you want to delete Kubernetes resources for this device?')) return;
    try {
      await api.deleteDeviceKubernetesResources(id);
      setK8sResources(null);
      alert('Kubernetes resources deleted');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete Kubernetes resources');
    }
  }

  async function handleLoadFromDevice() {
    if (!id) return;

    setLoadingDeviceConfig(true);
    setDeviceConfigError(null);

    try {
      const localConfig = await api.getDeviceLocalConfig(id);
      setDeviceConfig(localConfig);
      setDeviceConfigJson(JSON.stringify(localConfig, null, 2));
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to load device config';
      setDeviceConfigError(errorMsg);
      alert(errorMsg);
    } finally {
      setLoadingDeviceConfig(false);
    }
  }

  async function handleEditSave(data: {
    device_type: string;
    ip_address: string;
    port: number;
    reload_port: number;
  }) {
    if (!id) return;

    await api.patchDevice(id, data);

    // Refresh device status and config
    const [configData, statusData] = await Promise.all([
      api.getConfig(id),
      api.getDeviceStatus(id),
    ]);
    setConfigJson(JSON.stringify(configData, null, 2));
    setStatus(statusData);

    alert('Device updated successfully (reload not triggered)');
  }

  if (loading) {
    return <div className="text-center py-8" style={{ color: 'var(--color-pastel-text-light)' }}>Loading...</div>;
  }

  if (error) {
    return (
      <div className="rounded-xl p-4" style={{ backgroundColor: 'var(--color-pastel-danger-bg)', border: '1px solid var(--color-pastel-danger-light)' }}>
        <p style={{ color: 'var(--color-pastel-danger)' }}>{error}</p>
        <Link to="/devices" className="text-sm hover:opacity-80 mt-2 inline-block" style={{ color: 'var(--color-pastel-primary)' }}>
          Back to devices
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-4xl space-y-8">
      <div className="flex flex-col lg:flex-row lg:justify-between lg:items-start gap-6">
        <div>
          <Link to="/devices" className="text-sm hover:opacity-80 transition-opacity" style={{ color: 'var(--color-pastel-text-muted)' }}>
            ← Back to devices
          </Link>
          <h1 className="text-3xl lg:text-4xl font-bold mt-3" style={{ color: 'var(--color-pastel-text)' }}>{id}</h1>
        </div>
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
          <button
            onClick={() => setIsEditModalOpen(true)}
            className="px-6 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium whitespace-nowrap"
            style={{
              backgroundColor: 'var(--color-pastel-primary)',
              color: 'white'
            }}
          >
            Edit
          </button>
          <button
            onClick={handleReload}
            className="px-6 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium whitespace-nowrap"
            style={{
              backgroundColor: 'var(--color-pastel-card)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text-light)'
            }}
          >
            Reload
          </button>
          <button
            onClick={handleDelete}
            className="px-6 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium whitespace-nowrap"
            style={{
              backgroundColor: 'var(--color-pastel-danger)',
              color: 'white'
            }}
          >
            Delete
          </button>
        </div>
      </div>

      {/* Status */}
      {status && (
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <h2 className="text-lg sm:text-xl font-semibold mb-6" style={{ color: 'var(--color-pastel-text)' }}>Status</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Status</p>
              <StatusBadge status={status.status} />
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>IP Address</p>
              <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>{status.ip_address}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Port</p>
              <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>{status.port}</p>
            </div>
            {status.last_seen && (
              <div>
                <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Last Seen</p>
                <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>
                  {new Date(status.last_seen).toLocaleString()}
                </p>
              </div>
            )}
          </div>

          {/* Metrics Endpoint Link */}
          <div className="mt-6 pt-6" style={{ borderTop: '1px solid var(--color-pastel-border)' }}>
            <a
              href={`http://${status.ip_address}:${status.port}/metrics`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-5 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium text-sm"
              style={{
                backgroundColor: 'var(--color-pastel-primary-lighter)',
                color: 'var(--color-pastel-primary)'
              }}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              View Metrics Endpoint
            </a>
          </div>
        </div>
      )}

      {/* Kubernetes Resources */}
      {!k8sLoading && k8sResources && (
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3 mb-6">
            <h2 className="text-lg sm:text-xl font-semibold" style={{ color: 'var(--color-pastel-text)' }}>Kubernetes Resources</h2>
            <div className="flex flex-wrap gap-3">
              <button
                onClick={handleK8sSync}
                disabled={syncing}
                className="px-5 py-2 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium text-sm"
                style={{
                  backgroundColor: 'var(--color-pastel-primary-lighter)',
                  color: 'var(--color-pastel-primary)'
                }}
              >
                {syncing ? 'Syncing...' : 'Sync to K8s'}
              </button>
              <button
                onClick={handleK8sDelete}
                className="px-5 py-2 rounded-lg hover:opacity-90 transition-opacity font-medium text-sm"
                style={{
                  backgroundColor: 'var(--color-pastel-danger-bg)',
                  border: '1px solid var(--color-pastel-danger-light)',
                  color: 'var(--color-pastel-danger)'
                }}
              >
                Delete K8s Resources
              </button>
            </div>
          </div>

          <div className="space-y-5">
            {/* Service */}
            <div>
              <div className="flex items-center gap-3 mb-3">
                <p className="text-xs uppercase tracking-wide" style={{ color: 'var(--color-pastel-text-muted)' }}>Service</p>
                <span className={`text-xs px-2 py-1 rounded ${k8sResources.service.exists ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                  {k8sResources.service.exists ? 'EXISTS' : 'NOT FOUND'}
                </span>
              </div>
              {k8sResources.service.exists && (
                <div className="pl-4 space-y-2 text-sm" style={{ color: 'var(--color-pastel-text)' }}>
                  <p><strong>Name:</strong> {k8sResources.service.name}</p>
                  {k8sResources.service.cluster_ip && <p><strong>Cluster IP:</strong> {k8sResources.service.cluster_ip}</p>}
                  {k8sResources.service.ports && k8sResources.service.ports.length > 0 && (
                    <p><strong>Ports:</strong> {k8sResources.service.ports.map(p => `${p.name}:${p.port}`).join(', ')}</p>
                  )}
                </div>
              )}
            </div>

            {/* Endpoints */}
            <div>
              <div className="flex items-center gap-3 mb-3">
                <p className="text-xs uppercase tracking-wide" style={{ color: 'var(--color-pastel-text-muted)' }}>Endpoints</p>
                <span className={`text-xs px-2 py-1 rounded ${k8sResources.endpoints.exists ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                  {k8sResources.endpoints.exists ? 'EXISTS' : 'NOT FOUND'}
                </span>
              </div>
              {k8sResources.endpoints.exists && (
                <div className="pl-4 space-y-2 text-sm" style={{ color: 'var(--color-pastel-text)' }}>
                  <p><strong>Name:</strong> {k8sResources.endpoints.name}</p>
                  {k8sResources.endpoints.ready_addresses && k8sResources.endpoints.ready_addresses.length > 0 && (
                    <p><strong>Ready Addresses:</strong> {k8sResources.endpoints.ready_addresses.join(', ')}</p>
                  )}
                  {k8sResources.endpoints.not_ready_addresses && k8sResources.endpoints.not_ready_addresses.length > 0 && (
                    <p style={{ color: 'var(--color-pastel-danger)' }}>
                      <strong>Not Ready:</strong> {k8sResources.endpoints.not_ready_addresses.join(', ')}
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Prometheus Target */}
            {k8sResources.prometheus_target && (
              <div className="pt-4" style={{ borderTop: '1px solid var(--color-pastel-border)' }}>
                <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Prometheus Target</p>
                <code className="text-sm px-3 py-2 rounded block" style={{ backgroundColor: '#f8f9fc', color: 'var(--color-pastel-text)' }}>
                  {k8sResources.prometheus_target}
                </code>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Configuration Section */}
      <div className="rounded-xl shadow-sm p-6 sm:p-8 space-y-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
        <h2 className="text-lg sm:text-xl font-semibold" style={{ color: 'var(--color-pastel-text)' }}>Configuration</h2>

        {/* Server Configuration (읽기 전용) */}
        <div>
          <h3 className="text-base font-semibold mb-3" style={{ color: 'var(--color-pastel-text)' }}>
            Server Configuration (Database)
          </h3>
          <div
            className="rounded-lg p-4 font-mono text-sm overflow-auto"
            style={{
              backgroundColor: '#f8f9fc',
              border: '1px solid var(--color-pastel-border)',
              maxHeight: '300px'
            }}
          >
            <pre style={{ color: 'var(--color-pastel-text)' }}>{configJson}</pre>
          </div>
          <p className="text-sm mt-2" style={{ color: 'var(--color-pastel-text-muted)' }}>
            This is the configuration stored in the server database
          </p>
        </div>

        {/* Device Local Configuration (편집 가능) */}
        <div>
          <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3 mb-3">
            <h3 className="text-base font-semibold" style={{ color: 'var(--color-pastel-text)' }}>
              Device Local Configuration
            </h3>
            <button
              onClick={handleLoadFromDevice}
              disabled={loadingDeviceConfig}
              className="px-4 py-2 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium text-sm whitespace-nowrap"
              style={{
                backgroundColor: 'var(--color-pastel-primary)',
                color: 'white'
              }}
            >
              {loadingDeviceConfig ? 'Loading...' : 'Load from Device'}
            </button>
          </div>

          {deviceConfigError && (
            <div
              className="p-3 rounded-lg mb-3"
              style={{
                backgroundColor: 'var(--color-pastel-danger-bg)',
                color: 'var(--color-pastel-danger)'
              }}
            >
              {deviceConfigError}
            </div>
          )}

          {deviceConfig ? (
            <form onSubmit={handleSubmit}>
              <textarea
                value={deviceConfigJson}
                onChange={(e) => {
                  setDeviceConfigJson(e.target.value);
                  try {
                    JSON.parse(e.target.value);
                    setJsonError(null);
                  } catch {
                    setJsonError('Invalid JSON');
                  }
                }}
                className="block w-full h-80 sm:h-96 rounded-lg px-4 sm:px-5 py-3 sm:py-4 font-mono text-sm focus:outline-none resize-y"
                style={{
                  backgroundColor: '#f8f9fc',
                  border: '1px solid var(--color-pastel-border)',
                  color: 'var(--color-pastel-text)'
                }}
                spellCheck={false}
              />
              {jsonError && (
                <p className="text-sm mt-2" style={{ color: 'var(--color-pastel-danger)' }}>
                  {jsonError}
                </p>
              )}
              <div className="flex justify-end gap-3 mt-4">
                <button
                  type="submit"
                  disabled={saving || !!jsonError}
                  className="px-6 py-3 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium"
                  style={{
                    backgroundColor: 'var(--color-pastel-primary)',
                    color: 'white'
                  }}
                >
                  {saving ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          ) : (
            <div
              className="p-8 text-center rounded-lg"
              style={{
                backgroundColor: '#f8f9fc',
                border: '1px solid var(--color-pastel-border)'
              }}
            >
              <p style={{ color: 'var(--color-pastel-text-muted)' }}>
                Click "Load from Device" to fetch the actual config from the edge device
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Edit Device Modal */}
      {status && (
        <EditDeviceModal
          isOpen={isEditModalOpen}
          onClose={() => setIsEditModalOpen(false)}
          device={status}
          onSave={handleEditSave}
        />
      )}
    </div>
  );
}
