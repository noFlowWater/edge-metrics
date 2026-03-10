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

# Edge Metrics Exporter DaemonSet 제거
# 사용법: ./scripts/undeploy.sh [--jetson|--generic|--all]
# 환경변수: NAMESPACE

set -e

TARGET=${1:---all}
NAMESPACE=${NAMESPACE:-monitoring}

echo "=== Edge Metrics Exporter 제거 ==="
echo "  대상: $TARGET"
echo "  네임스페이스: $NAMESPACE"
echo ""

if [ "$TARGET" = "--jetson" ] || [ "$TARGET" = "--all" ]; then
    echo "[jetson] 제거 중..."
    kubectl delete servicemonitor -n "$NAMESPACE" edge-metrics-exporter-jetson --ignore-not-found
    kubectl delete svc -n "$NAMESPACE" edge-metrics-exporter-jetson --ignore-not-found
    kubectl delete ds -n "$NAMESPACE" edge-metrics-exporter-jetson --ignore-not-found
    echo ""
fi

if [ "$TARGET" = "--generic" ] || [ "$TARGET" = "--all" ]; then
    echo "[generic] 제거 중..."
    kubectl delete servicemonitor -n "$NAMESPACE" edge-metrics-exporter-generic --ignore-not-found
    kubectl delete svc -n "$NAMESPACE" edge-metrics-exporter-generic --ignore-not-found
    kubectl delete ds -n "$NAMESPACE" edge-metrics-exporter-generic --ignore-not-found
    echo ""
fi

echo "제거 완료"
