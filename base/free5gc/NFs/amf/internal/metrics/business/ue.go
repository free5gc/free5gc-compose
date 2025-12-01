package business

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/utils"
)

// ueConnectedGauge The UE connectivity to the coreNetwork.
// It is considered Connected when it is cm-connected and gmm.Registered
var ueConnectedGauge *prometheus.GaugeVec

func GetUEHandlerMetrics(namespace string) []prometheus.Collector {
	var collectors []prometheus.Collector

	ueConnectedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      UE_CONNECTIVITY_GAUGE_NAME,
			Help:      UE_CONNECTIVITY_GAUGE_DESC,
		},
		[]string{UE_CONNECTIVITY_ACCESS_TYPE_LABEL},
	)

	initUEConnectivityGauge()

	collectors = append(collectors, ueConnectedGauge)

	return collectors
}

func initUEConnectivityGauge() {
	for _, accessType := range AccessTypes {
		ueConnectedGauge.With(prometheus.Labels{UE_CONNECTIVITY_ACCESS_TYPE_LABEL: accessType}).Set(0)
	}
}

func IncrUeConnectivityGauge(accessType models.AccessType) {
	if utils.IsBusinessMetricsEnabled() && IsUeConnectivityMetricsEnabled() {
		ueConnectedGauge.With(prometheus.Labels{UE_CONNECTIVITY_ACCESS_TYPE_LABEL: string(accessType)}).Inc()
	}
}

func DecrUeConnectivityGauge(accessType models.AccessType) {
	if utils.IsBusinessMetricsEnabled() && IsUeConnectivityMetricsEnabled() {
		ueConnectedGauge.With(prometheus.Labels{UE_CONNECTIVITY_ACCESS_TYPE_LABEL: string(accessType)}).Dec()
	}
}
