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

# Edge Metrics Server Docker 이미지 빌드 스크립트
# 사용법: ./scripts/build.sh [VERSION]
# 예시: ./scripts/build.sh v1.0.0
# 환경변수:
#   REGISTRY= (예: myregistry.com)
#   PUSH=false (default)
#   PLATFORM= (예: linux/amd64,linux/arm64)

set -e  # 에러 발생 시 즉시 종료

# 설정
VERSION=${1:-latest}
REGISTRY=${REGISTRY:-}
IMAGE_NAME="edge-metrics-server"
PUSH=${PUSH:-false}
PLATFORM=${PLATFORM:-}

# 색상
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# 배너
echo -e "${GREEN}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  🔨 Edge Metrics Server Docker 빌드         ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}설정:${NC}"
echo "  버전: $VERSION"

# 이미지 이름 결정
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE="$REGISTRY/$IMAGE_NAME:$VERSION"
    echo "  레지스트리: $REGISTRY"
else
    FULL_IMAGE="$IMAGE_NAME:$VERSION"
    echo "  레지스트리: (로컬)"
fi

if [ -n "$PLATFORM" ]; then
    echo "  플랫폼: $PLATFORM"
else
    echo "  플랫폼: (기본)"
fi

echo "  푸시: $PUSH"
echo ""

# Docker 빌드
echo -e "${YELLOW}🔨 Docker 이미지 빌드 중...${NC}"
echo "  이미지: $FULL_IMAGE"
echo ""

if [ -n "$PLATFORM" ]; then
    # Multi-platform build (buildx 필요)
    if [ "$PUSH" = "true" ]; then
        docker buildx build \
            --platform "$PLATFORM" \
            --push \
            -t "$FULL_IMAGE" \
            .
        echo ""
        echo -e "${GREEN}✓ 멀티 플랫폼 빌드 및 푸시 완료${NC}"
    else
        docker buildx build \
            --platform "$PLATFORM" \
            --load \
            -t "$FULL_IMAGE" \
            .
        echo ""
        echo -e "${GREEN}✓ 멀티 플랫폼 빌드 완료 (로컬)${NC}"
    fi
else
    # 단일 플랫폼 빌드
    docker build -t "$FULL_IMAGE" .
    echo ""
    echo -e "${GREEN}✓ 빌드 완료${NC}"

    # 푸시 (레지스트리 사용 시)
    if [ "$PUSH" = "true" ] && [ -n "$REGISTRY" ]; then
        echo ""
        echo -e "${YELLOW}📤 이미지 푸시: $FULL_IMAGE${NC}"
        docker push "$FULL_IMAGE"
        echo -e "${GREEN}✓ 푸시 완료${NC}"
    fi
fi

# 완료 메시지
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ 빌드 완료!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "이미지 정보:"
docker images "$IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
echo ""

if [ "$PUSH" != "true" ] && [ -n "$REGISTRY" ]; then
    echo -e "${BLUE}💡 푸시하려면:${NC}"
    echo "  PUSH=true ./scripts/build.sh $VERSION"
    echo ""
fi

if [ -z "$PLATFORM" ]; then
    echo -e "${BLUE}💡 멀티 플랫폼 빌드:${NC}"
    echo "  PLATFORM=linux/amd64,linux/arm64 PUSH=true ./scripts/build.sh $VERSION"
    echo ""
fi
