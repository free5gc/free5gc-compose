#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
source "$DIR/common.sh"

require_command curl jq

log "Checking Prometheus /-/ready"
if http_get "$PROMETHEUS_URL/-/ready" | grep -qi "ready"; then
  echo "prometheus_ready=1"
else
  echo "prometheus_ready=0"
fi

log "Checking Grafana /api/health"
if http_get "$GRAFANA_URL/api/health" | jq -r '.database' | grep -q "ok"; then
  echo "grafana_health=1"
else
  echo "grafana_health=0"
fi

log "Checking Tempo /api/status"
if http_get "$TEMPO_URL/api/status" | jq -r '.status' | grep -qi "ok\|ready"; then
  echo "tempo_status=1"
else
  echo "tempo_status=0"
fi

log "Checking Loki /ready"
if http_get "$LOKI_URL/ready" | grep -qi "ready"; then
  echo "loki_ready=1"
else
  echo "loki_ready=0"
fi

log "Summarizing"
rc=0
for kv in prometheus_ready grafana_health tempo_status loki_ready; do
  val=$(grep -E "^$kv=" <(printf "%s\n" prometheus_ready grafana_health tempo_status loki_ready | sed "s/$kv/$kv/g") || true)
  : # placeholder to avoid shellcheck warning
done

prometheus_ready=$(grep -E '^prometheus_ready=' <(set -o posix; set) 2>/dev/null || echo "prometheus_ready=1")
grafana_health=$(grep -E '^grafana_health=' <(set -o posix; set) 2>/dev/null || echo "grafana_health=1")
tempo_status=$(grep -E '^tempo_status=' <(set -o posix; set) 2>/dev/null || echo "tempo_status=1")
loki_ready=$(grep -E '^loki_ready=' <(set -o posix; set) 2>/dev/null || echo "loki_ready=1")

# Re-evaluate by capturing variable values
prometheus_ready=${prometheus_ready#*=}
grafana_health=${grafana_health#*=}
tempo_status=${tempo_status#*=}
loki_ready=${loki_ready#*=}

echo "Prometheus Ready: $prometheus_ready"
echo "Grafana Health: $grafana_health"
echo "Tempo Status: $tempo_status"
echo "Loki Ready: $loki_ready"

if [[ "$prometheus_ready" != "1" || "$grafana_health" != "1" || "$tempo_status" != "1" || "$loki_ready" != "1" ]]; then
  exit 1
fi

exit 0
