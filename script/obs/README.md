# Observability Test Scripts

These scripts provide smoke tests for a Free5GC deployment with Prometheus, Grafana, Tempo, and Loki.

Prerequisites:
- Services running via `docker-compose.yaml` or `docker-compose-prometheus.yaml` with default ports.
- `curl` and `jq` installed on the host.

Environment variables:
- `PROMETHEUS_URL` (default `http://localhost:9090`)
- `GRAFANA_URL` (default `http://localhost:3000`)
- `TEMPO_URL` (default `http://localhost:3200`)
- `LOKI_URL` (default `http://localhost:3100`)
- `GRAFANA_USER`/`GRAFANA_PASS` (default `admin`/`admin`)

Usage:
```bash
chmod +x script/obs/*.sh

# Run all tests
script/obs/run_all.sh

# Individual checks
script/obs/observability_smoke.sh
PROM_QUERY='up{job="free5gc"}' script/obs/prom_metrics_check.sh
SERVICE_NAME=amf script/obs/tempo_trace_check.sh
LOKI_LABEL=job LOKI_LABEL_VALUE=free5gc LOG_MATCH=amf script/obs/loki_log_check.sh
GRAFANA_USER=admin GRAFANA_PASS=admin script/obs/grafana_api_check.sh
```

Exit codes are non-zero if checks fail, suitable for CI.
