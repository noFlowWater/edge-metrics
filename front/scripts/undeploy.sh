#!/bin/bash

# ============================================================
# DEPRECATED: This script is deprecated.
# Use Helm instead: helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics
# Removal date: 2026-06-01
# ============================================================
REMOVAL_DATE="2026-06-01"
echo ""
echo "WARNING: This script is DEPRECATED."
echo "Use Helm instead: helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics"
echo "This script will be removed after $REMOVAL_DATE."
echo ""
if [[ "$(date +%Y-%m-%d)" > "$REMOVAL_DATE" ]]; then
    echo "ERROR: This script has expired. Please use Helm."
    echo "  helm install edge-metrics oci://ghcr.io/noflowwater/charts/edge-metrics"
    exit 1
fi

# Edge Metrics Frontend 자동 삭제 스크립트
# 사용법: ./scripts/undeploy.sh
# 환경변수:
#   NAMESPACE=monitoring (default)
#   DELETE_IMAGE=false (default)
#   FORCE=false (default) - true 시 확인 없이 삭제

set -e  # 에러 발생 시 즉시 종료

# 설정
NAMESPACE=${NAMESPACE:-monitoring}
DELETE_IMAGE=${DELETE_IMAGE:-false}
FORCE=${FORCE:-false}
IMAGE_NAME="edge-metrics-front"

# 색상
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# 배너
echo -e "${RED}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${RED}║  🗑️  Edge Metrics Frontend 자동 삭제        ║${NC}"
echo -e "${RED}╚═══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}설정:${NC}"
echo "  네임스페이스: $NAMESPACE"
echo "  Docker 이미지 삭제: $DELETE_IMAGE"
echo ""

# 확인 메시지
if [ "$FORCE" != "true" ]; then
    echo -e "${YELLOW}⚠️  경고: 모든 Frontend 리소스가 삭제됩니다!${NC}"
    echo ""
    read -p "정말로 삭제하시겠습니까? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "삭제 취소"
        exit 0
    fi
    echo ""
fi

# 1. Service 삭제
echo -e "${YELLOW}[1/3] 🌐 Service 삭제...${NC}"
kubectl delete -f manifests/service.yaml --ignore-not-found=true
echo -e "${RED}✓ Service 삭제됨${NC}"
echo ""

# 2. Deployment 삭제
echo -e "${YELLOW}[2/3] 🚢 Deployment 삭제...${NC}"
kubectl delete -f manifests/deployment.yaml --ignore-not-found=true
echo -e "${RED}✓ Deployment 삭제됨${NC}"
echo ""

# Pod 종료 대기
echo -e "${YELLOW}[3/3] ⏳ Pod 종료 대기...${NC}"
if kubectl wait --for=delete pod -l app=edge-metrics-front -n "$NAMESPACE" --timeout=60s 2>/dev/null; then
    echo -e "${RED}✓ 모든 Pod 종료됨${NC}"
else
    echo "  (타임아웃 또는 Pod 없음)"
fi
echo ""

# Docker 이미지 삭제 (선택)
if [ "$DELETE_IMAGE" = "true" ]; then
    echo -e "${YELLOW}[추가] 🐳 Docker 이미지 삭제...${NC}"
    if docker rmi "$IMAGE_NAME:latest" 2>/dev/null; then
        echo -e "${RED}✓ 이미지 삭제됨${NC}"
    else
        echo "  (이미지 없음 또는 사용 중)"
    fi
    echo ""
fi

# 완료 메시지
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}✅ 삭제 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "남은 리소스 확인:"
echo ""
kubectl get all -n "$NAMESPACE" -l app=edge-metrics-front 2>/dev/null || echo "  (리소스 없음)"
echo ""
