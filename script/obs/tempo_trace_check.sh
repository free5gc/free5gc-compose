#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
source "$DIR/common.sh"

require_command curl jq

# Simple check: search for traces by service name
: "${SERVICE_NAME:=free5gc}" # e.g., amf, smf, upf, webui

log "Searching Tempo traces for service=$SERVICE_NAME"
resp=$(http_get "$TEMPO_URL/api/search?tags=service.name=$SERVICE_NAME" || echo '{}')

count=$(echo "$resp" | jq '.traces // [] | length')
echo "trace_count=$count"

if [[ "$count" -ge 1 ]]; then
  echo "$resp" | jq '.traces[0]' || true
  exit 0
else
  log "No traces found for $SERVICE_NAME (this is OK if no traffic yet)"
  exit 0
fi
