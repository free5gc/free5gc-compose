#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
source "$DIR/common.sh"

require_command curl jq

# Query example: check that free5gc-related metrics exist
: "${PROM_QUERY:=up}"

log "Querying Prometheus: $PROM_QUERY"
resp=$(http_get "$PROMETHEUS_URL/api/v1/query?query=$(printf %s "$PROM_QUERY" | sed 's/\+/\%2B/g')")
status=$(echo "$resp" | jq -r '.status')
result_count=$(echo "$resp" | jq -r '.data.result | length')

echo "status=$status"
echo "result_count=$result_count"

test ${status} = "success" || exit 1
test ${result_count} -ge 1 || exit 2

log "First metric sample"
echo "$resp" | jq '.data.result[0]'

exit 0
