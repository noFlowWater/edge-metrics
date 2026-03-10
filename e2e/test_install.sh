#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-monitoring}"
RELEASE="${RELEASE:-edge-metrics}"

echo "=== Installing Helm chart ==="
helm install "$RELEASE" charts/edge-metrics -n "$NAMESPACE" \
  --set server.image.repository=edge-metrics-server \
  --set server.image.tag=e2e \
  --set server.image.pullPolicy=Never \
  --set server.persistence.enabled=false \
  --set exporter.image.repository=edge-metrics-exporter \
  --set exporter.image.tag=e2e \
  --set exporter.image.pullPolicy=Never \
  --set exporter.profiles.jetson.enabled=false \
  --set exporter.profiles.generic.env.STUB_MODE=true \
  --set frontend.enabled=false

echo "=== Waiting for server pod ==="
kubectl rollout status deployment/"$RELEASE"-server -n "$NAMESPACE" --timeout=120s

echo "=== Testing /health ==="
kubectl exec -n "$NAMESPACE" deploy/"$RELEASE"-server -- wget -qO- http://localhost:8081/health

echo "=== Testing /devices ==="
kubectl exec -n "$NAMESPACE" deploy/"$RELEASE"-server -- wget -qO- http://localhost:8081/devices

echo "=== Install test PASSED ==="
