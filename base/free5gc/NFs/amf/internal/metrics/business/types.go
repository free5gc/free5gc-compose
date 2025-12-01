package business

import "github.com/free5gc/openapi/models"

// Global metric information
const (
	SUBSYSTEM_NAME          = "amf_business"
	CM_STATE_METRICS        = "cm-state"
	HANDOVER_METRICS        = "handover"
	PDU_METRICS             = "pdu"
	GMM_STATE_METRICS       = "gmm-state"
	UE_CONNECTIVITY_METRICS = "ue-connectivity"
)

// Collectors information
const (
	GMM_STATE_GAUGE_NAME        = "ue_gmm_state_count"
	GMM_STATE_GAUGE_DESC        = "Current number of UEs in each 5GMM state in the AMF"
	GMM_TRANSITION_COUNTER_NAME = "ue_gmm_transitions_total"
	GMM_TRANSITION_COUNTER_DESC = "Count of UE GMM state transitions in the AMF"
	GMM_DURATION_HISTOGRAM_NAME = "ue_gmm_state_duration_seconds"
	GMM_DURATION_HISTOGRAM_DESC = "Duration that UEs spend in a given GMM state before transitioning"

	UE_CM_STATE_GAUGE_NAME = "ue_cm_gmm_state_count"
	UE_CM_STATE_GAUGE_DESC = "Count of the UE in each Connection Management State (CM_IDLE, CM_CONNECTED) in the AMF"

	HANDOVER_IN_PROGRESS_GAUGE_NAME  = "handover_current_count"
	HANDOVER_IN_PROGRESS_GAUGE_DESC  = "Number of UEs currently in handover procedure (source AMF side)"
	HANDOVER_EVENT_COUNTER_NAME      = "handover_events_total"
	HANDOVER_EVENT_COUNTER_DESC      = "Count of handover events (attempts, successes, failures)"
	HANDOVER_DURATION_HISTOGRAM_NAME = "handover_duration_seconds"
	HANDOVER_DURATION_HISTOGRAM_DESC = "Histogram of the handover duration in seconds"

	PDU_SESSION_ACTIVE_GAUGE_NAME  = "active_pdu_session_current_count"
	PDU_SESSION_ACTIVE_GAUGE_DESC  = "Number of PDU Session currently active (source AMF side)"
	PDU_SESSION_EVENT_COUNTER_NAME = "pdu_session_events_total"
	PDU_SESSION_EVENT_COUNTER_DESC = "Count of pdu events (setup, release, modification)"

	UE_CONNECTIVITY_GAUGE_NAME = "ue_connectivity"
	UE_CONNECTIVITY_GAUGE_DESC = "Number of user equipment that are connected to the core network " +
		"(cm-connected + gmm-registered)"
)

// Label names
const (
	// States
	GMM_STATE_ACCESS_TYPE_LABEL = "access_type"
	GMM_STATE_LABEL             = "gmm_state"
	GMM_STATE_FROM_STATE_LABEL  = "from_state"
	GMM_STATE_TO_STATE_LABEL    = "to_state"

	// Connection Management
	UE_CM_STATE_LABEL       = "state"
	UE_CM_ACCESS_TYPE_LABEL = "access_type"

	// Handover
	HANDOVER_TYPE_LABEL  = "type"
	HANDOVER_EVENT_LABEL = "event"
	// HANDOVER_CAUSE_LABEL stores the potential error cause for a handover event
	HANDOVER_CAUSE_LABEL = "cause"

	// PDU
	PDU_SESSION_ACCESS_TYPE_LABEL = "access_type"
	PDU_SESSION_EVENT_LABEL       = "event"

	// UE-Connectivity
	UE_CONNECTIVITY_ACCESS_TYPE_LABEL = "access_type"
)

// Metrics Values
const (
	// Connection Management
	UE_CM_IDLE_VALUE      = "cm-idle"
	UE_CM_CONNECTED_VALUE = "cm-connected"

	// Handover
	HANDOVER_TYPE_XN_VALUE       = "xn"
	HANDOVER_TYPE_NGAP_VALUE     = "ngap"
	HANDOVER_EVENT_ATTEMPT_VALUE = "attempt"

	PDU_SESSION_CREATION_EVENT = "creation"
	PDU_SESSION_RELEASE_EVENT  = "release"
)

// Potential Causes
const (
	HANDOVER_RAN_UE_MISSING_ERR                        = "ran ue missing"
	HANDOVER_AMF_UE_MISSING_ERR                        = "amf ue missing"
	HANDOVER_TARGET_UE_MISSING_ERR                     = "target ue is missing"
	HANDOVER_SECURITY_CONTEXT_MISSING_ERR              = "security context missing"
	HANDOVER_SWITCH_RAN_ERR                            = "ue could not switch ran"
	HANDOVER_TARGET_ID_NOT_SUPPORTED_ERR               = "target id type is not supported"
	HANDOVER_PDU_SESSION_RES_REL_LIST_ERR              = "some pdu session could not been release for handover"
	HANDOVER_BETWEEN_DIFFERENT_AMF_NOT_SUPPORTED       = "handover between different amf has not been implemented yet"
	HANDOVER_NOT_YET_IMPLEMENT_N2_HANDOVER_BETWEEN_AMF = "n2 Handover between amf has not been implemented yet"
	HANDOVER_EMPTY_CAUSE                               = ""
)

var AccessTypes = []string{string(models.AccessType__3_GPP_ACCESS), string(models.AccessType_NON_3_GPP_ACCESS)}

var handoverMetricsEnabled bool

func IsHandoverMetricsEnabled() bool {
	return handoverMetricsEnabled
}

func EnableHandoverMetrics() {
	handoverMetricsEnabled = true
}

var ueCmMetricsEnabled bool

func IsUeCmMetricsEnabled() bool {
	return ueCmMetricsEnabled
}

func EnableUeCmMetrics() {
	ueCmMetricsEnabled = true
}

var pduMetricsEnabled bool

func IsPduMetricsEnabled() bool {
	return pduMetricsEnabled
}

func EnablePduMetrics() {
	pduMetricsEnabled = true
}

var gmmStateMetricsEnabled bool

func IsGmmStateMetricsEnabled() bool {
	return gmmStateMetricsEnabled
}

func EnableGmmStateMetrics() {
	gmmStateMetricsEnabled = true
}

var ueConnectivityMetricsEnabled bool

func IsUeConnectivityMetricsEnabled() bool {
	return ueConnectivityMetricsEnabled
}

func EnableUeConnectivityMetrics() {
	ueConnectivityMetricsEnabled = true
}
