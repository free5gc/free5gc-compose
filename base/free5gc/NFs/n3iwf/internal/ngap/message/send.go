package message

import (
	"runtime/debug"

	n3iwf_context "github.com/free5gc/n3iwf/internal/context"
	"github.com/free5gc/n3iwf/internal/logger"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/sctp"
	ngap_metrics "github.com/free5gc/util/metrics/ngap"
)

var emptyCause = ngapType.Cause{Present: 0}

func SendToAmf(amf *n3iwf_context.N3IWFAMF, pkt []byte) (bool, string) {
	ngapLog := logger.NgapLog
	if amf == nil {
		ngapLog.Errorf("AMF Context is nil ")
		return false, "AMF Context is nil"
	} else {
		if n, err := amf.SCTPConn.Write(pkt); err != nil {
			ngapLog.Errorf("Write to SCTP socket failed: %+v", err)
			return false, ngap_metrics.SCTP_SOCKET_WRITE_ERR
		} else {
			ngapLog.Tracef("Wrote %d bytes", n)
		}
	}
	return true, ""
}

func SendNGSetupRequest(
	conn *sctp.SCTPConn,
	n3iwfCtx *n3iwf_context.N3IWFContext,
) {
	ngapLog := logger.NgapLog
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			ngapLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	isNGSetupReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_SETUP_REQUEST, &isNGSetupReqSent, emptyCause, &additionalCause)

	ngapLog.Infoln("Send NG Setup Request")

	cfg := n3iwfCtx.Config()
	sctpAddr := conn.RemoteAddr().String()

	if available, _ := n3iwfCtx.AMFReInitAvailableListLoad(sctpAddr); !available {
		additionalCause = ngap_metrics.AMF_TIME_REINIT_ERR
		ngapLog.Warnf(
			"Please Wait at least for the indicated time before reinitiating toward same AMF[%s]",
			sctpAddr)
		return
	}
	pkt, err := BuildNGSetupRequest(
		cfg.GetGlobalN3iwfId(),
		cfg.GetRanNodeName(),
		cfg.GetSupportedTAList(),
	)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build NGSetup Request failed: %+v\n", err)
		return
	}

	if n, err := conn.Write(pkt); err != nil {
		ngapLog.Errorf("Write to SCTP socket failed: %+v", err)
		additionalCause = ngap_metrics.SCTP_SOCKET_WRITE_ERR
	} else {
		ngapLog.Tracef("Wrote %d bytes", n)
		isNGSetupReqSent = true
	}
}

// partOfNGInterface: if reset type is "reset all", set it to nil TS 38.413 9.2.6.11
func SendNGReset(
	amf *n3iwf_context.N3IWFAMF,
	cause ngapType.Cause,
	partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
) {
	isNGResetSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_RESET, &isNGResetSent, cause, &additionalCause)
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send NG Reset")

	pkt, err := BuildNGReset(cause, partOfNGInterface)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build NGReset failed : %s", err.Error())
		return
	}

	isNGResetSent, additionalCause = SendToAmf(amf, pkt)
}

func SendNGResetAcknowledge(
	amf *n3iwf_context.N3IWFAMF,
	partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send NG Reset Acknowledge")

	isNGResetAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_RESET_ACKNOWLEDGE,
		&isNGResetAckSent, emptyCause, &additionalCause)

	if partOfNGInterface != nil && len(partOfNGInterface.List) == 0 {
		ngapLog.Error("length of partOfNGInterface is 0")
		additionalCause = ngap_metrics.NF_INTERFACE_LEN_ZERO_ERR
		return
	}

	pkt, err := BuildNGResetAcknowledge(partOfNGInterface, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build NGReset Acknowledge failed : %s", err.Error())
		return
	}

	isNGResetAckSent, additionalCause = SendToAmf(amf, pkt)
}

