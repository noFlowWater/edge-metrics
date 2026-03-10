import { useState, useEffect } from 'react';
import type { DeviceState } from '~/types/device';

interface EditDeviceModalProps {
  isOpen: boolean;
  onClose: () => void;
  device: DeviceState;
  onSave: (data: {
    device_type: string;
    ip_address: string;
    port: number;
    reload_port: number;
  }) => Promise<void>;
}

export default function EditDeviceModal({
  isOpen,
  onClose,
  device,
  onSave,
}: EditDeviceModalProps) {
  const [deviceType, setDeviceType] = useState(device.device_type);
  const [ipAddress, setIpAddress] = useState(device.ip_address);
  const [port, setPort] = useState(device.port.toString());
  const [reloadPort, setReloadPort] = useState(device.reload_port.toString());
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Reset form when device changes or modal opens
  useEffect(() => {
    if (isOpen) {
      setDeviceType(device.device_type);
      setIpAddress(device.ip_address);
      setPort(device.port.toString());
      setReloadPort(device.reload_port.toString());
      setError(null);
    }
  }, [isOpen, device]);

  if (!isOpen) return null;

  const validateIP = (ip: string): boolean => {
    const ipPattern = /^(\d{1,3}\.){3}\d{1,3}$/;
    if (!ipPattern.test(ip)) return false;

    const parts = ip.split('.');
    return parts.every(part => {
      const num = parseInt(part, 10);
      return num >= 0 && num <= 255;
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (!deviceType.trim()) {
      setError('Device type is required');
      return;
    }

    if (!ipAddress.trim()) {
      setError('IP address is required');
      return;
    }

    if (!validateIP(ipAddress)) {
      setError('Invalid IP address format');
      return;
    }

    const portNum = parseInt(port, 10);
    const reloadPortNum = parseInt(reloadPort, 10);

    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      setError('Port must be between 1 and 65535');
      return;
    }

    if (isNaN(reloadPortNum) || reloadPortNum < 1 || reloadPortNum > 65535) {
      setError('Reload port must be between 1 and 65535');
      return;
    }

    setIsSaving(true);
    try {
      await onSave({
        device_type: deviceType,
        ip_address: ipAddress,
        port: portNum,
        reload_port: reloadPortNum,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update device');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    setError(null);
    onClose();
  };

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
      }}
      onClick={handleCancel}
    >
      <div
        style={{
          backgroundColor: 'var(--color-pastel-card)',
          borderRadius: '12px',
          padding: '24px',
          width: '90%',
          maxWidth: '500px',
          boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <h2
          style={{
            fontSize: '20px',
            fontWeight: '600',
            marginBottom: '20px',
            color: 'var(--color-pastel-text)',
          }}
        >
          Edit Device: {device.device_id}
        </h2>

        {/* Form */}
        <form onSubmit={handleSubmit}>
          {/* Device Type */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="device_type"
              style={{
                display: 'block',
                fontSize: '14px',
                fontWeight: '500',
                marginBottom: '6px',
                color: 'var(--color-pastel-text)',
              }}
            >
              Device Type
            </label>
            <input
              id="device_type"
              type="text"
              value={deviceType}
              onChange={(e) => setDeviceType(e.target.value)}
              required
              style={{
                width: '100%',
                padding: '8px 12px',
                border: '1px solid var(--color-pastel-border)',
                borderRadius: '6px',
                fontSize: '14px',
                backgroundColor: 'var(--color-pastel-bg)',
              }}
              placeholder="e.g., jetson_orin, raspberry_pi"
            />
          </div>

          {/* IP Address */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="ip_address"
              style={{
                display: 'block',
                fontSize: '14px',
                fontWeight: '500',
                marginBottom: '6px',
                color: 'var(--color-pastel-text)',
              }}
            >
              IP Address
            </label>
            <input
              id="ip_address"
              type="text"
              value={ipAddress}
              onChange={(e) => setIpAddress(e.target.value)}
              required
              style={{
                width: '100%',
                padding: '8px 12px',
                border: '1px solid var(--color-pastel-border)',
                borderRadius: '6px',
                fontSize: '14px',
                backgroundColor: 'var(--color-pastel-bg)',
              }}
              placeholder="e.g., 192.168.1.10"
            />
          </div>

          {/* Port */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="port"
              style={{
                display: 'block',
                fontSize: '14px',
                fontWeight: '500',
                marginBottom: '6px',
                color: 'var(--color-pastel-text)',
              }}
            >
              Metrics Port
            </label>
            <input
              id="port"
              type="number"
              value={port}
              onChange={(e) => setPort(e.target.value)}
              required
              min="1"
              max="65535"
              style={{
                width: '100%',
                padding: '8px 12px',
                border: '1px solid var(--color-pastel-border)',
                borderRadius: '6px',
                fontSize: '14px',
                backgroundColor: 'var(--color-pastel-bg)',
              }}
              placeholder="9100"
            />
          </div>

          {/* Reload Port */}
          <div style={{ marginBottom: '16px' }}>
            <label
              htmlFor="reload_port"
              style={{
                display: 'block',
                fontSize: '14px',
                fontWeight: '500',
                marginBottom: '6px',
                color: 'var(--color-pastel-text)',
              }}
            >
              Reload Port
            </label>
            <input
              id="reload_port"
              type="number"
              value={reloadPort}
              onChange={(e) => setReloadPort(e.target.value)}
              required
              min="1"
              max="65535"
              style={{
                width: '100%',
                padding: '8px 12px',
                border: '1px solid var(--color-pastel-border)',
                borderRadius: '6px',
                fontSize: '14px',
                backgroundColor: 'var(--color-pastel-bg)',
              }}
              placeholder="9101"
            />
          </div>

          {/* Error Message */}
          {error && (
            <div
              style={{
                padding: '10px',
                marginBottom: '16px',
                backgroundColor: '#fee',
                border: '1px solid var(--color-pastel-danger)',
                borderRadius: '6px',
                color: 'var(--color-pastel-danger)',
                fontSize: '14px',
              }}
            >
              {error}
            </div>
          )}

          {/* Info Message */}
          <div
            style={{
              padding: '10px',
              marginBottom: '16px',
              backgroundColor: '#f0f9ff',
              border: '1px solid #bfdbfe',
              borderRadius: '6px',
              color: '#1e40af',
              fontSize: '13px',
            }}
          >
            Note: This will only update the database. Device reload is not triggered.
          </div>

          {/* Buttons */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              gap: '12px',
            }}
          >
            <button
              type="button"
              onClick={handleCancel}
              disabled={isSaving}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                fontSize: '14px',
                fontWeight: '500',
                cursor: isSaving ? 'not-allowed' : 'pointer',
                backgroundColor: 'var(--color-pastel-card)',
                border: '1px solid var(--color-pastel-border)',
                color: 'var(--color-pastel-text-light)',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              style={{
                padding: '8px 16px',
                borderRadius: '6px',
                fontSize: '14px',
                fontWeight: '500',
                cursor: isSaving ? 'not-allowed' : 'pointer',
                backgroundColor: isSaving
                  ? '#9ca3af'
                  : 'var(--color-pastel-primary)',
                color: 'white',
                border: 'none',
              }}
            >
              {isSaving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
