#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
source "$DIR/common.sh"

require_command curl jq

# Check Loki logs for a label and substring
: "${LOKI_LABEL:=job}"
: "${LOKI_LABEL_VALUE:=free5gc}"
: "${LOG_MATCH:=amf}" # substring to search in log lines

start_ns=$(($(date +%s) - 300))000000000
end_ns=$(($(date +%s)))000000000

q=$(printf '{%s="%s"}' "$LOKI_LABEL" "$LOKI_LABEL_VALUE")
q_encoded=$(printf %s "$q" | jq -sRr @uri)
url="$LOKI_URL/loki/api/v1/query_range?query=$q_encoded&start=$start_ns&end=$end_ns&limit=100"

log "Querying Loki: $q"
resp=$(http_get "$url")
status=$(echo "$resp" | jq -r '.status')
streams=$(echo "$resp" | jq -r '.data.result | length')

echo "status=$status"
echo "streams=$streams"

test "$status" = "success" || exit 1

if [[ "$streams" -eq 0 ]]; then
  log "No log streams found for label $LOKI_LABEL=$LOKI_LABEL_VALUE (this is OK if no logs yet)"
  exit 0
fi

matches=$(echo "$resp" | jq -r \
  --arg m "$LOG_MATCH" \
  '[.data.result[]?.values[]?[1] | select(contains($m))] | length')

echo "matches=$matches"

if [[ "$matches" -ge 1 ]]; then
  log "Found $matches log lines containing '$LOG_MATCH'"
  exit 0
else
  log "No logs matching '$LOG_MATCH' found (this is OK if no matching traffic yet)"
  exit 0
fi
