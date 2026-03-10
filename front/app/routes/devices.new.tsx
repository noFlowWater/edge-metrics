import { useState } from 'react';
import { useNavigate, Link } from 'react-router';
import { api } from '~/lib/api';

const DEFAULT_CONFIG = {
  device_type: 'jetson_orin',
  ip_address: '',
  port: 9100,
  reload_port: 9101,
};

export function meta() {
  return [
    { title: 'Add Device - Edge Metrics' },
    { name: 'description', content: 'Add new device' },
  ];
}

export default function AddDevice() {
  const navigate = useNavigate();
  const [deviceId, setDeviceId] = useState('');
  const [ipAddress, setIpAddress] = useState('');
  const [configJson, setConfigJson] = useState(JSON.stringify(DEFAULT_CONFIG, null, 2));
  const [saving, setSaving] = useState(false);
  const [jsonError, setJsonError] = useState<string | null>(null);

  function handleJsonChange(value: string) {
    setConfigJson(value);
    try {
      JSON.parse(value);
      setJsonError(null);
    } catch {
      setJsonError('Invalid JSON');
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!deviceId.trim()) {
      alert('Device ID is required');
      return;
    }
    if (!ipAddress.trim()) {
      alert('IP Address is required');
      return;
    }
    if (jsonError) return;

    setSaving(true);
    try {
      const config = JSON.parse(configJson);
      await api.createConfig(deviceId, config);
      navigate(`/devices/${deviceId}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create device');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="max-w-4xl space-y-8">
      <div>
        <Link to="/devices" className="text-sm hover:opacity-80 transition-opacity" style={{ color: 'var(--color-pastel-text-muted)' }}>
          ← Back to devices
        </Link>
        <h1 className="text-3xl lg:text-4xl font-bold mt-3" style={{ color: 'var(--color-pastel-text)' }}>Add Device</h1>
      </div>

      <form onSubmit={handleSubmit} className="rounded-xl shadow-sm p-6 sm:p-8 space-y-6" style={{ backgroundColor: 'var(--color-pastel-card)', border: '1px solid var(--color-pastel-border)' }}>
        <div>
          <label className="block text-sm font-medium mb-3" style={{ color: 'var(--color-pastel-text)' }}>Device ID</label>
          <input
            type="text"
            value={deviceId}
            onChange={(e) => setDeviceId(e.target.value)}
            placeholder="e.g., edge-01"
            className="block w-full rounded-lg px-4 sm:px-5 py-3 focus:outline-none"
            style={{
              backgroundColor: 'var(--color-pastel-bg)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text)'
            }}
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-3" style={{ color: 'var(--color-pastel-text)' }}>IP Address</label>
          <input
            type="text"
            value={ipAddress}
            onChange={(e) => {
              setIpAddress(e.target.value);
              try {
                const config = JSON.parse(configJson);
                config.ip_address = e.target.value;
                setConfigJson(JSON.stringify(config, null, 2));
              } catch {}
            }}
            placeholder="e.g., 155.230.34.203"
            className="block w-full rounded-lg px-4 sm:px-5 py-3 focus:outline-none"
            style={{
              backgroundColor: 'var(--color-pastel-bg)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text)'
            }}
            required
          />
        </div>

        <div>
          <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2 mb-3">
            <label className="block text-sm font-medium" style={{ color: 'var(--color-pastel-text)' }}>Configuration (JSON)</label>
            {jsonError && (
              <span className="text-sm font-medium" style={{ color: 'var(--color-pastel-danger)' }}>{jsonError}</span>
            )}
          </div>
          <textarea
            value={configJson}
            onChange={(e) => handleJsonChange(e.target.value)}
            className="block w-full h-80 sm:h-96 rounded-lg px-4 sm:px-5 py-3 sm:py-4 font-mono text-sm focus:outline-none resize-y"
            style={{
              backgroundColor: '#f8f9fc',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text)'
            }}
            spellCheck={false}
          />
        </div>

        <div className="flex flex-col sm:flex-row justify-end gap-3 sm:gap-4 pt-2">
          <Link
            to="/devices"
            className="px-6 py-3 rounded-lg hover:opacity-90 transition-opacity font-medium text-center order-2 sm:order-1"
            style={{
              backgroundColor: 'var(--color-pastel-card)',
              border: '1px solid var(--color-pastel-border)',
              color: 'var(--color-pastel-text-light)'
            }}
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={saving || !!jsonError}
            className="px-6 py-3 rounded-lg hover:opacity-90 disabled:opacity-50 transition-opacity font-medium order-1 sm:order-2"
            style={{
              backgroundColor: 'var(--color-pastel-primary)',
              color: 'white'
            }}
          >
            {saving ? 'Creating...' : 'Create Device'}
          </button>
        </div>
      </form>
    </div>
  );
}
