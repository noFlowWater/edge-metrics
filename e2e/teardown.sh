#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-edge-metrics-e2e}"

echo "=== Deleting k3d cluster: $CLUSTER_NAME ==="
k3d cluster delete "$CLUSTER_NAME"

echo "=== Teardown complete ==="
