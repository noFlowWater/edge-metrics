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

# Edge Metrics Exporter DaemonSet 배포
# 사용법: ./scripts/deploy.sh [VERSION] [--jetson|--generic|--all]
# 환경변수: NAMESPACE, REGISTRY

set -e

VERSION=${1:-latest}
TARGET=${2:---all}
NAMESPACE=${NAMESPACE:-monitoring}
REGISTRY=${REGISTRY:-}
IMAGE_NAME="edge-metrics-exporter"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

if [ -n "$REGISTRY" ]; then
    FULL_IMAGE="$REGISTRY/$IMAGE_NAME:$VERSION"
else
    FULL_IMAGE="$IMAGE_NAME:$VERSION"
fi

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

deploy_manifests() {
    local PROFILE=$1  # jetson or generic
    local DS_FILE="$PROJECT_DIR/manifests/daemonset${PROFILE:+-$PROFILE}.yaml"
    local SVC_FILE="$PROJECT_DIR/manifests/service${PROFILE:+-$PROFILE}.yaml"
    local SM_FILE="$PROJECT_DIR/manifests/servicemonitor${PROFILE:+-$PROFILE}.yaml"

    # jetson uses daemonset.yaml (no suffix), generic uses daemonset-generic.yaml
    if [ "$PROFILE" = "jetson" ]; then
        DS_FILE="$PROJECT_DIR/manifests/daemonset.yaml"
        SVC_FILE="$PROJECT_DIR/manifests/service.yaml"
        SM_FILE="$PROJECT_DIR/manifests/servicemonitor.yaml"
    fi

    echo -e "${YELLOW}  [$PROFILE] DaemonSet + Service + ServiceMonitor 배포...${NC}"

    TEMP_DS=$(mktemp)
    sed "s|image: edge-metrics-exporter:latest|image: $FULL_IMAGE|g" \
        "$DS_FILE" > "$TEMP_DS"
    kubectl apply -f "$TEMP_DS"
    rm "$TEMP_DS"

    kubectl apply -f "$SVC_FILE"
    kubectl apply -f "$SM_FILE"
    echo -e "${GREEN}  [$PROFILE] 배포 완료${NC}"
}

echo -e "${GREEN}=== Edge Metrics Exporter DaemonSet 배포 ===${NC}"
echo "  버전: $VERSION"
echo "  대상: $TARGET"
echo "  네임스페이스: $NAMESPACE"
echo "  이미지: $FULL_IMAGE"
echo ""

# 네임스페이스 확인
echo -e "${YELLOW}[1/2] 네임스페이스 확인...${NC}"
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "  '$NAMESPACE' 이미 존재"
else
    kubectl create namespace "$NAMESPACE"
    echo "  '$NAMESPACE' 생성됨"
fi
echo ""

# 배포
echo -e "${YELLOW}[2/2] 매니페스트 배포...${NC}"
if [ "$TARGET" = "--jetson" ] || [ "$TARGET" = "--all" ]; then
    deploy_manifests "jetson"
fi
if [ "$TARGET" = "--generic" ] || [ "$TARGET" = "--all" ]; then
    deploy_manifests "generic"
fi
echo ""

# 상태 확인
echo -e "${BLUE}=== 배포 상태 ===${NC}"
echo ""
if [ "$TARGET" = "--jetson" ] || [ "$TARGET" = "--all" ]; then
    kubectl get ds -n "$NAMESPACE" edge-metrics-exporter-jetson 2>/dev/null || true
    echo ""
fi
if [ "$TARGET" = "--generic" ] || [ "$TARGET" = "--all" ]; then
    kubectl get ds -n "$NAMESPACE" edge-metrics-exporter-generic 2>/dev/null || true
    echo ""
fi
kubectl get pods -n "$NAMESPACE" -l 'app in (edge-metrics-exporter-jetson,edge-metrics-exporter-generic)' -o wide
echo ""
echo -e "${BLUE}로그 확인:${NC}"
echo "  kubectl logs -n $NAMESPACE -l 'app in (edge-metrics-exporter-jetson,edge-metrics-exporter-generic)' --tail=50 -f"
