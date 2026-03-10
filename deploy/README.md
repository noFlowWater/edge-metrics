# Edge Metrics 배포 가이드

> **DEPRECATED**: Shell 스크립트 기반 배포는 deprecated 되었습니다. Helm chart를 사용하세요.

## Helm 설치 (권장)

```bash
# OCI registry에서 직접 설치
helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics \
  --namespace monitoring --create-namespace

# 개발 환경
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-dev.yaml \
  --namespace monitoring --create-namespace

# 프로덕션 환경
helm install edge-metrics charts/edge-metrics \
  -f charts/edge-metrics/values-prod.yaml \
  --namespace monitoring --create-namespace

# 삭제
helm uninstall edge-metrics -n monitoring
```

### 주요 설정

```bash
# Stub 모드 (하드웨어 없이 테스트)
helm install edge-metrics charts/edge-metrics \
  --set exporter.profiles.jetson.env.STUB_MODE=true \
  --set exporter.profiles.generic.env.STUB_MODE=true

# ServiceMonitor 활성화 (Prometheus Operator 필요)
helm install edge-metrics charts/edge-metrics \
  --set server.serviceMonitor.enabled=true \
  --set exporter.serviceMonitor.enabled=true

# PVC 비활성화 (emptyDir 사용)
helm install edge-metrics charts/edge-metrics \
  --set server.persistence.enabled=false
```

### 포트포워드

```bash
# Backend API
kubectl port-forward -n monitoring svc/edge-metrics-server 8081:8081

# Frontend Dashboard
kubectl port-forward -n monitoring svc/edge-metrics-frontend 3000:3000
```

## Prerequisites

- Kubernetes cluster (K8s v1.31+)
- Helm 3.x
- kubectl configured
- 노드 라벨: `device-family=jetson` (Jetson 노드), `device-family=generic` (Generic 노드)

---

## Legacy: Shell 스크립트 배포 (Deprecated)

> 아래 스크립트들은 2026-06-01 이후 제거됩니다.

### 기본 사용법

```bash
# 전체 스택 배포 (로컬 이미지)
./deploy/deploy-all.sh

# 특정 버전 태그
./deploy/deploy-all.sh v1.0.0
```

### 환경변수 옵션

```bash
# 레지스트리 사용
REGISTRY=daclab ./deploy/deploy-all.sh v1.0.0

# 백엔드 PVC 사용 (데이터 영구 보존)
USE_PVC=true ./deploy/deploy-all.sh v1.0.0

# 백엔드 로컬 config.db 파일 마운트 (개발 전용!)
USE_LOCAL_DB=true ./deploy/deploy-all.sh v1.0.0

# ServiceMonitor 포함 (Prometheus Operator)
DEPLOY_SERVICEMONITOR=true ./deploy/deploy-all.sh v1.0.0

# 전체 옵션
NAMESPACE=monitoring \
REGISTRY=daclab \
USE_PVC=true \
DEPLOY_SERVICEMONITOR=true \
./deploy/deploy-all.sh v1.0.0
```

### 배포 순서

1. **Backend 배포**
   - Docker 이미지 빌드
   - 이미지 푸시 (레지스트리 사용 시)
   - RBAC, PVC, Deployment, Service 생성
   - ServiceMonitor 생성 (선택)

2. **Exporter DaemonSet 배포**
   - Jetson DaemonSet (privileged, tegrastats hostPath)
   - Generic DaemonSet (non-privileged, security hardened)
   - Headless Service + ServiceMonitor (5s interval) 각 프로파일

3. **Frontend 배포**
   - Docker 이미지 빌드
   - 이미지 푸시 (레지스트리 사용 시)
   - Deployment, Service 생성
   - 환경변수로 Backend Service DNS 주입

4. **상태 확인**
   - 전체 Pod 상태
   - 전체 Service 상태
   - 접속 정보 출력

## 통합 삭제 (전체 스택)

### 기본 사용법

```bash
# 전체 스택 삭제 (확인 프롬프트 표시)
./deploy/undeploy-all.sh

# 확인 없이 즉시 삭제
FORCE=true ./deploy/undeploy-all.sh
```

### 환경변수 옵션

```bash
# PVC까지 삭제 (백엔드 데이터 영구 손실!)
DELETE_PVC=true ./deploy/undeploy-all.sh

# Docker 이미지까지 삭제
DELETE_IMAGE=true ./deploy/undeploy-all.sh

# 전체 옵션
NAMESPACE=monitoring \
DELETE_PVC=true \
DELETE_IMAGE=true \
FORCE=true \
./deploy/undeploy-all.sh
```

### 삭제 순서

1. **Frontend 삭제**
   - Service → Deployment → Pod 종료 대기

2. **Exporter DaemonSet 삭제**
   - ServiceMonitor → Service → DaemonSet (Jetson + Generic 모두)

3. **Backend 삭제**
   - ServiceMonitor → Service → Deployment → PVC (선택) → RBAC

4. **상태 확인**
   - 남은 리소스 확인
   - PVC 보존 여부 안내

## 환경변수 레퍼런스

### deploy-all.sh

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | Kubernetes 네임스페이스 |
| `REGISTRY` | (없음) | Docker 레지스트리 주소 |
| `USE_PVC` | false | 백엔드 PVC 사용 여부 |
| `USE_LOCAL_DB` | false | 백엔드 로컬 config.db 파일 마운트 (개발 전용) |
| `DEPLOY_SERVICEMONITOR` | false | ServiceMonitor 배포 여부 |

