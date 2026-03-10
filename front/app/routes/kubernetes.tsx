import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { api } from '~/lib/api';
import type { KubernetesStatus, KubernetesHealth } from '~/types/device';

export function meta() {
  return [
    { title: 'Kubernetes - Edge Metrics' },
    { name: 'description', content: 'Manage Kubernetes resources' },
  ];
}

export default function Kubernetes() {
  const [status, setStatus] = useState<KubernetesStatus | null>(null);
  const [health, setHealth] = useState<KubernetesHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [namespace, setNamespace] = useState('monitoring');

  async function fetchData() {
    setLoading(true);
    setError(null);
    try {
      const [statusData, healthData] = await Promise.all([
        api.getKubernetesStatus(namespace).catch(() => null),
        api.getKubernetesHealth(namespace).catch(() => null),
      ]);
      setStatus(statusData);
      setHealth(healthData);

      if (!statusData && !healthData) {
        setError('Kubernetes is not available. Server may not be running in a Kubernetes environment.');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load Kubernetes status');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchData();
  }, [namespace]);

  async function handleSyncAll() {
    setSyncing(true);
    try {
      const result = await api.syncAllDevices(namespace);
      alert(`Sync completed: ${result.created.length} created, ${result.updated.length} updated, ${result.deleted.length} deleted, ${result.failed.length} failed`);
      await fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Sync failed');
    } finally {
      setSyncing(false);
    }
  }

  async function handleCleanup() {
    if (!confirm(`Are you sure you want to delete ALL edge-device-* resources in namespace "${namespace}"?`)) return;

    try {
      const result = await api.cleanupKubernetesResources(namespace);
      alert(`Cleanup completed: ${result.deleted_services.length} services and ${result.deleted_endpoints.length} endpoints deleted`);
      await fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Cleanup failed');
    }
  }

  async function handleDownloadManifests() {
    try {
      const yaml = await api.getKubernetesManifests(namespace);
      const blob = new Blob([yaml], { type: 'text/yaml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `edge-devices-${namespace}.yaml`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to download manifests');
    }
  }

  if (loading) {
    return <div className="text-center py-8" style={{ color: 'var(--color-pastel-text-light)' }}>Loading...</div>;
  }

  if (error && !status && !health) {
    return (
      <div className="space-y-8">
        <h1 className="text-3xl lg:text-4xl font-bold" style={{ color: 'var(--color-pastel-text)' }}>Kubernetes</h1>
        <div className="rounded-xl p-6" style={{ backgroundColor: 'var(--color-pastel-danger-bg)', border: '1px solid var(--color-pastel-danger-light)' }}>
          <p style={{ color: 'var(--color-pastel-danger)' }}>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col lg:flex-row lg:justify-between lg:items-start gap-6">
        <div>
          <h1 className="text-3xl lg:text-4xl font-bold" style={{ color: 'var(--color-pastel-text)' }}>Kubernetes</h1>
          <p className="mt-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Manage Kubernetes resources for edge devices</p>
        </div>
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>Namespace:</label>
          <input
            type="text"
            value={namespace}
            onChange={(e) => setNamespace(e.target.value)}
            className="px-4 py-2 rounded-lg focus:outline-none text-sm"
            style={{
              backgroundColor: 'var(--color-pastel-card)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text)'
            }}
            placeholder="monitoring"
          />
        </div>
      </div>

      {/* Health Status */}
      {health && (
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <h2 className="text-lg sm:text-xl font-semibold mb-6" style={{ color: 'var(--color-pastel-text)' }}>Health Status</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>K8s Available</p>
              <span className={`text-sm px-3 py-1 rounded ${health.kubernetes_available ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                {health.kubernetes_available ? 'Yes' : 'No'}
              </span>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Client Initialized</p>
              <span className={`text-sm px-3 py-1 rounded ${health.client_initialized ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                {health.client_initialized ? 'Yes' : 'No'}
              </span>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Namespace Accessible</p>
              <span className={`text-sm px-3 py-1 rounded ${health.namespace_accessible ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                {health.namespace_accessible ? 'Yes' : 'No'}
              </span>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide mb-3" style={{ color: 'var(--color-pastel-text-muted)' }}>RBAC Permissions</p>
              <div className="text-sm space-y-1" style={{ color: 'var(--color-pastel-text)' }}>
                {health.rbac_permissions.services === 'ok' && health.rbac_permissions.endpoints === 'ok' ? (
                  <span className="px-3 py-1 rounded bg-green-100 text-green-700">OK</span>
                ) : (
                  <span className="px-3 py-1 rounded bg-red-100 text-red-700">Check Failed</span>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Sync Status */}
      {status && (
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
            <h2 className="text-lg sm:text-xl font-semibold" style={{ color: 'var(--color-pastel-text)' }}>Synchronization Status</h2>
            <div className="flex flex-wrap gap-3">
              <button
                onClick={handleSyncAll}
                disabled={syncing}
                className="px-5 py-2 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium text-sm"
                style={{
                  backgroundColor: 'var(--color-pastel-primary)',
                  color: 'white'
                }}
              >
                {syncing ? 'Syncing...' : 'Sync All Devices'}
              </button>
              <button
                onClick={handleDownloadManifests}
                className="px-5 py-2 rounded-lg hover:opacity-90 transition-opacity font-medium text-sm"
                style={{
                  backgroundColor: 'var(--color-pastel-primary-lighter)',
                  color: 'var(--color-pastel-primary)'
                }}
              >
                Download YAML
              </button>
              <button
                onClick={handleCleanup}
                className="px-5 py-2 rounded-lg hover:opacity-90 transition-opacity font-medium text-sm"
                style={{
                  backgroundColor: 'var(--color-pastel-danger)',
                  color: 'white'
                }}
              >
                Cleanup All
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-6 mb-8">
            <div className="text-center p-4 rounded-lg" style={{ backgroundColor: 'var(--color-pastel-bg)' }}>
              <p className="text-2xl font-bold" style={{ color: 'var(--color-pastel-primary)' }}>{status.total_registered_devices}</p>
              <p className="text-xs uppercase tracking-wide mt-2" style={{ color: 'var(--color-pastel-text-muted)' }}>Registered</p>
            </div>
            <div className="text-center p-4 rounded-lg" style={{ backgroundColor: 'var(--color-pastel-bg)' }}>
              <p className="text-2xl font-bold text-green-600">{status.synced}</p>
              <p className="text-xs uppercase tracking-wide mt-2" style={{ color: 'var(--color-pastel-text-muted)' }}>Synced</p>
            </div>
            <div className="text-center p-4 rounded-lg" style={{ backgroundColor: 'var(--color-pastel-bg)' }}>
              <p className="text-2xl font-bold text-orange-600">{status.unsynced}</p>
              <p className="text-xs uppercase tracking-wide mt-2" style={{ color: 'var(--color-pastel-text-muted)' }}>Unsynced</p>
            </div>
            <div className="text-center p-4 rounded-lg" style={{ backgroundColor: 'var(--color-pastel-bg)' }}>
              <p className="text-2xl font-bold" style={{ color: 'var(--color-pastel-text)' }}>{status.total_k8s_resources}</p>
              <p className="text-xs uppercase tracking-wide mt-2" style={{ color: 'var(--color-pastel-text-muted)' }}>K8s Resources</p>
            </div>
          </div>

          {/* Device Resources Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr style={{ borderBottom: '2px solid var(--color-pastel-border)' }}>
                  <th className="text-left py-3 px-4" style={{ color: 'var(--color-pastel-text-muted)' }}>Device ID</th>
                  <th className="text-center py-3 px-4" style={{ color: 'var(--color-pastel-text-muted)' }}>Service</th>
                  <th className="text-center py-3 px-4" style={{ color: 'var(--color-pastel-text-muted)' }}>Endpoints</th>
                  <th className="text-center py-3 px-4" style={{ color: 'var(--color-pastel-text-muted)' }}>Status</th>
                  <th className="text-right py-3 px-4" style={{ color: 'var(--color-pastel-text-muted)' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {status.resources.map((resource) => {
                  const isSynced = resource.service_exists && resource.endpoints_exists;
                  return (
                    <tr key={resource.device_id} style={{ borderBottom: '1px solid var(--color-pastel-border)' }}>
                      <td className="py-4 px-4">
                        <Link
                          to={`/devices/${resource.device_id}`}
                          className="hover:opacity-80 font-medium"
                          style={{ color: 'var(--color-pastel-primary)' }}
                        >
                          {resource.device_id}
                        </Link>
                      </td>
                      <td className="text-center py-4 px-4">
                        <span className={`text-xs px-2 py-1 rounded ${resource.service_exists ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                          {resource.service_exists ? '✓' : '✗'}
                        </span>
                      </td>
                      <td className="text-center py-4 px-4">
                        <span className={`text-xs px-2 py-1 rounded ${resource.endpoints_exists ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                          {resource.endpoints_exists ? '✓' : '✗'}
                        </span>
                      </td>
                      <td className="text-center py-4 px-4">
                        <span className={`text-xs px-3 py-1 rounded font-medium ${isSynced ? 'bg-green-100 text-green-700' : 'bg-orange-100 text-orange-700'}`}>
                          {isSynced ? 'Synced' : 'Unsynced'}
                        </span>
                      </td>
                      <td className="text-right py-4 px-4">
                        <Link
                          to={`/devices/${resource.device_id}`}
                          className="text-xs px-3 py-2 rounded hover:opacity-80 transition-opacity font-medium inline-block"
                          style={{
                            backgroundColor: 'var(--color-pastel-primary-lighter)',
                            color: 'var(--color-pastel-primary)'
                          }}
                        >
                          View
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {status.resources.length === 0 && (
              <div className="text-center py-12" style={{ color: 'var(--color-pastel-text-muted)' }}>
                No devices found
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
