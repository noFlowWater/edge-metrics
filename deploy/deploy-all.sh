#!/bin/bash

# ============================================================
# DEPRECATED: This script is deprecated.
# Use Helm instead: helm uninstall edge-metrics -n monitoring
# Removal date: 2026-06-01
# ============================================================
REMOVAL_DATE="2026-06-01"
echo ""
echo "WARNING: This script is DEPRECATED."
echo "Use Helm instead: helm install/uninstall edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics"
echo "This script will be removed after $REMOVAL_DATE."
echo ""
if [[ "$(date +%Y-%m-%d)" > "$REMOVAL_DATE" ]]; then
    echo "ERROR: This script has expired. Please use Helm."
    exit 1
fi

# Edge Metrics 전체 스택 배포 스크립트 (Backend + Frontend)
# 사용법: ./deploy/deploy-all.sh [VERSION]
# 예시: ./deploy/deploy-all.sh v1.0.0
# 환경변수:
#   NAMESPACE=monitoring (default)
#   REGISTRY= (예: myregistry.com)
#   USE_PVC=false (백엔드 PVC 사용 여부)
#   USE_LOCAL_DB=false (백엔드 로컬 config.db 파일 마운트, 개발 전용)
#   DEPLOY_SERVICEMONITOR=false (ServiceMonitor 배포 여부)

set -e  # 에러 발생 시 즉시 종료

# 설정
VERSION=${1:-latest}
NAMESPACE=${NAMESPACE:-monitoring}
REGISTRY=${REGISTRY:-}
USE_PVC=${USE_PVC:-false}
USE_LOCAL_DB=${USE_LOCAL_DB:-false}
DEPLOY_SERVICEMONITOR=${DEPLOY_SERVICEMONITOR:-false}

# 색상
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'  # No Color

# 스크립트 실행 위치 확인
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# 배너
echo -e "${CYAN}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  🚀 Edge Metrics 전체 스택 배포             ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}설정:${NC}"
echo "  버전: $VERSION"
echo "  네임스페이스: $NAMESPACE"
echo "  레지스트리: ${REGISTRY:-(로컬)}"
echo "  백엔드 PVC: $USE_PVC"
echo "  로컬 DB 파일: $USE_LOCAL_DB"
echo "  ServiceMonitor: $DEPLOY_SERVICEMONITOR"
echo ""

# 환경변수 export (하위 스크립트에서 사용)
export NAMESPACE
export REGISTRY
export USE_PVC
export USE_LOCAL_DB
export DEPLOY_SERVICEMONITOR

# 1. Backend 배포
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[1/3] 🔧 Backend 배포 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-server"
./scripts/deploy.sh "$VERSION"

echo ""
echo -e "${GREEN}✓ Backend 배포 완료${NC}"
echo ""

# 2. Exporter DaemonSet 배포
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[2/3] 📊 Exporter DaemonSet 배포 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-exporter"
./scripts/deploy.sh "$VERSION"

echo ""
echo -e "${GREEN}✓ Exporter 배포 완료${NC}"
echo ""

# 3. Frontend 배포
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}[3/3] 🎨 Frontend 배포 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-front"
./scripts/deploy.sh "$VERSION"

echo ""
echo -e "${GREEN}✓ Frontend 배포 완료${NC}"
echo ""

# 3. 전체 상태 확인
cd "$ROOT_DIR"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}✅ 전체 스택 배포 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "전체 리소스 상태:"
echo ""

# Pods
echo -e "${YELLOW}📦 Pods:${NC}"
kubectl get pods -n "$NAMESPACE" -l 'app in (edge-metrics-server,edge-metrics-exporter-jetson,edge-metrics-exporter-generic,edge-metrics-front)'
echo ""

# Services
echo -e "${YELLOW}🌐 Services:${NC}"
kubectl get svc -n "$NAMESPACE" -l 'app in (edge-metrics-server,edge-metrics-exporter-jetson,edge-metrics-exporter-generic,edge-metrics-front)'
echo ""

# 접속 정보
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}접속 정보:${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# NodePort 정보 가져오기
BACKEND_NODEPORT=$(kubectl get svc -n "$NAMESPACE" edge-metrics-server -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")
FRONTEND_NODEPORT=$(kubectl get svc -n "$NAMESPACE" edge-metrics-front -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")

echo -e "${GREEN}Frontend (Dashboard):${NC}"
echo "  NodePort: http://<NodeIP>:$FRONTEND_NODEPORT"
echo "  Port Forward: kubectl port-forward -n $NAMESPACE svc/edge-metrics-front 3000:3000"
echo ""
echo -e "${GREEN}Backend (API):${NC}"
echo "  NodePort: http://<NodeIP>:$BACKEND_NODEPORT"
echo "  Port Forward: kubectl port-forward -n $NAMESPACE svc/edge-metrics-server 8081:8081"
echo ""

echo -e "${BLUE}로그 확인:${NC}"
echo "  Backend:  kubectl logs -n $NAMESPACE -l app=edge-metrics-server --tail=50 -f"
echo "  Frontend: kubectl logs -n $NAMESPACE -l app=edge-metrics-front --tail=50 -f"
echo ""

echo -e "${BLUE}전체 삭제:${NC}"
echo "  ./deploy/undeploy-all.sh"
echo ""
