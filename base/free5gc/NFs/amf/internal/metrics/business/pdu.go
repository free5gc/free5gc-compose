package business

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/utils"
)

var (
	// pduSessionActiveGauge Gauge for the number of PDU Session that are currently active in the AMF, labeled by
	// AccessType
	pduSessionActiveGauge *prometheus.GaugeVec
	// pduSessionEventCounter Counter for PDU Session event (creation, release), labeled by AccessType and type of event.
	pduSessionEventCounter *prometheus.CounterVec
)

func GetPDUHandlerMetrics(namespace string) []prometheus.Collector {
	var collectors []prometheus.Collector

	pduSessionActiveGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      PDU_SESSION_ACTIVE_GAUGE_NAME,
			Help:      PDU_SESSION_ACTIVE_GAUGE_DESC,
		}, []string{PDU_SESSION_ACCESS_TYPE_LABEL},
	)

	pduSessionActiveGauge.With(prometheus.Labels{
		PDU_SESSION_ACCESS_TYPE_LABEL: string(models.AccessType__3_GPP_ACCESS),
	}).Set(0)
	pduSessionActiveGauge.With(prometheus.Labels{
		PDU_SESSION_ACCESS_TYPE_LABEL: string(models.AccessType_NON_3_GPP_ACCESS),
	}).Set(0)

	collectors = append(collectors, pduSessionActiveGauge)

	pduSessionEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      PDU_SESSION_EVENT_COUNTER_NAME,
			Help:      PDU_SESSION_EVENT_COUNTER_DESC,
		}, []string{PDU_SESSION_ACCESS_TYPE_LABEL, PDU_SESSION_EVENT_LABEL},
	)

	collectors = append(collectors, pduSessionEventCounter)

	return collectors
}

func incrActivePduSession(accessType string) {
	if utils.IsBusinessMetricsEnabled() && IsPduMetricsEnabled() {
		pduSessionActiveGauge.With(prometheus.Labels{PDU_SESSION_ACCESS_TYPE_LABEL: accessType}).Inc()
	}
}

func decrActivePduSession(accessType string) {
	if utils.IsBusinessMetricsEnabled() && IsPduMetricsEnabled() {
		pduSessionActiveGauge.With(prometheus.Labels{PDU_SESSION_ACCESS_TYPE_LABEL: accessType}).Dec()
	}
}

func IncrPduSessionEventCounter(accessType string, event string) {
	if utils.IsBusinessMetricsEnabled() && IsPduMetricsEnabled() {
		switch event {
		case PDU_SESSION_CREATION_EVENT:
			incrActivePduSession(accessType)
		case PDU_SESSION_RELEASE_EVENT:
			decrActivePduSession(accessType)
		}

		pduSessionEventCounter.With(prometheus.Labels{
			PDU_SESSION_ACCESS_TYPE_LABEL: accessType,
			PDU_SESSION_EVENT_LABEL:       event,
		}).Inc()
	}
}