### undeploy-all.sh

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | monitoring | 삭제할 네임스페이스 |
| `DELETE_PVC` | false | 백엔드 PVC 삭제 여부 (데이터 손실 주의!) |
| `DELETE_IMAGE` | false | Docker 이미지 삭제 여부 |
| `FORCE` | false | 확인 없이 즉시 삭제 |

## 개별 배포 (컴포넌트별)

### Backend만 배포

```bash
cd edge-metrics-server
./scripts/deploy.sh v1.0.0
```

### Exporter만 배포

```bash
cd edge-metrics-exporter

# 전체 (Jetson + Generic)
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --all

# Jetson만
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --jetson

# Generic만
REGISTRY=daclab ./scripts/deploy.sh v1.0.0 --generic
```

### Frontend만 배포

```bash
cd edge-metrics-front
./scripts/deploy.sh v1.0.0
```

각 프로젝트의 README에서 자세한 옵션을 확인하세요:
- [Backend README](../edge-metrics-server/README.md)
- [Exporter README](../edge-metrics-exporter/README.md)
- [Frontend README](../edge-metrics-front/README.md)

## 배포 후 접속

### Frontend (Dashboard)

- **NodePort**: `http://<NodeIP>:30080`
- **Port Forward**:
  ```bash
  kubectl port-forward -n monitoring svc/edge-metrics-front 3000:3000
  # http://localhost:3000
  ```

### Backend (API)

- **NodePort**: `http://<NodeIP>:31716`
- **Port Forward**:
  ```bash
  kubectl port-forward -n monitoring svc/edge-metrics-server 8081:8081
  # http://localhost:8081
  ```

### Exporter (Metrics)

각 엣지 노드 IP에서 직접 접근 (hostNetwork):
```bash
# 메트릭
curl http://<edge-node-ip>:9102/metrics

# Health
curl http://<edge-node-ip>:9101/health
```

## 로그 확인

```bash
# Backend 로그
kubectl logs -n monitoring -l app=edge-metrics-server --tail=50 -f

# Exporter 로그 (Jetson + Generic)
kubectl logs -n monitoring -l 'app in (edge-metrics-exporter-jetson,edge-metrics-exporter-generic)' --tail=50 -f

# Frontend 로그
kubectl logs -n monitoring -l app=edge-metrics-front --tail=50 -f
```

## 트러블슈팅

### Frontend가 Backend에 연결되지 않음

**원인**: API URL 환경변수 미설정 또는 Service DNS 문제

**확인**:
```bash
# Frontend Pod의 환경변수 확인
kubectl get pod -n monitoring -l app=edge-metrics-front -o jsonpath='{.items[0].spec.containers[0].env}'

# Backend Service 존재 확인
kubectl get svc -n monitoring edge-metrics-server
```

### Exporter Pod가 CrashLoopBackOff

**원인**: config 서버 연결 실패 + 로컬 config 없음

**확인**:
```bash
# Pod 로그 확인
kubectl logs -n monitoring -l app=edge-metrics-exporter-jetson --tail=50

# EdgeMesh 라벨 확인
kubectl get svc -n monitoring edge-metrics-server --show-labels
```

**해결**:
1. `edge-metrics-server` 서비스에 EdgeMesh 라벨 추가:
   ```bash
   kubectl label svc -n monitoring edge-metrics-server \
       kubeedge.io/edgemesh-service=true \
       service.edgemesh.kubeedge.io/service-proxy-name=edgemesh
   ```
2. 로컬 config fallback: `/var/lib/edge-metrics/config.yaml` 존재 확인

### Pod가 ImagePullBackOff 상태

**원인**: Docker 이미지를 찾을 수 없음

**해결**:
- 레지스트리 사용 시: 이미지가 푸시되었는지 확인
  ```bash
  docker pull daclab/edge-metrics-exporter:v1.0.0
  ```
- KubeEdge 엣지 노드: TLS 프록시/네트워크 이슈 확인

### PVC가 Pending 상태

**원인**: StorageClass가 없거나 프로비저너 미설정

**확인**:
```bash
kubectl get storageclass
kubectl describe pvc -n monitoring edge-metrics-data
```

**해결**:
1. `manifests/pvc.yaml`에서 `storageClassName` 주석 해제 및 설정
2. 또는 `USE_PVC=false`로 emptyDir 사용

## Architecture

```
Kubernetes Cluster (monitoring namespace)
│
├── Backend (edge-metrics-server)
│   ├── Deployment: edge-metrics-server
│   ├── Service: :8081 (NodePort)
│   ├── PVC: edge-metrics-data (선택)
│   └── RBAC, ServiceMonitor
│
├── Exporter (edge-metrics-exporter)
│   ├── DaemonSet: edge-metrics-exporter-jetson (Jetson 3대, privileged)
│   ├── DaemonSet: edge-metrics-exporter-generic (Generic 3대, non-privileged)
│   ├── Headless Service x2 (ServiceMonitor 연동)
│   └── ServiceMonitor x2 (5s interval)
│
└── Frontend (edge-metrics-front)
    ├── Deployment: edge-metrics-front
    ├── Service: :3000 (NodePort 30080)
    └── API 연동: http://edge-metrics-server:8081 (Service DNS)
```

## License

MIT
