import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { api } from '~/lib/api';
import type { DevicesResponse, DeviceStatus, DeviceType } from '~/types/device';
import { StatusBadge } from '~/components/StatusBadge';

export function meta() {
  return [
    { title: 'Devices - Edge Metrics' },
    { name: 'description', content: 'Manage edge devices' },
  ];
}

export default function Devices() {
  const [devices, setDevices] = useState<DevicesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloading, setReloading] = useState<string | null>(null);
  const [filter, setFilter] = useState<{ status: string; type: string }>({
    status: 'all',
    type: 'all',
  });

  async function fetchDevices() {
    try {
      const data = await api.getDevices();
      setDevices(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load devices');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchDevices();
  }, []);

  async function handleReload(deviceId: string) {
    setReloading(deviceId);
    try {
      await api.reloadDevice(deviceId);
      await fetchDevices();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Reload failed');
    } finally {
      setReloading(null);
    }
  }

  async function handleReloadAll() {
    setReloading('all');
    try {
      const result = await api.reloadAllDevices();
      alert(`Reload complete: ${result.success} success, ${result.failed} failed`);
      await fetchDevices();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Reload failed');
    } finally {
      setReloading(null);
    }
  }

  if (loading) {
    return <div className="text-center py-8" style={{ color: 'var(--color-pastel-text-light)' }}>Loading...</div>;
  }

  if (error) {
    return (
      <div className="rounded-xl p-4" style={{ backgroundColor: 'var(--color-pastel-danger-bg)', border: '1px solid var(--color-pastel-danger-light)' }}>
        <p style={{ color: 'var(--color-pastel-danger)' }}>{error}</p>
      </div>
    );
  }

  const filteredDevices = devices?.devices.filter((device) => {
    if (filter.status !== 'all' && device.status !== filter.status) return false;
    if (filter.type !== 'all' && device.device_type !== filter.type) return false;
    return true;
  });

  const deviceTypes = [...new Set(devices?.devices?.map((d) => d.device_type) || [])];

  return (
    <div className="space-y-8">
      <div className="flex flex-col lg:flex-row lg:justify-between lg:items-center gap-6">
        <div>
          <h1 className="text-3xl lg:text-4xl font-bold" style={{ color: 'var(--color-pastel-text)' }}>Devices</h1>
          <p className="mt-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Manage your edge devices</p>
        </div>
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
          <button
            onClick={handleReloadAll}
            disabled={reloading === 'all'}
            className="px-6 py-3 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium whitespace-nowrap"
            style={{
              backgroundColor: 'var(--color-pastel-card)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text-light)'
            }}
          >
            {reloading === 'all' ? 'Reloading...' : 'Reload All'}
          </button>
          <Link
            to="/devices/new"
            className="px-6 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium text-center whitespace-nowrap"
            style={{
              backgroundColor: 'var(--color-pastel-primary)',
              color: 'white'
            }}
          >
            Add Device
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <select
          value={filter.status}
          onChange={(e) => setFilter({ ...filter, status: e.target.value })}
          className="flex-1 sm:flex-none rounded-lg px-5 py-3 focus:outline-none transition-all font-medium text-sm"
          style={{
            backgroundColor: 'var(--color-pastel-card)',
            border: '1px solid var(--color-pastel-border)',
            color: 'var(--color-pastel-text)'
          }}
        >
          <option value="all">All Status</option>
          <option value="healthy">Healthy</option>
          <option value="unhealthy">Unhealthy</option>
          <option value="unreachable">Unreachable</option>
        </select>
        <select
          value={filter.type}
          onChange={(e) => setFilter({ ...filter, type: e.target.value })}
          className="flex-1 sm:flex-none rounded-lg px-5 py-3 focus:outline-none transition-all font-medium text-sm"
          style={{
            backgroundColor: 'var(--color-pastel-card)',
            border: '1px solid var(--color-pastel-border)',
            color: 'var(--color-pastel-text)'
          }}
        >
          <option value="all">All Types</option>
          {deviceTypes.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
      </div>

      {/* Device Cards */}
      <div className="grid grid-cols-1 gap-6">
        {filteredDevices?.map((device) => (
          <Link
            key={device.device_id}
            to={`/devices/${device.device_id}`}
            className="rounded-xl shadow-sm p-6 sm:p-8 hover:shadow-md transition-all"
            style={{
              backgroundColor: 'var(--color-pastel-card)',
              border: '1px solid var(--color-pastel-border)'
            }}
          >
            <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-3 mb-4">
                  <h3 className="text-lg sm:text-xl font-semibold" style={{ color: 'var(--color-pastel-primary)' }}>
                    {device.device_id}
                  </h3>
                  <StatusBadge status={device.status} />
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                  <div>
                    <p className="text-xs uppercase tracking-wide mb-2" style={{ color: 'var(--color-pastel-text-muted)' }}>
                      Type
                    </p>
                    <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>
                      {device.device_type}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-wide mb-2" style={{ color: 'var(--color-pastel-text-muted)' }}>
                      IP Address
                    </p>
                    <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>
                      {device.ip_address}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-wide mb-2" style={{ color: 'var(--color-pastel-text-muted)' }}>
                      Port
                    </p>
                    <p className="text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>
                      {device.port}
                    </p>
                  </div>
                </div>
              </div>
              <button
                onClick={(e) => {
                  e.preventDefault();
                  handleReload(device.device_id);
                }}
                disabled={reloading === device.device_id}
                className="w-full lg:w-auto lg:ml-4 px-5 py-3 rounded-lg hover:opacity-80 disabled:opacity-50 transition-opacity text-sm font-medium whitespace-nowrap"
                style={{
                  backgroundColor: 'var(--color-pastel-primary-lighter)',
                  color: 'var(--color-pastel-primary)'
                }}
              >
                {reloading === device.device_id ? 'Reloading...' : 'Reload'}
              </button>
            </div>
          </Link>
        ))}
        {(!filteredDevices || filteredDevices.length === 0) && (
          <div className="text-center py-16 rounded-xl" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
            <p style={{ color: 'var(--color-pastel-text-muted)' }}>No devices found</p>
          </div>
        )}
      </div>
    </div>
  );
}
