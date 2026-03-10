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

# Edge Metrics 전체 스택 삭제 스크립트 (Frontend + Backend)
# 사용법: ./deploy/undeploy-all.sh
# 환경변수:
#   NAMESPACE=monitoring (default)
#   DELETE_PVC=false (백엔드 PVC 삭제 여부)
#   DELETE_IMAGE=false (Docker 이미지 삭제 여부)
#   FORCE=false (확인 없이 삭제)

set -e  # 에러 발생 시 즉시 종료

# 설정
NAMESPACE=${NAMESPACE:-monitoring}
DELETE_PVC=${DELETE_PVC:-false}
DELETE_IMAGE=${DELETE_IMAGE:-false}
FORCE=${FORCE:-false}

# 색상
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'  # No Color

# 스크립트 실행 위치 확인
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# 배너
echo -e "${RED}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${RED}║  🗑️  Edge Metrics 전체 스택 삭제            ║${NC}"
echo -e "${RED}╚═══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}설정:${NC}"
echo "  네임스페이스: $NAMESPACE"
echo "  PVC 삭제: $DELETE_PVC"
echo "  Docker 이미지 삭제: $DELETE_IMAGE"
echo ""

# 확인 메시지
if [ "$FORCE" != "true" ]; then
    echo -e "${YELLOW}⚠️  경고: 전체 스택의 모든 리소스가 삭제됩니다!${NC}"
    if [ "$DELETE_PVC" = "true" ]; then
        echo -e "${RED}⚠️  PVC도 삭제됩니다! (백엔드 데이터 영구 손실)${NC}"
    fi
    echo ""
    read -p "정말로 삭제하시겠습니까? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "삭제 취소"
        exit 0
    fi
    echo ""
fi

# 환경변수 export (하위 스크립트에서 사용)
export NAMESPACE
export DELETE_PVC
export DELETE_IMAGE
export FORCE=true  # 하위 스크립트는 확인 없이 실행

# 1. Frontend 삭제
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}[1/3] 🎨 Frontend 삭제 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-front"
./scripts/undeploy.sh

echo ""
echo -e "${RED}✓ Frontend 삭제 완료${NC}"
echo ""

# 2. Exporter DaemonSet 삭제
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}[2/3] 📊 Exporter DaemonSet 삭제 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-exporter"
./scripts/undeploy.sh

echo ""
echo -e "${RED}✓ Exporter 삭제 완료${NC}"
echo ""

# 3. Backend 삭제
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}[3/3] 🔧 Backend 삭제 중...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$ROOT_DIR/edge-metrics-server"
./scripts/undeploy.sh

echo ""
echo -e "${RED}✓ Backend 삭제 완료${NC}"
echo ""

# 3. 전체 상태 확인
cd "$ROOT_DIR"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}✅ 전체 스택 삭제 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "남은 리소스 확인:"
echo ""
kubectl get all -n "$NAMESPACE" -l 'app in (edge-metrics-server,edge-metrics-exporter-jetson,edge-metrics-exporter-generic,edge-metrics-front)' 2>/dev/null || echo "  (리소스 없음)"
echo ""

if [ "$DELETE_PVC" != "true" ]; then
    echo -e "${BLUE}💡 PVC가 보존되었습니다:${NC}"
    kubectl get pvc -n "$NAMESPACE" 2>/dev/null || echo "  (PVC 없음)"
    echo ""
    echo "  PVC를 삭제하려면:"
    echo "  DELETE_PVC=true ./deploy/undeploy-all.sh"
    echo ""
fi
