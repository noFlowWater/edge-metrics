#!/bin/sh
set -e

# shelly_server 백그라운드 기동
if [ "${SHELLY_ENABLED:-true}" = "true" ]; then
    python3 /app/shelly_server.py &
    SHELLY_PID=$!
    echo "Shelly server started (PID: $SHELLY_PID)"
fi

# exporter 포그라운드 기동
exec python3 /app/exporter.py