func SendInitialContextSetupResponse(
	ranUe n3iwf_context.RanUe,
	responseList *ngapType.PDUSessionResourceSetupListCxtRes,
	failedList *ngapType.PDUSessionResourceFailedToSetupListCxtRes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Initial Context Setup Response")

	isInitialCtxRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_CONTEXT_SETUP_RESPONSE,
		&isInitialCtxRespSent, emptyCause, &additionalCause)

	if responseList != nil && len(responseList.List) > n3iwf_context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngapLog.Errorln("Pdu List out of range")
		return
	}

	if failedList != nil && len(failedList.List) > n3iwf_context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngapLog.Errorln("Pdu List out of range")
		return
	}

	pkt, err := BuildInitialContextSetupResponse(ranUe, responseList, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Initial Context Setup Response failed : %+v\n", err)
		return
	}

	isInitialCtxRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendInitialContextSetupFailure(
	ranUe n3iwf_context.RanUe,
	cause ngapType.Cause,
	failedList *ngapType.PDUSessionResourceFailedToSetupListCxtFail,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Initial Context Setup Failure")

	isInitialCtxFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_CONTEXT_SETUP_RESPONSE,
		&isInitialCtxFailureSent, cause, &additionalCause)

	if failedList != nil && len(failedList.List) > n3iwf_context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ngapLog.Errorln("Pdu List out of range")
		return
	}

	pkt, err := BuildInitialContextSetupFailure(ranUe, cause, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Initial Context Setup Failure failed : %+v\n", err)
		return
	}

	isInitialCtxFailureSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendUEContextModificationResponse(
	ranUe n3iwf_context.RanUe,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Context Modification Response")

	isUECtxModificationRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_MODIFICATION_RESPONSE,
		&isUECtxModificationRespSent, emptyCause, &additionalCause)

	pkt, err := BuildUEContextModificationResponse(ranUe, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build UE Context Modification Response failed : %+v\n", err)
		return
	}

	isUECtxModificationRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendUEContextModificationFailure(
	ranUe n3iwf_context.RanUe,
	cause ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Context Modification Failure")

	isUECtxModificationFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_MODIFICATION_FAILURE,
		&isUECtxModificationFailureSent, cause, &additionalCause)

	pkt, err := BuildUEContextModificationFailure(ranUe, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build UE Context Modification Failure failed : %+v\n", err)
		return
	}

	isUECtxModificationFailureSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendUEContextReleaseComplete(
	ranUe n3iwf_context.RanUe,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Context Release Complete")

	isUECtxReleaseCompleteSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_RELEASE_COMPLETE,
		&isUECtxReleaseCompleteSent, emptyCause, &additionalCause)

	pkt, err := BuildUEContextReleaseComplete(ranUe, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build UE Context Release Complete failed : %+v\n", err)
		return
	}

	isUECtxReleaseCompleteSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendUEContextReleaseRequest(
	ranUe n3iwf_context.RanUe, cause ngapType.Cause,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Context Release Request")

	isUECtxReleaseReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_CONTEXT_RELEASE_REQUEST,
		&isUECtxReleaseReqSent, cause, &additionalCause)

	pkt, err := BuildUEContextReleaseRequest(ranUe, cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build UE Context Release Request failed : %+v\n", err)
		return
	}

	isUECtxReleaseReqSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendInitialUEMessage(amf *n3iwf_context.N3IWFAMF,
	ranUe n3iwf_context.RanUe, nasPdu []byte,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Initial UE Message")

	isInitialUEMessageSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.INITIAL_UE_MESSAGE,
		&isInitialUEMessageSent, emptyCause, &additionalCause)

	// Attach To AMF

	pkt, err := BuildInitialUEMessage(ranUe, nasPdu, nil)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Initial UE Message failed : %+v\n", err)
		return
	}

	isInitialUEMessageSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
	// ranUe.AttachAMF() // TODO: Check AttachAMF if is necessary
}

func SendUplinkNASTransport(
	ranUe n3iwf_context.RanUe,
	nasPdu []byte,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Uplink NAS Transport")

	isUplinkNasTransportSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UPLINK_NAS_TRANSPORT,
		&isUplinkNasTransportSent, emptyCause, &additionalCause)

	if len(nasPdu) == 0 {
		additionalCause = ngap_metrics.NAS_PDU_NIL_ERR
		ngapLog.Errorln("NAS Pdu is nil")
		return
	}

	pkt, err := BuildUplinkNASTransport(ranUe, nasPdu)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Uplink NAS Transport failed : %+v\n", err)
		return
	}

	isUplinkNasTransportSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendNASNonDeliveryIndication(
	ranUe n3iwf_context.RanUe,
	nasPdu []byte,
	cause ngapType.Cause,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send NAS NonDelivery Indication")

	isNasNonDeliveryIndicationSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NAS_NON_DELIVERY_INDICATION,
		&isNasNonDeliveryIndicationSent, cause, &additionalCause)

	if len(nasPdu) == 0 {
		additionalCause = ngap_metrics.NAS_PDU_NIL_ERR
		ngapLog.Errorln("NAS Pdu is nil")
		return
	}

	pkt, err := BuildNASNonDeliveryIndication(ranUe, nasPdu, cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build NAS Non Delivery Indication failed : %+v\n", err)
		return
	}

	isNasNonDeliveryIndicationSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendRerouteNASRequest() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Reroute NAS Request")
}

