#!/bin/bash
# Jetson systemd → DaemonSet 마이그레이션 스크립트
# 사용법: ./scripts/migrate-jetson.sh
# 각 Jetson 노드의 기존 config를 /var/lib/edge-metrics/로 복사 후 systemd 중지

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 노드별 SSH 호스트 + config 경로
declare -A NODE_SSH=(
    [jetsono]="orin"
    [jetson-nano]="jetson"
    [jetsonx]="xavier"
)

declare -A NODE_CONFIG=(
    [jetsono]="/home/orin/ETRI/edge-metrics-exporter/config.yaml"
    [jetson-nano]="/home/jetson/ETRI/edge-metrics-exporter/config.yaml"
    [jetsonx]="/home/dev/ETRI/edge-metrics-exporter/config.yaml"
)

TARGET_DIR="/var/lib/edge-metrics"

echo -e "${GREEN}=== Jetson systemd -> DaemonSet 마이그레이션 ===${NC}"
echo ""

for NODE in jetsono jetson-nano jetsonx; do
    SSH_HOST="${NODE_SSH[$NODE]}"
    CONFIG_SRC="${NODE_CONFIG[$NODE]}"

    echo -e "${YELLOW}[$NODE] (ssh: $SSH_HOST)${NC}"

    # 1. config 복사
    echo "  [1/3] config 복사: $CONFIG_SRC -> $TARGET_DIR/config.yaml"
    if ssh "$SSH_HOST" "sudo mkdir -p $TARGET_DIR && sudo cp $CONFIG_SRC $TARGET_DIR/config.yaml" 2>/dev/null; then
        echo -e "  ${GREEN}OK${NC}"
    else
        echo -e "  ${RED}FAILED — 수동 복사 필요${NC}"
        continue
    fi

    # 2. systemd 중지
    echo "  [2/3] systemd 서비스 중지..."
    if ssh "$SSH_HOST" "sudo systemctl stop edge-metrics-exporter shelly-server" 2>/dev/null; then
        echo -e "  ${GREEN}OK${NC}"
    else
        echo -e "  ${RED}FAILED${NC}"
        continue
    fi

    # 3. 포트 해제 확인
    echo "  [3/3] 포트 해제 확인..."
    PORTS=$(ssh "$SSH_HOST" "ss -tlnp 2>/dev/null | grep -cE ':9102|:9101|:8765|:8766'" 2>/dev/null || echo "0")
    if [ "$PORTS" = "0" ]; then
        echo -e "  ${GREEN}포트 해제 완료${NC}"
    else
        echo -e "  ${RED}경고: $PORTS 개 포트 아직 사용 중${NC}"
    fi

    echo ""
done

echo -e "${GREEN}마이그레이션 준비 완료. DaemonSet 배포:${NC}"
echo "  ./scripts/deploy.sh <VERSION>"
