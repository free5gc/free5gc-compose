package message

import (
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/sctp"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/context"
	ngap_metrics "github.com/free5gc/util/metrics/ngap"
)

var (
	ngaplog    *logrus.Entry
	emptyCause = ngapType.Cause{Present: 0}
)

func init() {
	ngaplog = logger.NgapLog
}

func SendToAmf(amf *context.TNGFAMF, pkt []byte) (bool, string) {
	if amf == nil {
		ngaplog.Errorf("[TNGF] AMF Context is nil ")
		return false, "AMF Context is nil"
	} else {
		if n, err := amf.SCTPConn.Write(pkt); err != nil {
			ngaplog.Errorf("Write to SCTP socket failed: %+v", err)
			return false, ngap_metrics.SCTP_SOCKET_WRITE_ERR
		} else {
			ngaplog.Tracef("Wrote %d bytes", n)
		}
	}
	return true, ""
}

func SendNGSetupRequest(conn *sctp.SCTPConn) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NgapLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	isNGSetupReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_SETUP_REQUEST, &isNGSetupReqSent, emptyCause, &additionalCause)

	ngaplog.Infoln("[TNGF] Send NG Setup Request")

	sctpAddr := conn.RemoteAddr().String()

	if available, _ := context.TNGFSelf().AMFReInitAvailableListLoad(sctpAddr); !available {
		ngaplog.Warnf("[TNGF] Please Wait at least for the indicated time before reinitiating toward same AMF[%s]", sctpAddr)
		return
	}
	pkt, err := BuildNGSetupRequest()
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build NGSetup Request failed: %+v\n", err)
		return
	}

	if n, write_err := conn.Write(pkt); write_err != nil {
		ngaplog.Errorf("Write to SCTP socket failed: %+v", write_err)
		additionalCause = ngap_metrics.SCTP_SOCKET_WRITE_ERR
	} else {
		ngaplog.Tracef("Wrote %d bytes", n)
		isNGSetupReqSent = true
	}
}

// partOfNGInterface: if reset type is "reset all", set it to nil TS 38.413 9.2.6.11
func SendNGReset(
	amf *context.TNGFAMF,
	cause ngapType.Cause,
	partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
) {
	isNGResetSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_RESET, &isNGResetSent, cause, &additionalCause)

	ngaplog.Infoln("[TNGF] Send NG Reset")

	pkt, err := BuildNGReset(cause, partOfNGInterface)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build NGReset failed : %s", err.Error())
		return
	}

	isNGResetSent, additionalCause = SendToAmf(amf, pkt)
}

func SendNGResetAcknowledge(
	amf *context.TNGFAMF,
	partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send NG Reset Acknowledge")

	isNGResetAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_RESET_ACKNOWLEDGE,
		&isNGResetAckSent, emptyCause, &additionalCause)

	if partOfNGInterface != nil && len(partOfNGInterface.List) == 0 {
		ngaplog.Error("length of partOfNGInterface is 0")
		additionalCause = ngap_metrics.NF_INTERFACE_LEN_ZERO_ERR
		return
	}

	pkt, err := BuildNGResetAcknowledge(partOfNGInterface, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build NGReset Acknowledge failed : %s", err.Error())
		return
	}

	isNGResetAckSent, additionalCause = SendToAmf(amf, pkt)
}

func SendInitialContextSetupResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	responseList *ngapType.PDUSessionResourceSetupListCxtRes,
	failedList *ngapType.PDUSessionResourceFailedToSetupListCxtRes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send Initial Context Setup Response")

	isInitialCtxRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_CONTEXT_SETUP_RESPONSE,
		&isInitialCtxRespSent, emptyCause, &additionalCause)

	if responseList != nil && len(responseList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngaplog.Errorln("Pdu List out of range")
		return
	}

	if failedList != nil && len(failedList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngaplog.Errorln("Pdu List out of range")
		return
	}

	pkt, err := BuildInitialContextSetupResponse(ue, responseList, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Initial Context Setup Response failed : %+v\n", err)
		return
	}

	isInitialCtxRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendInitialContextSetupFailure(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	cause ngapType.Cause,
	failedList *ngapType.PDUSessionResourceFailedToSetupListCxtFail,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send Initial Context Setup Failure")

	isInitialCtxFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_CONTEXT_SETUP_RESPONSE,
		&isInitialCtxFailureSent, cause, &additionalCause)

	if failedList != nil && len(failedList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngaplog.Errorln("Pdu List out of range")
		return
	}

	pkt, err := BuildInitialContextSetupFailure(ue, cause, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Initial Context Setup Failure failed : %+v\n", err)
		return
	}

	isInitialCtxFailureSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUEContextModificationResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send UE Context Modification Response")

	isUECtxModificationRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_MODIFICATION_RESPONSE,
		&isUECtxModificationRespSent, emptyCause, &additionalCause)

	pkt, err := BuildUEContextModificationResponse(ue, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build UE Context Modification Response failed : %+v\n", err)
		return
	}

	isUECtxModificationRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUEContextModificationFailure(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	cause ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send UE Context Modification Failure")

	isUECtxModificationFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_MODIFICATION_FAILURE,
		&isUECtxModificationFailureSent, cause, &additionalCause)

	pkt, err := BuildUEContextModificationFailure(ue, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build UE Context Modification Failure failed : %+v\n", err)
		return
	}

	isUECtxModificationFailureSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUEContextReleaseComplete(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send UE Context Release Complete")

	isUECtxReleaseCompleteSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_RELEASE_COMPLETE,
		&isUECtxReleaseCompleteSent, emptyCause, &additionalCause)

	pkt, err := BuildUEContextReleaseComplete(ue, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build UE Context Release Complete failed : %+v\n", err)
		return
	}

	isUECtxReleaseCompleteSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUEContextReleaseRequest(
	amf *context.TNGFAMF,
	ue *context.TNGFUe, cause ngapType.Cause,
) {
	ngaplog.Infoln("[TNGF] Send UE Context Release Request")

	isUECtxReleaseReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_RELEASE_REQUEST,
		&isUECtxReleaseReqSent, cause, &additionalCause)

	pkt, err := BuildUEContextReleaseRequest(ue, cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build UE Context Release Request failed : %+v\n", err)
		return
	}

	isUECtxReleaseReqSent, additionalCause = SendToAmf(amf, pkt)
}

func SendInitialUEMessage(amf *context.TNGFAMF,
	ue *context.TNGFUe, nasPdu []byte,
) {
	ngaplog.Infoln("[TNGF] Send Initial UE Message")
	// Attach To AMF

	isInitialUEMessageSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_UE_MESSAGE,
		&isInitialUEMessageSent, emptyCause, &additionalCause)

	pkt, err := BuildInitialUEMessage(ue, nasPdu, nil)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Initial UE Message failed : %+v\n", err)
		return
	}

	isInitialUEMessageSent, additionalCause = SendToAmf(amf, pkt)
	// ue.AttachAMF()
}

func SendUplinkNASTransport(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	nasPdu []byte,
) {
	ngaplog.Infoln("[TNGF] Send Uplink NAS Transport")

	isUplinkNasTransportSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UPLINK_NAS_TRANSPORT,
		&isUplinkNasTransportSent, emptyCause, &additionalCause)

	if len(nasPdu) == 0 {
		additionalCause = ngap_metrics.NAS_PDU_NIL_ERR
		ngaplog.Errorln("NAS Pdu is nil")
		return
	}

	pkt, err := BuildUplinkNASTransport(ue, nasPdu)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Uplink NAS Transport failed : %+v\n", err)
		return
	}

	isUplinkNasTransportSent, additionalCause = SendToAmf(amf, pkt)
}

func SendNASNonDeliveryIndication(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	nasPdu []byte,
	cause ngapType.Cause,
) {
	ngaplog.Infoln("[TNGF] Send NAS NonDelivery Indication")

	isNasNonDeliveryIndicationSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NAS_NON_DELIVERY_INDICATION,
		&isNasNonDeliveryIndicationSent, cause, &additionalCause)

	if len(nasPdu) == 0 {
		additionalCause = ngap_metrics.NAS_PDU_NIL_ERR
		ngaplog.Errorln("NAS Pdu is nil")
		return
	}

	pkt, err := BuildNASNonDeliveryIndication(ue, nasPdu, cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Uplink NAS Transport failed : %+v\n", err)
		return
	}

	isNasNonDeliveryIndicationSent, additionalCause = SendToAmf(amf, pkt)
}

func SendRerouteNASRequest() {
	ngaplog.Infoln("[TNGF] Send Reroute NAS Request")
}

func SendPDUSessionResourceSetupResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	responseList *ngapType.PDUSessionResourceSetupListSURes,
	failedListSURes *ngapType.PDUSessionResourceFailedToSetupListSURes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send PDU Session Resource Setup Response")

	isPduSessionResourceSetupRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_SETUP_RESPONSE,
		&isPduSessionResourceSetupRespSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngaplog.Error("UE context is nil, this information is mandatory.")
		return
	}

	pkt, err := BuildPDUSessionResourceSetupResponse(ue, responseList, failedListSURes, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build PDU Session Resource Setup Response failed : %+v", err)
		return
	}

	isPduSessionResourceSetupRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendPDUSessionResourceModifyResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	responseList *ngapType.PDUSessionResourceModifyListModRes,
	failedList *ngapType.PDUSessionResourceFailedToModifyListModRes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send PDU Session Resource Modify Response")

	isPduSessionResourceModifyRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_MODIFY_RESPONSE,
		&isPduSessionResourceModifyRespSent, emptyCause, &additionalCause)

	if ue == nil && criticalityDiagnostics == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngaplog.Error("UE context is nil, this information is mandatory")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyResponse(ue, responseList, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build PDU Session Resource Modify Response failed : %+v", err)
		return
	}

	isPduSessionResourceModifyRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendPDUSessionResourceModifyIndication(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	modifyList []ngapType.PDUSessionResourceModifyItemModInd,
) {
	ngaplog.Infoln("[TNGF] Send PDU Session Resource Modify Indication")

	isPduSessionResourceModifyIndicationSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_MODIFY_INDICATION,
		&isPduSessionResourceModifyIndicationSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngaplog.Error("UE context is nil, this information is mandatory")
		return
	}
	if modifyList == nil {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_MODIFY_LIST_NIL_ERR
		ngaplog.Errorln("PDU Session Resource Modify Indication List is nil. This message shall contain at least one Item")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyIndication(ue, modifyList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build PDU Session Resource Modify Indication failed : %+v", err)
		return
	}

	isPduSessionResourceModifyIndicationSent, additionalCause = SendToAmf(amf, pkt)
}

func SendPDUSessionResourceNotify(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	notiList *ngapType.PDUSessionResourceNotifyList,
	relList *ngapType.PDUSessionResourceReleasedListNot,
) {
	ngaplog.Infoln("[TNGF] Send PDU Session Resource Notify")

	isPduSessionResourceNotifySent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_NOTIFY,
		&isPduSessionResourceNotifySent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngaplog.Error("UE context is nil, this information is mandatory")
		return
	}

	pkt, err := BuildPDUSessionResourceNotify(ue, notiList, relList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build PDUSession Resource Notify failed : %+v", err)
		return
	}

	isPduSessionResourceNotifySent, additionalCause = SendToAmf(amf, pkt)
}

func SendPDUSessionResourceReleaseResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	relList ngapType.PDUSessionResourceReleasedListRelRes,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send PDU Session Resource Release Response")

	isPduSessionResourceReleaseRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_RELEASE_RESPONSE,
		&isPduSessionResourceReleaseRespSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngaplog.Error("UE context is nil, this information is mandatory")
		return
	}
	if len(relList.List) < 1 {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_RELEASED_LIST_NIL_ERR
		ngaplog.Errorln("PDUSessionResourceReleasedListRelRes is nil. This message shall contain at least one Item")
		return
	}

	pkt, err := BuildPDUSessionResourceReleaseResponse(ue, relList, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build PDU Session Resource Release Response failed : %+v", err)
		return
	}

	isPduSessionResourceReleaseRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendErrorIndication(
	amf *context.TNGFAMF,
	amfUENGAPID *int64,
	ranUENGAPID *int64,
	cause *ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send Error Indication")

	isErrorIndicationSent := false
	additionalCause := ""
	if cause != nil {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION,
			&isErrorIndicationSent, *cause, &additionalCause)
	} else {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION,
			&isErrorIndicationSent, emptyCause, &additionalCause)
	}

	if (cause == nil) && (criticalityDiagnostics == nil) {
		additionalCause = ngap_metrics.ERROR_INDICATION_CAUSE_AND_CRITICALITY_NIL_ERR
		ngaplog.Errorln("Both cause and criticality is nil. This message shall contain at least one of them.")
		return
	}

	pkt, err := BuildErrorIndication(amfUENGAPID, ranUENGAPID, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Error Indication failed : %+v\n", err)
		return
	}

	isErrorIndicationSent, additionalCause = SendToAmf(amf, pkt)
}