func SendPDUSessionResourceSetupResponse(
	ranUe n3iwf_context.RanUe,
	responseList *ngapType.PDUSessionResourceSetupListSURes,
	failedListSURes *ngapType.PDUSessionResourceFailedToSetupListSURes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send PDU Session Resource Setup Response")

	isPduSessionResourceSetupRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_SETUP_RESPONSE,
		&isPduSessionResourceSetupRespSent, emptyCause, &additionalCause)

	if ranUe == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngapLog.Error("UE context is nil, this information is mandatory.")
		return
	}

	pkt, err := BuildPDUSessionResourceSetupResponse(ranUe, responseList, failedListSURes, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build PDU Session Resource Setup Response failed : %+v", err)
		return
	}

	isPduSessionResourceSetupRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendPDUSessionResourceModifyResponse(
	ranUe n3iwf_context.RanUe,
	responseList *ngapType.PDUSessionResourceModifyListModRes,
	failedList *ngapType.PDUSessionResourceFailedToModifyListModRes,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send PDU Session Resource Modify Response")

	isPduSessionResourceModifyRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_MODIFY_RESPONSE,
		&isPduSessionResourceModifyRespSent, emptyCause, &additionalCause)

	if ranUe == nil && criticalityDiagnostics == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngapLog.Error("UE context is nil, this information is mandatory")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyResponse(ranUe, responseList, failedList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build PDU Session Resource Modify Response failed : %+v", err)
		return
	}

	isPduSessionResourceModifyRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendPDUSessionResourceModifyIndication(
	ranUe n3iwf_context.RanUe,
	modifyList []ngapType.PDUSessionResourceModifyItemModInd,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send PDU Session Resource Modify Indication")

	isPduSessionResourceModifyIndicationSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_MODIFY_INDICATION,
		&isPduSessionResourceModifyIndicationSent, emptyCause, &additionalCause)

	if ranUe == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngapLog.Error("UE context is nil, this information is mandatory")
		return
	}
	if modifyList == nil {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_MODIFY_LIST_NIL_ERR
		ngapLog.Errorln(
			"PDU Session Resource Modify Indication List is nil. This message shall contain at least one Item")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyIndication(ranUe, modifyList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build PDU Session Resource Modify Indication failed : %+v", err)
		return
	}

	isPduSessionResourceModifyIndicationSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendPDUSessionResourceNotify(
	ranUe n3iwf_context.RanUe,
	notiList *ngapType.PDUSessionResourceNotifyList,
	relList *ngapType.PDUSessionResourceReleasedListNot,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send PDU Session Resource Notify")

	isPduSessionResourceNotifySent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_NOTIFY,
		&isPduSessionResourceNotifySent, emptyCause, &additionalCause)

	if ranUe == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngapLog.Error("UE context is nil, this information is mandatory")
		return
	}

	pkt, err := BuildPDUSessionResourceNotify(ranUe, notiList, relList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build PDUSession Resource Notify failed : %+v", err)
		return
	}

	isPduSessionResourceNotifySent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendPDUSessionResourceReleaseResponse(
	ranUe n3iwf_context.RanUe,
	relList ngapType.PDUSessionResourceReleasedListRelRes,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send PDU Session Resource Release Response")

	isPduSessionResourceReleaseRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PDUSESSION_RESOURCE_RELEASE_RESPONSE,
		&isPduSessionResourceReleaseRespSent, emptyCause, &additionalCause)

	if ranUe == nil {
		additionalCause = ngap_metrics.UE_CTX_NIL
		ngapLog.Error("UE context is nil, this information is mandatory")
		return
	}
	if len(relList.List) < 1 {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_RELEASED_LIST_NIL_ERR
		ngapLog.Errorln(
			"PDUSessionResourceReleasedListRelRes is nil. This message shall contain at least one Item")
		return
	}

	pkt, err := BuildPDUSessionResourceReleaseResponse(ranUe, relList, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build PDU Session Resource Release Response failed : %+v", err)
		return
	}

	isPduSessionResourceReleaseRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendErrorIndication(
	amf *n3iwf_context.N3IWFAMF,
	amfUENGAPID *int64,
	ranUENGAPID *int64,
	cause *ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Error Indication")

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
		ngapLog.Errorln("Both cause and criticality is nil. This message shall contain at least one of them.")
		return
	}

	pkt, err := BuildErrorIndication(amfUENGAPID, ranUENGAPID, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Error Indication failed : %+v\n", err)
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
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Error Indication")

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
		ngapLog.Errorln("Both cause and criticality is nil. This message shall contain at least one of them.")
		return
	}

	pkt, err := BuildErrorIndication(amfUENGAPID, ranUENGAPID, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build Error Indication failed : %+v\n", err)
		return
	}

	if n, err := sctpConn.Write(pkt); err != nil {
		additionalCause = ngap_metrics.SCTP_SOCKET_WRITE_ERR
		ngapLog.Errorf("Write to SCTP socket failed: %+v", err)
	} else {
		isErrorIndicationWithSctpConnSent = true
		ngapLog.Tracef("Wrote %d bytes", n)
	}
}

