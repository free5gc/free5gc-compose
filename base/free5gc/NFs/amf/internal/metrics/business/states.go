package business

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/free5gc/util/metrics/utils"
)

var (
	// GmmStateGauge Gauge for number of UEs in each GMM state, labeled by state name (and access type)
	// labeled by AccessType and GMM state.
	GmmStateGauge *prometheus.GaugeVec
	// GmmTransitionCounter Counter for GMM state transitions,
	// labeled by from_states to to_state
	GmmTransitionCounter *prometheus.CounterVec
	// GmmStateDurationHist Histogram for time spent in a GMM state (seconds)
	// labeled by AccessType and GMM state.
	GmmStateDurationHist *prometheus.HistogramVec
)

func GetGMMStatesHandlerMetrics(namespace string) []prometheus.Collector {
	var collectors []prometheus.Collector

	GmmStateGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      GMM_STATE_GAUGE_NAME,
			Help:      GMM_STATE_GAUGE_DESC,
		},
		[]string{GMM_STATE_ACCESS_TYPE_LABEL, GMM_STATE_LABEL},
	)

	collectors = append(collectors, GmmStateGauge)

	GmmTransitionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      GMM_TRANSITION_COUNTER_NAME,
			Help:      GMM_TRANSITION_COUNTER_DESC,
		},
		[]string{GMM_STATE_FROM_STATE_LABEL, GMM_STATE_TO_STATE_LABEL},
	)

	collectors = append(collectors, GmmTransitionCounter)

	GmmStateDurationHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      GMM_DURATION_HISTOGRAM_NAME,
			Help:      GMM_DURATION_HISTOGRAM_DESC,
			Buckets: []float64{
				0.0001,
				0.0050,
				0.0100,
				0.0200,
				0.0250,
				0.0500,
			},
		},
		[]string{GMM_STATE_ACCESS_TYPE_LABEL, GMM_STATE_LABEL},
	)

	collectors = append(collectors, GmmStateDurationHist)

	return collectors
}

func IncrGmmTransitionCounter(fromState string, toState string) {
	if utils.IsBusinessMetricsEnabled() && IsGmmStateMetricsEnabled() {
		GmmTransitionCounter.With(prometheus.Labels{
			GMM_STATE_FROM_STATE_LABEL: fromState,
			GMM_STATE_TO_STATE_LABEL:   toState,
		}).Add(1)
	}
}

func IncrGmmStateGauge(accessType string, state string) {
	if utils.IsBusinessMetricsEnabled() && IsGmmStateMetricsEnabled() {
		GmmStateGauge.With(prometheus.Labels{
			GMM_STATE_ACCESS_TYPE_LABEL: accessType,
			GMM_STATE_LABEL:             state,
		}).Add(1)
	}
}

func DecrGmmStateGauge(accessType string, state string, enterTime time.Time) {
	if utils.IsBusinessMetricsEnabled() && IsGmmStateMetricsEnabled() {
		GmmStateGauge.With(prometheus.Labels{
			GMM_STATE_ACCESS_TYPE_LABEL: accessType,
			GMM_STATE_LABEL:             state,
		}).Dec()

		ObserveGmmTransitionDuration(accessType, state, enterTime)
	}
}

func ObserveGmmTransitionDuration(accessType string, gmmState string, enterTime time.Time) {
	if utils.IsBusinessMetricsEnabled() && IsGmmStateMetricsEnabled() {
		duration := time.Since(enterTime).Seconds()

		GmmStateDurationHist.With(prometheus.Labels{
			GMM_STATE_ACCESS_TYPE_LABEL: accessType,
			GMM_STATE_LABEL:             gmmState,
		}).Observe(duration)
	}
}