func SendErrorIndicationWithSctpConn(
	sctpConn *sctp.SCTPConn,
	amfUENGAPID *int64,
	ranUENGAPID *int64,
	cause *ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send Error Indication")

	isErrorIndicationWithSctpConnSent := false
	additionalCause := ""
	if cause != nil {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION_WITH_SCTP_CONN,
			&isErrorIndicationWithSctpConnSent, *cause, &additionalCause)
	} else {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION_WITH_SCTP_CONN,
			&isErrorIndicationWithSctpConnSent, emptyCause, &additionalCause)
	}

	if (cause == nil) && (criticalityDiagnostics == nil) {
		additionalCause = ngap_metrics.ERROR_INDICATION_CAUSE_AND_CRITICALITY_NIL_ERR
		ngaplog.Errorln("Both cause and criticality is nil. This message shall contain at least one of them.")
		return
	}

	pkt, err := BuildErrorIndication(amfUENGAPID, ranUENGAPID, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build Error Indication failed : %+v\n", err)
		return
	}

	if n, write_err := sctpConn.Write(pkt); write_err != nil {
		additionalCause = ngap_metrics.SCTP_SOCKET_WRITE_ERR
		ngaplog.Errorf("Write to SCTP socket failed: %+v", write_err)
	} else {
		isErrorIndicationWithSctpConnSent = true
		ngaplog.Tracef("Wrote %d bytes", n)
	}
}

func SendUERadioCapabilityInfoIndication() {
	ngaplog.Infoln("[TNGF] Send UE Radio Capability Info Indication")
}

func SendUERadioCapabilityCheckResponse(
	amf *context.TNGFAMF,
	ue *context.TNGFUe,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send UE Radio Capability Check Response")

	isUeRadioCapabilityCheckRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_RADIO_CAPABILITY_CHECK_RESPONSE,
		&isUeRadioCapabilityCheckRespSent, emptyCause, &additionalCause)

	pkt, err := BuildUERadioCapabilityCheckResponse(ue, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build UERadio Capability Check Response failed : %+v\n", err)
		return
	}
	isUeRadioCapabilityCheckRespSent, additionalCause = SendToAmf(amf, pkt)
}

func SendAMFConfigurationUpdateAcknowledge(
	amf *context.TNGFAMF,
	setupList *ngapType.AMFTNLAssociationSetupList,
	failList *ngapType.TNLAssociationList,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send AMF Configuration Update Acknowledge")

	isAmfConfUpdateAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.AMF_CONFIGURATION_UPDATE_ACKNOWLEDGE,
		&isAmfConfUpdateAckSent, emptyCause, &additionalCause)

	pkt, err := BuildAMFConfigurationUpdateAcknowledge(setupList, failList, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build AMF Configuration Update Acknowledge failed : %+v\n", err)
		return
	}

	isAmfConfUpdateAckSent, additionalCause = SendToAmf(amf, pkt)
}

func SendAMFConfigurationUpdateFailure(
	amf *context.TNGFAMF,
	ngCause ngapType.Cause,
	time *ngapType.TimeToWait,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngaplog.Infoln("[TNGF] Send AMF Configuration Update Failure")

	isAmfConfUpdateFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.AMF_CONFIGURATION_UPDATE_FAILURE,
		&isAmfConfUpdateFailureSent, ngCause, &additionalCause)

	pkt, err := BuildAMFConfigurationUpdateFailure(ngCause, time, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build AMF Configuration Update Failure failed : %+v\n", err)
		return
	}

	isAmfConfUpdateFailureSent, additionalCause = SendToAmf(amf, pkt)
}

func SendRANConfigurationUpdate(amf *context.TNGFAMF) {
	ngaplog.Infoln("[TNGF] Send RAN Configuration Update")

	isRanConfUpdateSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.RAN_CONFIGURATION_UPDATE_UPDATE,
		&isRanConfUpdateSent, emptyCause, &additionalCause)

	if available, _ := context.TNGFSelf().AMFReInitAvailableListLoad(amf.SCTPAddr); !available {
		additionalCause = ngap_metrics.AMF_TIME_REINIT_ERR
		ngaplog.Warnf(
			"[TNGF] Please Wait at least for the indicated time before reinitiating toward same AMF[%s]", amf.SCTPAddr)
		return
	}

	pkt, err := BuildRANConfigurationUpdate()
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngaplog.Errorf("Build AMF Configuration Update Failure failed : %+v\n", err)
		return
	}

	isRanConfUpdateSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUplinkRANConfigurationTransfer() {
	ngaplog.Infoln("[TNGF] Send Uplink RAN Configuration Transfer")
}

func SendUplinkRANStatusTransfer() {
	ngaplog.Infoln("[TNGF] Send Uplink RAN Status Transfer")
}

func SendLocationReportingFailureIndication() {
	ngaplog.Infoln("[TNGF] Send Location Reporting Failure Indication")
}

func SendLocationReport() {
	ngaplog.Infoln("[TNGF] Send Location Report")
}

func SendRRCInactiveTransitionReport() {
	ngaplog.Infoln("[TNGF] Send RRC Inactive Transition Report")
}
