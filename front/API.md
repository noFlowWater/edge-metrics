# Edge Metrics Frontend

React-based management UI for Edge Metrics Server.

## Tech Stack

- **Framework**: React Router v7
- **Styling**: Tailwind CSS
- **Language**: TypeScript

## Pages

### Dashboard (`/`)
- Total device count, Healthy/Unhealthy summary
- Device type distribution
- Recent device list

### Devices (`/devices`)
- Device list table
- Filter by status/type
- Individual/batch reload

### Device Detail (`/devices/:id`)
- Device status display
- Config view/edit form
- Reload/Delete actions

### Add Device (`/devices/new`)
- New device registration form
- Device-type-specific extra config (Jetson, Shelly)

## API Integration

API calls are made via `app/lib/api.ts`:

```typescript
import { api } from '~/lib/api';

// Usage
const devices = await api.getDevices();
const config = await api.getConfig('edge-01');
await api.updateConfig('edge-01', configData);
await api.reloadDevice('edge-01');
```

### API Functions

| Function | Method | Endpoint | Description |
|----------|--------|----------|-------------|
| `health()` | GET | /health | Server health check |
| `getConfigs()` | GET | /config | List all configs |
| `getConfig(id)` | GET | /config/:id | Get device config |
| `createConfig(id, config)` | POST | /config/:id | Register new device |
| `updateConfig(id, config)` | PUT | /config/:id | Update config |
| `patchConfig(id, config)` | PATCH | /config/:id | Partial config update |
| `deleteConfig(id)` | DELETE | /config/:id | Delete device |
| `getDevices()` | GET | /devices | List device statuses |
| `getDeviceStatus(id)` | GET | /devices/:id/status | Get device status |
| `reloadDevice(id)` | POST | /devices/:id/reload | Reload device |
| `reloadAllDevices()` | POST | /devices/reload | Reload all devices |
| `getMetricsSummary()` | GET | /metrics/summary | Summary statistics |

## Configuration

API Base URL is configured in `app/lib/api.ts`:

```typescript
const API_BASE = 'http://localhost:8081';
```

## Running

```bash
# Development server
npm run dev

# Build
npm run build

# Production
npm start
```
