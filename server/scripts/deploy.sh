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

# Edge Metrics Server 자동 배포 스크립트
# 사용법: ./scripts/deploy.sh [VERSION]
# 예시: ./scripts/deploy.sh v1.0.0
# 환경변수:
#   NAMESPACE=monitoring (default)
#   REGISTRY= (예: myregistry.com)
#   USE_PVC=false (default)
#   USE_LOCAL_DB=false (default, 로컬 config.db 파일 마운트)
#   DEPLOY_SERVICEMONITOR=false (default)

set -e  # 에러 발생 시 즉시 종료

# 설정
VERSION=${1:-latest}
NAMESPACE=${NAMESPACE:-monitoring}
REGISTRY=${REGISTRY:-}
IMAGE_NAME="edge-metrics-server"
USE_PVC=${USE_PVC:-false}
USE_LOCAL_DB=${USE_LOCAL_DB:-false}
DEPLOY_SERVICEMONITOR=${DEPLOY_SERVICEMONITOR:-false}

# 색상
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# 배너
echo -e "${GREEN}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  🚀 Edge Metrics Server 자동 배포           ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}설정:${NC}"
echo "  버전: $VERSION"
echo "  네임스페이스: $NAMESPACE"
echo "  PVC 사용: $USE_PVC"
echo "  로컬 DB 파일: $USE_LOCAL_DB"
echo "  ServiceMonitor: $DEPLOY_SERVICEMONITOR"

# 이미지 이름 결정
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE="$REGISTRY/$IMAGE_NAME:$VERSION"
    echo "  레지스트리: $REGISTRY"
else
    FULL_IMAGE="$IMAGE_NAME:$VERSION"
    echo "  레지스트리: (로컬)"
fi
echo ""

# USE_LOCAL_DB 경고
if [ "$USE_LOCAL_DB" = "true" ]; then
    echo -e "${YELLOW}⚠️  USE_LOCAL_DB=true: 로컬 config.db 파일 직접 마운트 (개발 전용!)${NC}"
    echo -e "${YELLOW}⚠️  보안 위험: 프로덕션 환경에서 절대 사용 금지${NC}"
    echo -e "${YELLOW}⚠️  단일 노드 클러스터만 지원 (minikube, Docker Desktop 등)${NC}"
    echo -e "${YELLOW}⚠️  Pod가 현재 노드($(hostname))에 고정됩니다${NC}"
    echo ""
fi

# 1. Docker 이미지 빌드
echo -e "${YELLOW}[1/7] 🔨 Docker 이미지 빌드...${NC}"
docker build -t "$FULL_IMAGE" .
echo -e "${GREEN}✓ 빌드 완료${NC}"
echo ""

# 2. 이미지 푸시 (레지스트리 사용 시)
if [ -n "$REGISTRY" ]; then
    echo -e "${YELLOW}[2/7] 📤 이미지 푸시: $FULL_IMAGE${NC}"
    docker push "$FULL_IMAGE"
    echo -e "${GREEN}✓ 푸시 완료${NC}"
else
    echo -e "${YELLOW}[2/7] 📤 이미지 푸시 생략 (로컬 이미지)${NC}"
fi
echo ""

# 3. 네임스페이스 생성 (없으면)
echo -e "${YELLOW}[3/7] 🏗️  네임스페이스 확인...${NC}"
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "  네임스페이스 '$NAMESPACE' 이미 존재"
else
    kubectl create namespace "$NAMESPACE"
    echo "  네임스페이스 '$NAMESPACE' 생성됨"
fi
echo -e "${GREEN}✓ 네임스페이스 준비됨${NC}"
echo ""

# 4. RBAC 배포
echo -e "${YELLOW}[4/7] 🔐 RBAC 배포...${NC}"
kubectl apply -f manifests/rbac.yaml
echo -e "${GREEN}✓ RBAC 배포 완료${NC}"
echo ""

