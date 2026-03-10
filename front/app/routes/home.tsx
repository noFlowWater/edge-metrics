import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { api } from '~/lib/api';
import type { MetricsSummary, DevicesResponse } from '~/types/device';
import { StatusBadge } from '~/components/StatusBadge';

export function meta() {
  return [
    { title: 'Dashboard - Edge Metrics' },
    { name: 'description', content: 'Edge Metrics Dashboard' },
  ];
}

export default function Dashboard() {
  const [summary, setSummary] = useState<MetricsSummary | null>(null);
  const [devices, setDevices] = useState<DevicesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        const [summaryData, devicesData] = await Promise.all([
          api.getMetricsSummary(),
          api.getDevices(),
        ]);
        setSummary(summaryData);
        setDevices(devicesData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load data');
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

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

  return (
    <div className="space-y-10">
      <div>
        <h1 className="text-3xl lg:text-4xl font-bold" style={{ color: 'var(--color-pastel-text)' }}>Dashboard</h1>
        <p className="mt-3" style={{ color: 'var(--color-pastel-text-muted)' }}>Edge device monitoring overview</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <p className="text-sm font-medium uppercase tracking-wide" style={{ color: 'var(--color-pastel-text-muted)' }}>Total Devices</p>
          <p className="mt-4 text-4xl sm:text-5xl font-bold" style={{ color: 'var(--color-pastel-primary)' }}>{summary?.total || 0}</p>
        </div>
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <p className="text-sm font-medium uppercase tracking-wide" style={{ color: 'var(--color-pastel-text-muted)' }}>Healthy</p>
          <p className="mt-4 text-4xl sm:text-5xl font-bold" style={{ color: 'var(--color-pastel-success)' }}>{summary?.healthy || 0}</p>
        </div>
        <div className="rounded-xl shadow-sm p-6 sm:p-8 sm:col-span-2 lg:col-span-1" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <p className="text-sm font-medium uppercase tracking-wide" style={{ color: 'var(--color-pastel-text-muted)' }}>Unhealthy</p>
          <p className="mt-4 text-4xl sm:text-5xl font-bold" style={{ color: 'var(--color-pastel-danger)' }}>{summary?.unhealthy || 0}</p>
        </div>
      </div>

      {/* Device Type Distribution */}
      {summary?.by_device_type && Object.keys(summary.by_device_type).length > 0 && (
        <div className="rounded-xl shadow-sm p-6 sm:p-8" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <h2 className="text-lg sm:text-xl font-semibold mb-6" style={{ color: 'var(--color-pastel-text)' }}>Device Types</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
            {Object.entries(summary.by_device_type).map(([type, count]) => (
              <div key={type} className="text-center p-5 rounded-lg" style={{ backgroundColor: 'var(--color-pastel-primary-lighter)' }}>
                <p className="text-3xl sm:text-4xl font-bold" style={{ color: 'var(--color-pastel-primary)' }}>{count}</p>
                <p className="mt-2 text-sm" style={{ color: 'var(--color-pastel-text-light)' }}>{type}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent Devices */}
      {devices?.devices && devices.devices.length > 0 && (
        <div className="rounded-xl shadow-sm" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
          <div className="px-6 sm:px-8 py-6 flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3" style={{ borderBottom: '1px solid var(--color-pastel-border)' }}>
            <h2 className="text-lg sm:text-xl font-semibold" style={{ color: 'var(--color-pastel-text)' }}>Devices</h2>
            <Link
              to="/devices"
              className="text-sm font-medium hover:opacity-80 transition-opacity"
              style={{ color: 'var(--color-pastel-primary)' }}
            >
              View all →
            </Link>
          </div>
          <ul className="divide-y" style={{ borderColor: 'var(--color-pastel-border)' }}>
            {devices.devices.slice(0, 5).map((device) => (
              <li key={device.device_id} className="px-6 sm:px-8 py-5 sm:py-6 hover:bg-opacity-50 transition-colors" style={{ '--hover-bg': 'var(--color-pastel-primary-lighter)' } as any}>
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div>
                    <Link
                      to={`/devices/${device.device_id}`}
                      className="text-sm font-medium hover:opacity-80 transition-opacity"
                      style={{ color: 'var(--color-pastel-text)' }}
                    >
                      {device.device_id}
                    </Link>
                    <p className="mt-1 text-sm" style={{ color: 'var(--color-pastel-text-muted)' }}>{device.device_type}</p>
                  </div>
                  <StatusBadge status={device.status} />
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
