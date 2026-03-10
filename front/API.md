# Edge Metrics Frontend

Edge Metrics Server를 위한 React 기반 관리 UI입니다.

## Tech Stack

- **Framework**: React Router v7
- **Styling**: Tailwind CSS
- **Language**: TypeScript

## 페이지 구조

### Dashboard (`/`)
- 전체 디바이스 수, Healthy/Unhealthy 요약
- 디바이스 타입별 분포
- 최근 디바이스 목록

### Devices (`/devices`)
- 디바이스 목록 테이블
- 상태/타입별 필터링
- 개별/전체 Reload 기능

### Device Detail (`/devices/:id`)
- 디바이스 상태 표시
- 설정 조회/편집 폼
- Reload/Delete 기능

### Add Device (`/devices/new`)
- 새 디바이스 등록 폼
- device_type에 따른 추가 설정 (Jetson, Shelly)

## API 연동

`app/lib/api.ts`에서 edge-metrics-server API 호출:

```typescript
import { api } from '~/lib/api';

// 사용 예시
const devices = await api.getDevices();
const config = await api.getConfig('edge-01');
await api.updateConfig('edge-01', configData);
await api.reloadDevice('edge-01');
```

### API 함수 목록

| Function | Method | Endpoint | Description |
|----------|--------|----------|-------------|
| `health()` | GET | /health | 서버 상태 확인 |
| `getConfigs()` | GET | /config | 전체 설정 조회 |
| `getConfig(id)` | GET | /config/:id | 디바이스 설정 조회 |
| `createConfig(id, config)` | POST | /config/:id | 새 디바이스 등록 |
| `updateConfig(id, config)` | PUT | /config/:id | 설정 업데이트 |
| `patchConfig(id, config)` | PATCH | /config/:id | 설정 부분 업데이트 |
| `deleteConfig(id)` | DELETE | /config/:id | 디바이스 삭제 |
| `getDevices()` | GET | /devices | 디바이스 상태 목록 |
| `getDeviceStatus(id)` | GET | /devices/:id/status | 디바이스 상태 조회 |
| `reloadDevice(id)` | POST | /devices/:id/reload | 디바이스 리로드 |
| `reloadAllDevices()` | POST | /devices/reload | 전체 리로드 |
| `getMetricsSummary()` | GET | /metrics/summary | 요약 통계 |

## 설정

API Base URL은 `app/lib/api.ts`에서 설정:

```typescript
const API_BASE = 'http://localhost:8081';
```

## 실행

```bash
# 개발 서버
npm run dev

# 빌드
npm run build

# 프로덕션 실행
npm start
```
