#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-monitoring}"
RELEASE="${RELEASE:-edge-metrics}"
SERVER_POD=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/name=server -o jsonpath='{.items[0].metadata.name}')

echo "=== Creating test device ==="
kubectl exec -n "$NAMESPACE" "$SERVER_POD" -- wget -qO- --post-data='{"device_type":"stub","ip_address":"10.0.0.99","port":9100,"reload_port":9101}' \
  --header='Content-Type: application/json' http://localhost:8081/config/test-device

echo "=== Verifying device in list ==="
kubectl exec -n "$NAMESPACE" "$SERVER_POD" -- wget -qO- http://localhost:8081/devices

echo "=== Deleting test device ==="
kubectl exec -n "$NAMESPACE" "$SERVER_POD" -- wget -qO- --method=DELETE http://localhost:8081/config/test-device

echo "=== Device lifecycle test PASSED ==="
