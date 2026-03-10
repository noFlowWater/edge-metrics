#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-edge-metrics-e2e}"
NAMESPACE="${NAMESPACE:-monitoring}"

echo "=== Creating k3d cluster: $CLUSTER_NAME ==="
k3d cluster create "$CLUSTER_NAME" --wait

echo "=== Creating namespace: $NAMESPACE ==="
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

echo "=== Building images ==="
docker build -t edge-metrics-server:e2e ./server
docker build -t edge-metrics-exporter:e2e ./exporter
docker build -t edge-metrics-front:e2e ./front

echo "=== Loading images into k3d ==="
k3d image import edge-metrics-server:e2e edge-metrics-exporter:e2e edge-metrics-front:e2e -c "$CLUSTER_NAME"

echo "=== Setup complete ==="
