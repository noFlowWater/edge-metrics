import type { DeviceStatus } from '~/types/device';

interface StatusBadgeProps {
  status: DeviceStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const styles: Record<DeviceStatus, { bg: string; color: string }> = {
    healthy: { bg: 'var(--color-pastel-success-bg)', color: 'var(--color-pastel-success)' },
    unhealthy: { bg: 'var(--color-pastel-danger-bg)', color: 'var(--color-pastel-danger)' },
    unreachable: { bg: 'var(--color-pastel-warning-bg)', color: 'var(--color-pastel-warning)' },
    unknown: { bg: '#f1f5f9', color: 'var(--color-pastel-text-muted)' },
  };

  return (
    <span
      className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
      style={{ backgroundColor: styles[status].bg, color: styles[status].color }}
    >
      {status}
    </span>
  );
}
