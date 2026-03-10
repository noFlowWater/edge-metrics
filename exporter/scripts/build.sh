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

# Edge Metrics Exporter 멀티아키텍처 빌드
# 사용법: ./scripts/build.sh [VERSION]
# 환경변수: REGISTRY (Docker Hub username/org, 필수)

set -e

VERSION=${1:-latest}
REGISTRY=${REGISTRY:-}
IMAGE_NAME="edge-metrics-exporter"

if [ -z "$REGISTRY" ]; then
    echo "REGISTRY 환경변수 필요 (Docker Hub username/org)"
    echo "예: REGISTRY=myuser ./scripts/build.sh v1.0.0"
    exit 1
fi

FULL_IMAGE="$REGISTRY/$IMAGE_NAME:$VERSION"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo "=== Edge Metrics Exporter 빌드 ==="
echo "  이미지: $FULL_IMAGE"
echo "  플랫폼: linux/arm64, linux/amd64"
echo ""

# buildx 빌더 생성 (없으면)
if ! docker buildx inspect edge-builder >/dev/null 2>&1; then
    echo "buildx 빌더 생성..."
    docker buildx create --name edge-builder --use
    docker buildx inspect edge-builder --bootstrap
fi

docker buildx use edge-builder

# 멀티아키텍처 빌드 + push
echo "빌드 + push 중..."
docker buildx build \
    --platform linux/arm64,linux/amd64 \
    --tag "$FULL_IMAGE" \
    --push \
    .

echo ""
echo "빌드 완료: $FULL_IMAGE"