func SendUERadioCapabilityInfoIndication() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Radio Capability Info Indication")
}

func SendUERadioCapabilityCheckResponse(
	amf *n3iwf_context.N3IWFAMF,
	ranUe n3iwf_context.RanUe,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send UE Radio Capability Check Response")

	isUeRadioCapabilityCheckRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.UE_RADIO_CAPABILITY_CHECK_RESPONSE,
		&isUeRadioCapabilityCheckRespSent, emptyCause, &additionalCause)

	pkt, err := BuildUERadioCapabilityCheckResponse(ranUe, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build UERadio Capability Check Response failed : %+v\n", err)
		return
	}
	isUeRadioCapabilityCheckRespSent, additionalCause = SendToAmf(ranUe.GetSharedCtx().AMF, pkt)
}

func SendAMFConfigurationUpdateAcknowledge(
	amf *n3iwf_context.N3IWFAMF,
	setupList *ngapType.AMFTNLAssociationSetupList,
	failList *ngapType.TNLAssociationList,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send AMF Configuration Update Acknowledge")

	isAmfConfUpdateAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.AMF_CONFIGURATION_UPDATE_ACKNOWLEDGE,
		&isAmfConfUpdateAckSent, emptyCause, &additionalCause)

	pkt, err := BuildAMFConfigurationUpdateAcknowledge(setupList, failList, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build AMF Configuration Update Acknowledge failed : %+v\n", err)
		return
	}

	isAmfConfUpdateAckSent, additionalCause = SendToAmf(amf, pkt)
}

func SendAMFConfigurationUpdateFailure(
	amf *n3iwf_context.N3IWFAMF,
	ngCause ngapType.Cause,
	time *ngapType.TimeToWait,
	diagnostics *ngapType.CriticalityDiagnostics,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send AMF Configuration Update Failure")

	isAmfConfUpdateFailureSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.AMF_CONFIGURATION_UPDATE_FAILURE,
		&isAmfConfUpdateFailureSent, ngCause, &additionalCause)

	pkt, err := BuildAMFConfigurationUpdateFailure(ngCause, time, diagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build AMF Configuration Update Failure failed : %+v\n", err)
		return
	}

	isAmfConfUpdateFailureSent, additionalCause = SendToAmf(amf, pkt)
}

func SendRANConfigurationUpdate(
	n3iwfCtx *n3iwf_context.N3IWFContext,
	amf *n3iwf_context.N3IWFAMF,
) {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send RAN Configuration Update")

	isRanConfUpdateSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.RAN_CONFIGURATION_UPDATE_UPDATE,
		&isRanConfUpdateSent, emptyCause, &additionalCause)

	available, _ := n3iwfCtx.AMFReInitAvailableListLoad(amf.SCTPAddr)
	if !available {
		additionalCause = ngap_metrics.AMF_TIME_REINIT_ERR
		ngapLog.Warnf(
			"Please Wait at least for the indicated time before reinitiating toward same AMF[%s]",
			amf.SCTPAddr)
		return
	}

	cfg := n3iwfCtx.Config()
	pkt, err := BuildRANConfigurationUpdate(
		cfg.GetRanNodeName(),
		cfg.GetSupportedTAList())
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ngapLog.Errorf("Build AMF Configuration Update Failure failed : %+v\n", err)
		return
	}

	isRanConfUpdateSent, additionalCause = SendToAmf(amf, pkt)
}

func SendUplinkRANConfigurationTransfer() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Uplink RAN Configuration Transfer")
}

func SendUplinkRANStatusTransfer() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Uplink RAN Status Transfer")
}

func SendLocationReportingFailureIndication() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Location Reporting Failure Indication")
}

func SendLocationReport() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send Location Report")
}

func SendRRCInactiveTransitionReport() {
	ngapLog := logger.NgapLog
	ngapLog.Infoln("Send RRC Inactive Transition Report")
}
