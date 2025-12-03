#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"

require_command() {
  for cmd in "$@"; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      echo "Missing command: $cmd" >&2
      exit 127
    fi
  done
}

require_command docker

# Precheck: ensure required Free5GC and observability containers exist and are running
echo "==== Precheck: containers existence ===="

obs_names=(prometheus grafana tempo loki)
fivegc_names=(amf smf nrf pcf udm udr webui)

missing=()

for name in "${obs_names[@]}"; do
  if ! docker ps --format '{{.Names}} {{.Status}}' | grep -iE "${name}" >/dev/null 2>&1; then
    missing+=("$name (observ)")
  fi
done

for name in "${fivegc_names[@]}"; do
  if ! docker ps --format '{{.Names}} {{.Status}}' | grep -iE "${name}" >/dev/null 2>&1; then
    missing+=("$name (5gc)")
  fi
done

# Optional: check for UPF (may not exist in all setups)
if docker ps --format '{{.Names}} {{.Status}}' | grep -iE "upf" >/dev/null 2>&1; then
  echo "UPF container detected"
else
  echo "Warning: UPF container not found (optional, skipping)" >&2
fi

if [[ ${#missing[@]} -gt 0 ]]; then
  echo "Missing or not running containers:" >&2
  for m in "${missing[@]}"; do echo "- $m" >&2; done
  echo "Hint: start stack via commands below: 
        bash ./script/build.sh aarch64
        docker compose -f docker-compose-prometheus.yaml up -d
        docker compose -f docker-compose-build.yaml up -d
        or
        docker compose -f docker-compose-build.yaml -f docker-compose-prometheus.yaml up -d
        " >&2
  exit 1
fi

tests=(
  "observability_smoke.sh"
  "grafana_api_check.sh"
  "prom_metrics_check.sh"
  "tempo_trace_check.sh"
  "loki_log_check.sh"
)

pass=0
fail=0

for t in "${tests[@]}"; do
  echo "==== Running $t ===="
  if "$DIR/$t"; then
    echo "==== $t: PASS ===="
    pass=$((pass+1))
  else
    echo "==== $t: FAIL ===="
    fail=$((fail+1))
  fi
done

echo "Passed: $pass"
echo "Failed: $fail"

if [[ $fail -gt 0 ]]; then
  exit 1
fi

exit 0
