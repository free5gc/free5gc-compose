package business

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/free5gc/util/metrics/utils"
)

var (
	// handoverInProgressGauge Gauge for the number of handover currently in progress in the AMF
	// labeled with handover type.
	handoverInProgressGauge *prometheus.GaugeVec
	// handoverEventCounter Counter for the different handover event (attempt, success, failure),
	// labeled with handover Type, handover Event and handoverCause
	handoverEventCounter *prometheus.CounterVec
	// handoverDuration Histogram for time spent doing a handover in seconds
	// labeled by handover Type and handover Event
	handoverDuration *prometheus.HistogramVec
)

func GetHandoverHandlerMetrics(namespace string) []prometheus.Collector {
	var collectors []prometheus.Collector

	handoverInProgressGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      HANDOVER_IN_PROGRESS_GAUGE_NAME,
			Help:      HANDOVER_IN_PROGRESS_GAUGE_DESC,
		},
		[]string{HANDOVER_TYPE_LABEL},
	)

	handoverInProgressGauge.With(prometheus.Labels{HANDOVER_TYPE_LABEL: HANDOVER_TYPE_XN_VALUE}).Set(0)
	handoverInProgressGauge.With(prometheus.Labels{HANDOVER_TYPE_LABEL: HANDOVER_TYPE_NGAP_VALUE}).Set(0)

	handoverEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      HANDOVER_EVENT_COUNTER_NAME,
			Help:      HANDOVER_EVENT_COUNTER_DESC,
		},
		[]string{HANDOVER_TYPE_LABEL, HANDOVER_EVENT_LABEL, HANDOVER_CAUSE_LABEL},
	)

	handoverDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: SUBSYSTEM_NAME,
			Name:      HANDOVER_DURATION_HISTOGRAM_NAME,
			Help:      HANDOVER_DURATION_HISTOGRAM_DESC,
			Buckets: []float64{
				0.0010,
				0.0050,
				0.0100,
				0.0500,
				0.1000,
				0.5000,
				1.0000,
			},
		},
		[]string{HANDOVER_TYPE_LABEL, HANDOVER_EVENT_LABEL},
	)

	collectors = append(collectors, handoverInProgressGauge, handoverEventCounter, handoverDuration)

	return collectors
}

func incrHoInProgressGauge(hoType string) {
	if utils.IsBusinessMetricsEnabled() && IsHandoverMetricsEnabled() {
		handoverInProgressGauge.With(prometheus.Labels{
			HANDOVER_TYPE_LABEL: hoType,
		}).Inc()
	}
}

func decrHoInProgressGauge(hoType string) {
	if utils.IsBusinessMetricsEnabled() && IsHandoverMetricsEnabled() {
		handoverInProgressGauge.With(prometheus.Labels{
			HANDOVER_TYPE_LABEL: hoType,
		}).Dec()
	}
}

func incrHoEventDurationCounter(handoverType string, handoverEvent string, hoStartTime time.Time) {
	if utils.IsBusinessMetricsEnabled() && IsHandoverMetricsEnabled() {
		if hoStartTime.IsZero() {
			return
		}

		duration := time.Since(hoStartTime).Seconds()

		handoverDuration.With(prometheus.Labels{
			HANDOVER_TYPE_LABEL:  handoverType,
			HANDOVER_EVENT_LABEL: handoverEvent,
		}).Observe(duration)
	}
}

func IncrHoEventCounter(handoverType string, handoverEvent string, handoverCause string, hoStartTime time.Time) {
	if utils.IsBusinessMetricsEnabled() && IsHandoverMetricsEnabled() {
		handoverEventCounter.With(prometheus.Labels{
			HANDOVER_TYPE_LABEL:  handoverType,
			HANDOVER_EVENT_LABEL: handoverEvent,
			HANDOVER_CAUSE_LABEL: handoverCause,
		}).Inc()

		switch handoverEvent {
		case HANDOVER_EVENT_ATTEMPT_VALUE:
			incrHoInProgressGauge(handoverType)
		case utils.FailureMetric:
			fallthrough
		case utils.SuccessMetric:
			incrHoEventDurationCounter(handoverType, handoverEvent, hoStartTime)
			decrHoInProgressGauge(handoverType)
		}
	}
}
