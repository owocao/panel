#!/usr/bin/env bash
set -euo pipefail

PORT="${BIU_PANEL_SMOKE_PORT:-55188}"
DATA_DIR="${BIU_PANEL_SMOKE_DATA:-/tmp/biu-panel-smoke}"
STATIC_DIR="${BIU_PANEL_SMOKE_STATIC:-$(pwd)/frontend/dist}"
COOKIE_FILE="${BIU_PANEL_SMOKE_COOKIE:-/tmp/biu-panel-smoke-cookies.txt}"

rm -rf "$DATA_DIR" "$COOKIE_FILE"
(cd frontend && npm run build >/dev/null)
(
  cd backend
  BIU_PANEL_DATA_DIR="$DATA_DIR" \
  BIU_PANEL_STATIC_DIR="$STATIC_DIR" \
  BIU_PANEL_PORT="$PORT" \
  BIU_PANEL_ADMIN_USER=admin \
  BIU_PANEL_ADMIN_PASSWORD=password123 \
  go run ./cmd/server
) &
PID=$!
trap 'kill "$PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 40); do
  if curl -fsS "http://127.0.0.1:$PORT/api/health" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -fsS "http://127.0.0.1:$PORT/" >/dev/null
curl -fsS -c "$COOKIE_FILE" -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"password123"}' \
  "http://127.0.0.1:$PORT/api/auth/login" >/dev/null
curl -fsS -b "$COOKIE_FILE" -H 'Content-Type: application/json' \
  -d '{"name":"默认","sort":1}' \
  "http://127.0.0.1:$PORT/api/navigation/groups" >/dev/null
curl -fsS -b "$COOKIE_FILE" "http://127.0.0.1:$PORT/api/navigation" | grep -q '"items":\[\]'

echo "Smoke test passed on port $PORT"
