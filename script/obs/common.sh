#!/usr/bin/env bash
set -euo pipefail

# Shared helpers for observability tests

: "${FREE5GC_HOST:=localhost}"
: "${PROMETHEUS_URL:=http://localhost:9090}"
: "${GRAFANA_URL:=http://localhost:3001}"
: "${TEMPO_URL:=http://localhost:3200}"
: "${LOKI_URL:=http://localhost:3100}"

# Grafana credentials (default admin/admin unless overridden)
: "${GRAFANA_USER:=admin}"
: "${GRAFANA_PASS:=admin}"

log() {
  echo "[obs] $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

require_command() {
  for cmd in "$@"; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      echo "Missing command: $cmd" >&2
      exit 127
    fi
  done
}

http_get() {
  local url=$1
  curl -fsSL --max-time 10 "$url"
}

http_get_grafana() {
  local path=$1
  curl -fsSL --max-time 10 -u "$GRAFANA_USER:$GRAFANA_PASS" "$GRAFANA_URL$path"
}

ok() { echo "OK"; }
fail() { echo "FAIL"; }