# 5. PVC 배포 (선택)
if [ "$USE_PVC" = "true" ]; then
    echo -e "${YELLOW}[5/7] 💾 PVC 배포...${NC}"
    kubectl apply -f manifests/pvc.yaml
    echo -e "${GREEN}✓ PVC 배포 완료${NC}"
else
    echo -e "${YELLOW}[5/7] 💾 PVC 배포 생략 (emptyDir 사용)${NC}"
fi
echo ""

# 6. Deployment & Service 배포
echo -e "${YELLOW}[6/7] 🚢 Deployment & Service 배포...${NC}"

# deployment.yaml 동적 수정
TEMP_DEPLOYMENT=$(mktemp)

# 이미지 치환
if [ -n "$REGISTRY" ]; then
    cat manifests/deployment.yaml | sed "s|image: edge-metrics-server:latest|image: $FULL_IMAGE|g" > "$TEMP_DEPLOYMENT"
else
    cp manifests/deployment.yaml "$TEMP_DEPLOYMENT"
fi

# PVC 또는 HostPath 설정
if [ "$USE_PVC" = "true" ]; then
    # PVC 사용: emptyDir 주석 처리, PVC 활성화
    sed -i 's|emptyDir: {}|# emptyDir: {}|g' "$TEMP_DEPLOYMENT"
    sed -i 's|# persistentVolumeClaim:|persistentVolumeClaim:|g' "$TEMP_DEPLOYMENT"
    sed -i 's|#   claimName: edge-metrics-data|  claimName: edge-metrics-data|g' "$TEMP_DEPLOYMENT"
elif [ "$USE_LOCAL_DB" = "true" ]; then
    # HostPath 사용: emptyDir 주석 처리, hostPath 활성화, nodeSelector 활성화
    PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
    HOSTNAME=$(hostname)

    # Volume 설정
    sed -i 's|emptyDir: {}|# emptyDir: {}|g' "$TEMP_DEPLOYMENT"
    sed -i 's|# hostPath:|hostPath:|g' "$TEMP_DEPLOYMENT"
    sed -i "s|#   path: /home/genesis1/ETRI/research/edge-metrics/edge-metrics-server|  path: $PROJECT_DIR|g" "$TEMP_DEPLOYMENT"
    sed -i 's|#   type: Directory|  type: Directory|g' "$TEMP_DEPLOYMENT"

    # nodeSelector 활성화 (현재 호스트로 고정)
    sed -i 's|# nodeSelector:|nodeSelector:|g' "$TEMP_DEPLOYMENT"
    sed -i "s|#   kubernetes.io/hostname: genesis1|  kubernetes.io/hostname: $HOSTNAME|g" "$TEMP_DEPLOYMENT"
fi

kubectl apply -f "$TEMP_DEPLOYMENT"
rm "$TEMP_DEPLOYMENT"

kubectl apply -f manifests/service.yaml
echo -e "${GREEN}✓ Deployment & Service 배포 완료${NC}"
echo ""

# 7. ServiceMonitor 배포 (선택)
if [ "$DEPLOY_SERVICEMONITOR" = "true" ]; then
    echo -e "${YELLOW}[7/7] 📊 ServiceMonitor 배포...${NC}"
    kubectl apply -f manifests/servicemonitor.yaml
    echo -e "${GREEN}✓ ServiceMonitor 배포 완료${NC}"
else
    echo -e "${YELLOW}[7/7] 📊 ServiceMonitor 배포 생략${NC}"
fi
echo ""

# 8. 배포 상태 확인
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ 배포 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "배포 상태 확인:"
echo ""
kubectl get pods -n "$NAMESPACE" -l app=edge-metrics-server
echo ""
kubectl get svc -n "$NAMESPACE" edge-metrics-server
echo ""
echo -e "${BLUE}로그 확인:${NC}"
echo "  kubectl logs -n $NAMESPACE -l app=edge-metrics-server --tail=50 -f"
echo ""
echo -e "${BLUE}포트포워드:${NC}"
echo "  kubectl port-forward -n $NAMESPACE svc/edge-metrics-server 8081:8081"
