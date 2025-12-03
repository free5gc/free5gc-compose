#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
source "$DIR/common.sh"

require_command curl jq

log "Checking Grafana health"
http_get_grafana "/api/health" | jq '.'

log "List datasources"
ds=$(http_get_grafana "/api/datasources")
count=$(echo "$ds" | jq 'length')
echo "datasource_count=$count"
test "$count" -ge 1 || exit 1

log "List dashboards"
dash=$(http_get_grafana "/api/search?query=")
dcount=$(echo "$dash" | jq 'length')
echo "dashboard_count=$dcount"
test "$dcount" -ge 1 || exit 2

exit 0
