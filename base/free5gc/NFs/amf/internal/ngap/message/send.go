package message

import (
	"time"

	"github.com/free5gc/amf/internal/context"
	"github.com/free5gc/amf/internal/logger"
	business_metrics "github.com/free5gc/amf/internal/metrics/business"
	callback "github.com/free5gc/amf/internal/sbi/processor/notifier"
	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	ngap_metrics "github.com/free5gc/util/metrics/ngap"
	"github.com/free5gc/util/metrics/utils"
)

var emptyCause = ngapType.Cause{Present: 0}

func SendToRan(ran *context.AmfRan, packet []byte) (bool, string) {
	defer func() {
		// This is workaround.
		// TODO: Handle ran.Conn close event correctly
		err := recover()
		if err != nil {
			logger.NgapLog.Warnf("Send error, gNB may have been lost: %+v", err)
		}
	}()

	if ran == nil {
		logger.NgapLog.Error("Ran is nil")
		return false, ngap_metrics.RAN_NIL_ERR
	}

	if len(packet) == 0 {
		ran.Log.Error("packet len is 0")
		return false, "packet len is 0"
	}

	if ran.Conn == nil {
		ran.Log.Error("Ran conn is nil")
		return false, "Ran conn is nil"
	}

	if ran.Conn.RemoteAddr() == nil {
		ran.Log.Error("Ran addr is nil")
		return false, "Ran addr is nil"
	}

	ran.Log.Debugf("Send NGAP message To Ran")

	if n, err := ran.Conn.Write(packet); err != nil {
		ran.Log.Errorf("Send error: %+v", err)
		return false, ngap_metrics.SCTP_SOCKET_WRITE_ERR
	} else {
		ran.Log.Debugf("Write %d bytes", n)
	}
	return true, ""
}

func SendToRanUe(ue *context.RanUe, packet []byte) (bool, string) {
	var ran *context.AmfRan

	if ue == nil {
		logger.NgapLog.Error("RanUe is nil")
		return false, ngap_metrics.RAN_UE_NIL_ERR
	}

	if ran = ue.Ran; ran == nil {
		logger.NgapLog.Error("Ran is nil")
		return false, ngap_metrics.RAN_NIL_ERR
	}

	if ue.AmfUe == nil {
		ue.Log.Warn("AmfUe is nil")
	}

	return SendToRan(ran, packet)
}

func NasSendToRan(ue *context.AmfUe, accessType models.AccessType, packet []byte) (bool, string) {
	if ue == nil {
		logger.NgapLog.Error("AmfUe is nil")
		return false, ngap_metrics.AMF_UE_NIL_ERR
	}

	ranUe := ue.RanUe[accessType]
	if ranUe == nil {
		logger.NgapLog.Error("RanUe is nil")
		return false, ngap_metrics.RAN_UE_NIL_ERR
	}

	return SendToRanUe(ranUe, packet)
}

func SendNGSetupResponse(ran *context.AmfRan) {
	isNGSetupRespSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_SETUP_RESPONSE, &isNGSetupRespSent, emptyCause, &additionalCause)

	ran.Log.Info("Send NG-Setup response")

	pkt, err := BuildNGSetupResponse()
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build NGSetupResponse failed : %s", err.Error())
		return
	}

	isNGSetupRespSent, additionalCause = SendToRan(ran, pkt)
}

func SendNGSetupFailure(ran *context.AmfRan, cause ngapType.Cause) {
	isNGSetupFailSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_SETUP_FAILURE, &isNGSetupFailSent, cause, &additionalCause)

	ran.Log.Info("Send NG-Setup failure")

	if cause.Present == ngapType.CausePresentNothing {
		additionalCause = ngap_metrics.CAUSE_NIL_ERR
		ran.Log.Errorf("Cause present is nil")
		return
	}

	pkt, err := BuildNGSetupFailure(cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build NGSetupFailure failed : %s", err.Error())
		return
	}
	isNGSetupFailSent, additionalCause = SendToRan(ran, pkt)
}

// partOfNGInterface: if reset type is "reset all", set it to nil TS 38.413 9.2.6.11
func SendNGReset(ran *context.AmfRan, cause ngapType.Cause,
	partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
) {
	isNGResetSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.NG_RESET, &isNGResetSent, cause, &additionalCause)

	ran.Log.Info("Send NG Reset")

	pkt, err := BuildNGReset(cause, partOfNGInterface)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build NGReset failed : %s", err.Error())
		return
	}
	isNGResetSent, additionalCause = SendToRan(ran, pkt)
}

func SendNGResetAcknowledge(ran *context.AmfRan, partOfNGInterface *ngapType.UEAssociatedLogicalNGConnectionList,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isNGResetAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.NG_RESET_ACKNOWLEDGE, &isNGResetAckSent, emptyCause, &additionalCause)

	ran.Log.Info("Send NG Reset Acknowledge")

	if partOfNGInterface != nil && len(partOfNGInterface.List) == 0 {
		additionalCause = "length of partOfNGInterface is 0"
		ran.Log.Error("length of partOfNGInterface is 0")
		return
	}

	pkt, err := BuildNGResetAcknowledge(partOfNGInterface, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build NGResetAcknowledge failed : %s", err.Error())
		return
	}
	isNGResetAckSent, additionalCause = SendToRan(ran, pkt)
}

func SendDownlinkNasTransport(ue *context.RanUe, nasPdu []byte,
	mobilityRestrictionList *ngapType.MobilityRestrictionList,
) {
	isDLNASTransportSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.DOWNLINK_NAS_TRANSPORT, &isDLNASTransportSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Downlink Nas Transport")

	if len(nasPdu) == 0 {
		ue.Log.Errorf("Send DownlinkNasTransport Error: nasPdu is nil")
	}

	pkt, err := BuildDownlinkNasTransport(ue, nasPdu, mobilityRestrictionList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build DownlinkNasTransport failed : %s", err.Error())
		return
	}
	isDLNASTransportSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendPDUSessionResourceReleaseCommand(ue *context.RanUe, nasPdu []byte,
	pduSessionResourceReleasedList ngapType.PDUSessionResourceToReleaseListRelCmd,
) {
	isPDUSessResRelCmdSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PDUSESSION_RESOURCE_RELEASE_COMMAND, &isPDUSessResRelCmdSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send PDU Session Resource Release Command")

	pkt, err := BuildPDUSessionResourceReleaseCommand(ue, nasPdu, pduSessionResourceReleasedList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build PDUSessionResourceReleaseCommand failed : %s", err.Error())
		return
	}
	isPDUSessResRelCmdSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendUEContextReleaseCommand(ue *context.RanUe, action context.RelAction, causePresent int, cause aper.Enumerated) {
	isUECtxReleaseCmd := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.UE_CONTEXT_RELEASE_COMMAND, &isUECtxReleaseCmd, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send UE Context Release Command")

	pkt, err := BuildUEContextReleaseCommand(ue, causePresent, cause)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build UEContextReleaseCommand failed : %s", err.Error())
		return
	}
	ue.ReleaseAction = action
	if ue.AmfUe != nil && ue.Ran != nil {
		ue.AmfUe.ReleaseCause[ue.Ran.AnType] = &context.CauseAll{
			NgapCause: &models.NgApCause{
				Group: int32(causePresent),
				Value: int32(cause),
			},
		}
	}
	ue.InitialContextSetup = false
	isUECtxReleaseCmd, additionalCause = SendToRanUe(ue, pkt)
}

func SendErrorIndication(ran *context.AmfRan, amfUeNgapId *ngapType.AMFUENGAPID, ranUeNgapId *ngapType.RANUENGAPID,
	cause *ngapType.Cause, criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isErrorIndicationSent := false
	additionalCause := ""
	if cause == nil {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION, &isErrorIndicationSent,
			emptyCause, &additionalCause)
	} else {
		defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.ERROR_INDICATION, &isErrorIndicationSent, *cause,
			&additionalCause)
	}

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Error Indication")

	var amfUeNgapIdValue *int64
	if amfUeNgapId != nil {
		amfUeNgapIdValue = &amfUeNgapId.Value
	}
	var ranUeNgapIdValue *int64
	if ranUeNgapId != nil {
		ranUeNgapIdValue = &ranUeNgapId.Value
	}

	pkt, err := BuildErrorIndication(amfUeNgapIdValue, ranUeNgapIdValue, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build ErrorIndication failed : %s", err.Error())
		return
	}
	isErrorIndicationSent, additionalCause = SendToRan(ran, pkt)
}

func SendUERadioCapabilityCheckRequest(ue *context.RanUe) {
	isUERadioCapCheckReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.UE_RADIO_CAPABILITY_CHECK_REQUEST, &isUERadioCapCheckReqSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send UE Radio Capability Check Request")

	pkt, err := BuildUERadioCapabilityCheckRequest(ue)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build UERadioCapabilityCheckRequest failed : %s", err.Error())
		return
	}
	isUERadioCapCheckReqSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendHandoverCancelAcknowledge(ue *context.RanUe, criticalityDiagnostics *ngapType.CriticalityDiagnostics) {
	isHoCancelAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.HANDOVER_CANCEL_ACKNOWLEDGE, &isHoCancelAckSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Handover Cancel Acknowledge")

	pkt, err := BuildHandoverCancelAcknowledge(ue, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build HandoverCancelAcknowledge failed : %s", err.Error())
		return
	}
	isHoCancelAckSent, additionalCause = SendToRanUe(ue, pkt)
}

// nasPDU: from nas layer
// pduSessionResourceSetupRequestList: provided by AMF, and transfer data is from SMF
func SendPDUSessionResourceSetupRequest(ue *context.RanUe, nasPdu []byte,
	pduSessionResourceSetupRequestList *ngapType.PDUSessionResourceSetupListSUReq,
) {
	isPDUSessResSetupReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PDUSESSION_RESOURCE_SETUP_REQUEST, &isPDUSessResSetupReqSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send PDU Session Resource Setup Request")

	if len(pduSessionResourceSetupRequestList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ue.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildPDUSessionResourceSetupRequest(ue, nasPdu, pduSessionResourceSetupRequestList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build PDUSessionResourceSetupRequest failed : %s", err.Error())
		return
	}
	isPDUSessResSetupReqSent, additionalCause = SendToRanUe(ue, pkt)
}

// pduSessionResourceModifyConfirmList: provided by AMF, and transfer data is return from SMF
// pduSessionResourceFailedToModifyList: provided by AMF, and transfer data is return from SMF
func SendPDUSessionResourceModifyConfirm(
	ue *context.RanUe,
	pduSessionResourceModifyConfirmList ngapType.PDUSessionResourceModifyListModCfm,
	pduSessionResourceFailedToModifyList ngapType.PDUSessionResourceFailedToModifyListModCfm,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isPDUSessResModifyConfirmSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PDUSESSION_RESOURCE_MODIFY_CONFIRM, &isPDUSessResModifyConfirmSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send PDU Session Resource Modify Confirm")

	if len(pduSessionResourceModifyConfirmList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ue.Log.Error("Pdu List out of range")
		return
	}

	if len(pduSessionResourceFailedToModifyList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ue.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyConfirm(ue, pduSessionResourceModifyConfirmList,
		pduSessionResourceFailedToModifyList, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build PDUSessionResourceModifyConfirm failed : %s", err.Error())
		return
	}
	isPDUSessResModifyConfirmSent, additionalCause = SendToRanUe(ue, pkt)
}

// pduSessionResourceModifyRequestList: from SMF
func SendPDUSessionResourceModifyRequest(ue *context.RanUe,
	pduSessionResourceModifyRequestList ngapType.PDUSessionResourceModifyListModReq,
) {
	isPDUSessResModifyReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PDUSESSION_RESOURCE_MODIFY_REQUEST, &isPDUSessResModifyReqSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send PDU Session Resource Modify Request")

	if len(pduSessionResourceModifyRequestList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ue.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildPDUSessionResourceModifyRequest(ue, pduSessionResourceModifyRequestList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build PDUSessionResourceModifyRequest failed : %s", err.Error())
		return
	}
	isPDUSessResModifyReqSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendInitialContextSetupRequest(
	amfUe *context.AmfUe,
	anType models.AccessType,
	nasPdu []byte,
	pduSessionResourceSetupRequestList *ngapType.PDUSessionResourceSetupListCxtReq,
	rrcInactiveTransitionReportRequest *ngapType.RRCInactiveTransitionReportRequest,
	coreNetworkAssistanceInfo *ngapType.CoreNetworkAssistanceInformation,
	emergencyFallbackIndicator *ngapType.EmergencyFallbackIndicator,
) {
	isInitialCtxSetupReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.INITIAL_CONTEXT_SETUP_REQUEST, &isInitialCtxSetupReqSent, emptyCause, &additionalCause)

	if amfUe == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	amfUe.RanUe[anType].Log.Info("Send Initial Context Setup Request")

	if pduSessionResourceSetupRequestList != nil {
		if len(pduSessionResourceSetupRequestList.List) > context.MaxNumOfPDUSessions {
			additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
			amfUe.RanUe[anType].Log.Error("Pdu List out of range")
			return
		}
	}

	pkt, err := BuildInitialContextSetupRequest(amfUe, anType, nasPdu, pduSessionResourceSetupRequestList,
		rrcInactiveTransitionReportRequest, coreNetworkAssistanceInfo, emergencyFallbackIndicator)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		amfUe.RanUe[anType].Log.Errorf("Build InitialContextSetupRequest failed : %s", err.Error())
		return
	}

	isInitialCtxSetupReqSent, additionalCause = NasSendToRan(amfUe, anType, pkt)
}

func SendUEContextModificationRequest(
	amfUe *context.AmfUe,
	anType models.AccessType,
	oldAmfUeNgapID *int64,
	rrcInactiveTransitionReportRequest *ngapType.RRCInactiveTransitionReportRequest,
	coreNetworkAssistanceInfo *ngapType.CoreNetworkAssistanceInformation,
	mobilityRestrictionList *ngapType.MobilityRestrictionList,
	emergencyFallbackIndicator *ngapType.EmergencyFallbackIndicator,
) {
	isUeCtxModifReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.UE_CONTEXT_MODIFICATION_REQUEST, &isUeCtxModifReqSent, emptyCause, &additionalCause)

	if amfUe == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	amfUe.RanUe[anType].Log.Info("Send UE Context Modification Request")

	pkt, err := BuildUEContextModificationRequest(amfUe, anType, oldAmfUeNgapID, rrcInactiveTransitionReportRequest,
		coreNetworkAssistanceInfo, mobilityRestrictionList, emergencyFallbackIndicator)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		amfUe.RanUe[anType].Log.Errorf("Build UEContextModificationRequest failed : %s", err.Error())
		return
	}
	isUeCtxModifReqSent, additionalCause = NasSendToRan(amfUe, anType, pkt)
}

// pduSessionResourceHandoverList: provided by amf and transfer is return from smf
// pduSessionResourceToReleaseList: provided by amf and transfer is return from smf
// criticalityDiagnostics = criticalityDiagonstics IE in receiver node's error indication
// when received node can't comprehend the IE or missing IE
func SendHandoverCommand(
	sourceUe *context.RanUe,
	pduSessionResourceHandoverList ngapType.PDUSessionResourceHandoverList,
	pduSessionResourceToReleaseList ngapType.PDUSessionResourceToReleaseListHOCmd,
	container ngapType.TargetToSourceTransparentContainer,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isHoCmdSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.HANDOVER_COMMAND, &isHoCmdSent, emptyCause, &additionalCause)

	if sourceUe == nil {
		additionalCause = ngap_metrics.SOURCE_UE_NIL_ERR
		logger.NgapLog.Error("SourceUe is nil")
		return
	}

	sourceUe.Log.Info("Send Handover Command")

	if len(pduSessionResourceHandoverList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		sourceUe.Log.Error("Pdu List out of range")
		return
	}

	if len(pduSessionResourceToReleaseList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		sourceUe.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildHandoverCommand(sourceUe, pduSessionResourceHandoverList, pduSessionResourceToReleaseList,
		container, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		sourceUe.Log.Errorf("Build HandoverCommand failed : %s", err.Error())
		return
	}
	isHoCmdSent, additionalCause = SendToRanUe(sourceUe, pkt)
}

// cause = initiate the Handover Cancel procedure with the appropriate value for the Cause IE.
// criticalityDiagnostics = criticalityDiagonstics IE in receiver node's error indication
// when received node can't comprehend the IE or missing IE
func SendHandoverPreparationFailure(sourceUe *context.RanUe, cause ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isHoPrepFailSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.HANDOVER_PREPARATION_FAILURE, &isHoPrepFailSent, cause, &additionalCause)

	if sourceUe == nil {
		additionalCause = ngap_metrics.SOURCE_UE_NIL_ERR
		logger.NgapLog.Error("SourceUe is nil")
		return
	}

	sourceUe.Log.Info("Send Handover Preparation Failure")

	amfUe := sourceUe.AmfUe
	if amfUe == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		sourceUe.Log.Error("amfUe is nil")
		return
	}
	amfUe.SetOnGoing(sourceUe.Ran.AnType, &context.OnGoing{
		Procedure: context.OnGoingProcedureNothing,
	})
	pkt, err := BuildHandoverPreparationFailure(sourceUe, cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		sourceUe.Log.Errorf("Build HandoverPreparationFailure failed : %s", err.Error())
		return
	}
	isHoPrepFailSent, additionalCause = SendToRanUe(sourceUe, pkt)
}

/*The PGW-C+SMF (V-SMF in the case of home-routed roaming scenario only) sends
a Nsmf_PDUSession_CreateSMContext Response(N2 SM Information (PDU Session ID, cause code)) to the AMF.*/
// Cause is from SMF
// pduSessionResourceSetupList provided by AMF, and the transfer data is from SMF
// sourceToTargetTransparentContainer is received from S-RAN
// nsci: new security context indicator, if amfUe has updated security context, set nsci to true, otherwise set to false
// N2 handover in same AMF
func SendHandoverRequest(sourceUe *context.RanUe, targetRan *context.AmfRan, cause ngapType.Cause,
	pduSessionResourceSetupListHOReq ngapType.PDUSessionResourceSetupListHOReq,
	sourceToTargetTransparentContainer ngapType.SourceToTargetTransparentContainer, nsci bool,
) {
	isHoReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.HANDOVER_REQUEST, &isHoReqSent, cause, &additionalCause)

	defer func(msgSent *bool, cause ngapType.Cause) {
		hoCause := ngap_metrics.GetCauseErrorStr(&cause)
		if hoCause == "unknown ngapType.Cause" {
			hoCause = additionalCause
		}

		if msgSent != nil && !*msgSent {
			business_metrics.IncrHoEventCounter(business_metrics.HANDOVER_TYPE_NGAP_VALUE,
				utils.FailureMetric, hoCause, sourceUe.HandOverStartTime)
		}
	}(&isHoReqSent, cause)

	if sourceUe == nil {
		additionalCause = ngap_metrics.SOURCE_UE_NIL_ERR
		logger.NgapLog.Error("sourceUe is nil")
		return
	}

	sourceUe.Log.Info("Send Handover Request")

	amfUe := sourceUe.AmfUe
	if amfUe == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		sourceUe.Log.Error("amfUe is nil")
		return
	}
	if targetRan == nil {
		additionalCause = ngap_metrics.TARGET_RAN_NIL_ERR
		sourceUe.Log.Error("targetRan is nil")
		return
	}

	if sourceUe.TargetUe != nil {
		additionalCause = ngap_metrics.HANDOVER_REQUIRED_DUP_ERR
		sourceUe.Log.Error("Handover Required Duplicated")
		return
	}

	if len(pduSessionResourceSetupListHOReq.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		sourceUe.Log.Error("Pdu List out of range")
		return
	}

	if len(sourceToTargetTransparentContainer.Value) == 0 {
		additionalCause = ngap_metrics.SRC_TO_TARGET_TRANSPARENT_CONTAINER_NIL_ERR
		sourceUe.Log.Error("Source To Target TransparentContainer is nil")
		return
	}

	var targetUe *context.RanUe
	if targetUeTmp, err := targetRan.NewRanUe(context.RanUeNgapIdUnspecified); err != nil {
		sourceUe.Log.Errorf("Create target UE error: %+v", err)
	} else {
		targetUe = targetUeTmp
	}

	sourceUe.Log.Tracef("Source : AMF_UE_NGAP_ID[%d], RAN_UE_NGAP_ID[%d]", sourceUe.AmfUeNgapId, sourceUe.RanUeNgapId)

	// Possible nil pointer here
	if targetUe != nil {
		sourceUe.Log.Tracef("Target : AMF_UE_NGAP_ID[%d], RAN_UE_NGAP_ID[Unknown]", targetUe.AmfUeNgapId)
	}

	context.AttachSourceUeTargetUe(sourceUe, targetUe)

	pkt, err := BuildHandoverRequest(targetUe, cause, pduSessionResourceSetupListHOReq,
		sourceToTargetTransparentContainer, nsci)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		sourceUe.Log.Errorf("Build HandoverRequest failed : %s", err.Error())
		return
	}
	isHoReqSent, additionalCause = SendToRanUe(targetUe, pkt)
}

// pduSessionResourceSwitchedList: provided by AMF, and the transfer data is from SMF
// pduSessionResourceReleasedList: provided by AMF, and the transfer data is from SMF
// newSecurityContextIndicator: if AMF has activated a new 5G NAS security context, set it to true,
// otherwise set to false
// coreNetworkAssistanceInformation: provided by AMF, based on collection of UE behavior statistics
// and/or other available
// information about the expected UE behavior. TS 23.501 5.4.6, 5.4.6.2
// rrcInactiveTransitionReportRequest: configured by amf
// criticalityDiagnostics: from received node when received not comprehended IE or missing IE
func SendPathSwitchRequestAcknowledge(
	ue *context.RanUe,
	pduSessionResourceSwitchedList ngapType.PDUSessionResourceSwitchedList,
	pduSessionResourceReleasedList ngapType.PDUSessionResourceReleasedListPSAck,
	newSecurityContextIndicator bool,
	coreNetworkAssistanceInformation *ngapType.CoreNetworkAssistanceInformation,
	rrcInactiveTransitionReportRequest *ngapType.RRCInactiveTransitionReportRequest,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
	hoStartTime time.Time,
) {
	isPathSwitchReqAckSent := false
	business_metrics.IncrHoEventCounter(business_metrics.HANDOVER_TYPE_XN_VALUE,
		utils.SuccessMetric, business_metrics.HANDOVER_EMPTY_CAUSE, hoStartTime)
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PATH_SWITCH_REQUEST_ACKNOWLEDGE, &isPathSwitchReqAckSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Path Switch Request Acknowledge")

	if len(pduSessionResourceSwitchedList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_SWITCH_OOO_ERR
		ue.Log.Error("Pdu Session Resource Switched List out of range")
		return
	}

	if len(pduSessionResourceReleasedList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_SESS_RESOURCE_SWITCH_OOO_ERR
		ue.Log.Error("Pdu Session Resource Released List out of range")
		return
	}

	pkt, err := BuildPathSwitchRequestAcknowledge(ue, pduSessionResourceSwitchedList, pduSessionResourceReleasedList,
		newSecurityContextIndicator, coreNetworkAssistanceInformation, rrcInactiveTransitionReportRequest,
		criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build PathSwitchRequestAcknowledge failed : %s", err.Error())
		return
	}
	isPathSwitchReqAckSent, additionalCause = SendToRanUe(ue, pkt)
}

// pduSessionResourceReleasedList: provided by AMF, and the transfer data is from SMF
// criticalityDiagnostics: from received node when received not comprehended IE or missing IE
func SendPathSwitchRequestFailure(
	ran *context.AmfRan,
	amfUeNgapId,
	ranUeNgapId int64,
	pduSessionResourceReleasedList *ngapType.PDUSessionResourceReleasedListPSFail,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
	commonError string,
	hoStartTime time.Time,
) {
	business_metrics.IncrHoEventCounter(business_metrics.HANDOVER_TYPE_XN_VALUE, utils.FailureMetric, commonError,
		hoStartTime)

	isPathSwitchReqFailSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.PATH_SWITCH_REQUEST_FAILURE, &isPathSwitchReqFailSent, emptyCause, &additionalCause)

	ran.Log.Info("Send Path Switch Request Failure")

	if pduSessionResourceReleasedList != nil && len(pduSessionResourceReleasedList.List) > context.MaxNumOfPDUSessions {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ran.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildPathSwitchRequestFailure(amfUeNgapId, ranUeNgapId, pduSessionResourceReleasedList,
		criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build PathSwitchRequestFailure failed : %s", err.Error())
		return
	}
	isPathSwitchReqFailSent, additionalCause = SendToRan(ran, pkt)
}

// RanStatusTransferTransparentContainer from Uplink Ran Configuration Transfer
func SendDownlinkRanStatusTransfer(ue *context.RanUe, container ngapType.RANStatusTransferTransparentContainer) {
	isDLRANStatusTransfertSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.DOWNLINK_RAN_STATUS_TRANSFER, &isDLRANStatusTransfertSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Downlink Ran Status Transfer")

	if len(container.DRBsSubjectToStatusTransferList.List) > context.MaxNumOfDRBs {
		additionalCause = ngap_metrics.PDU_LIST_OOR_ERR
		ue.Log.Error("Pdu List out of range")
		return
	}

	pkt, err := BuildDownlinkRanStatusTransfer(ue, container)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build DownlinkRanStatusTransfer failed : %s", err.Error())
		return
	}

	isDLRANStatusTransfertSent, additionalCause = SendToRanUe(ue, pkt)
}

// anType indicate amfUe send this msg for which accessType
// Paging Priority: is included only if the AMF receives an Namf_Communication_N1N2MessageTransfer message
// with an ARP value associated with
// priority services (e.g., MPS, MCS), as configured by the operator. (TS 23.502 4.2.3.3, TS 23.501 5.22.3)
// pagingOriginNon3GPP: TS 23.502 4.2.3.3 step 4b: If the UE is simultaneously registered over 3GPP and non-3GPP
// accesses in the same PLMN,
// the UE is in CM-IDLE state in both 3GPP access and non-3GPP access, and the PDU Session ID in step 3a
// is associated with non-3GPP access, the AMF sends a Paging message with associated access "non-3GPP" to
// NG-RAN node(s) via 3GPP access.
// more paging policy with 3gpp/non-3gpp access is described in TS 23.501 5.6.8
func SendPaging(ue *context.AmfUe, ngapBuf []byte) {
	isPagingSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.PAGING, &isPagingSent, emptyCause, &additionalCause)

	// var pagingPriority *ngapType.PagingPriority
	if ue == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	// if ppi != nil {
	// pagingPriority = new(ngapType.PagingPriority)
	// pagingPriority.Value = aper.Enumerated(*ppi)
	// }
	// pkt, err := BuildPaging(ue, pagingPriority, pagingOriginNon3GPP)
	// if err != nil {
	// 	ngaplog.Errorf("Build Paging failed : %s", err.Error())
	// }
	taiList := ue.RegistrationArea[models.AccessType__3_GPP_ACCESS]
	context.GetSelf().AmfRanPool.Range(func(key, value interface{}) bool {
		ran := value.(*context.AmfRan)
		for _, item := range ran.SupportedTAList {
			if context.InTaiList(item.Tai, taiList) {
				ue.GmmLog.Infof("Send Paging to TAI(%+v, Tac:%+v)", item.Tai.PlmnId, item.Tai.Tac)
				isPagingSent, additionalCause = SendToRan(ran, ngapBuf)
				break
			}
		}
		return true
	})

	if context.GetSelf().T3513Cfg.Enable {
		cfg := context.GetSelf().T3513Cfg
		ue.GmmLog.Infof("Start T3513 timer")
		ue.T3513 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			ue.GmmLog.Warnf("T3513 expires, retransmit Paging (retry: %d)", expireTimes)
			context.GetSelf().AmfRanPool.Range(func(key, value interface{}) bool {
				ran := value.(*context.AmfRan)
				for _, item := range ran.SupportedTAList {
					if context.InTaiList(item.Tai, taiList) {
						isPagingSent, additionalCause = SendToRan(ran, ngapBuf)
						break
					}
				}
				return true
			})
		}, func() {
			ue.GmmLog.Warnf("T3513 expires %d times, abort paging procedure", cfg.MaxRetryTimes)
			ue.T3513 = nil // clear the timer
			if ue.OnGoing(models.AccessType__3_GPP_ACCESS).Procedure != context.OnGoingProcedureN2Handover {
				callback.SendN1N2TransferFailureNotification(ue, models.N1N2MessageTransferCause_UE_NOT_RESPONDING)
			}
		})
	}
}

// TS 23.502 4.2.2.2.3
// anType: indicate amfUe send this msg for which accessType
// amfUeNgapID: initial AMF get it from target AMF
// ngapMessage: initial UE Message to reroute
// allowedNSSAI: provided by AMF, and AMF get it from NSSF (4.2.2.2.3 step 4b)
func SendRerouteNasRequest(ue *context.AmfUe, anType models.AccessType, amfUeNgapID *int64, ngapMessage []byte,
	allowedNSSAI *ngapType.AllowedNSSAI,
) {
	isRerouteNasReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.REROUTE_NAS_REQUEST, &isRerouteNasReqSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	ue.RanUe[anType].Log.Info("Send Reroute Nas Request")

	if len(ngapMessage) == 0 {
		additionalCause = ngap_metrics.NGAP_MSG_NIL_ERR
		ue.RanUe[anType].Log.Error("Ngap Message is nil")
		return
	}

	pkt, err := BuildRerouteNasRequest(ue, anType, amfUeNgapID, ngapMessage, allowedNSSAI)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.RanUe[anType].Log.Errorf("Build RerouteNasRequest failed : %s", err.Error())
		return
	}
	isRerouteNasReqSent, additionalCause = NasSendToRan(ue, anType, pkt)
}

// criticality ->from received node when received node can't comprehend the IE or missing IE
func SendRanConfigurationUpdateAcknowledge(
	ran *context.AmfRan, criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isRanConfigurationUpdateAckSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.RAN_CONFIGURATION_UPDATE_ACKNOWLEDGE, &isRanConfigurationUpdateAckSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Ran Configuration Update Acknowledge")

	pkt, err := BuildRanConfigurationUpdateAcknowledge(criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build RanConfigurationUpdateAcknowledge failed : %s", err.Error())
		return
	}
	isRanConfigurationUpdateAckSent, additionalCause = SendToRan(ran, pkt)
}

// criticality ->from received node when received node can't comprehend the IE or missing IE
// If the AMF cannot accept the update,
// it shall respond with a RAN CONFIGURATION UPDATE FAILURE message and appropriate cause value.
func SendRanConfigurationUpdateFailure(ran *context.AmfRan, cause ngapType.Cause,
	criticalityDiagnostics *ngapType.CriticalityDiagnostics,
) {
	isRanConfigurationUpdateFailSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.RAN_CONFIGURATION_UPDATE_FAILURE, &isRanConfigurationUpdateFailSent, cause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Ran Configuration Update Failure")

	pkt, err := BuildRanConfigurationUpdateFailure(cause, criticalityDiagnostics)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build RanConfigurationUpdateFailure failed : %s", err.Error())
		return
	}

	isRanConfigurationUpdateFailSent, additionalCause = SendToRan(ran, pkt)
}

// An AMF shall be able to instruct other peer CP NFs, subscribed to receive such a notification,
// that it will be unavailable on this AMF and its corresponding target AMF(s).
// If CP NF does not subscribe to receive AMF unavailable notification, the CP NF may attempt
// forwarding the transaction towards the old AMF and detect that the AMF is unavailable. When
// it detects unavailable, it marks the AMF and its associated GUAMI(s) as unavailable.
// Defined in 23.501 5.21.2.2.2
func SendAMFStatusIndication(ran *context.AmfRan, unavailableGUAMIList ngapType.UnavailableGUAMIList) {
	isAMFStatusIndicationSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.AMF_STATUS_INDICATION, &isAMFStatusIndicationSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send AMF Status Indication")

	if len(unavailableGUAMIList.List) > context.MaxNumOfServedGuamiList {
		additionalCause = ngap_metrics.GUAMI_LIST_OOR_ERR
		ran.Log.Error("GUAMI List out of range")
		return
	}

	pkt, err := BuildAMFStatusIndication(unavailableGUAMIList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build AMFStatusIndication failed : %s", err.Error())
		return
	}

	isAMFStatusIndicationSent, additionalCause = SendToRan(ran, pkt)
}

// TS 23.501 5.19.5.2
// amfOverloadResponse: the required behavior of NG-RAN, provided by AMF
// amfTrafficLoadReductionIndication(int 1~99): indicates the percentage of the type, set to 0 if does not need this ie
// of traffic relative to the instantaneous incoming rate at the NG-RAN node, provided by AMF
// overloadStartNSSAIList: overload slices, provide by AMF
func SendOverloadStart(
	ran *context.AmfRan,
	amfOverloadResponse *ngapType.OverloadResponse,
	amfTrafficLoadReductionIndication int64,
	overloadStartNSSAIList *ngapType.OverloadStartNSSAIList,
) {
	isOverloadStartSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.OVERLOAD_START, &isOverloadStartSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Overload Start")

	if amfTrafficLoadReductionIndication != 0 &&
		(amfTrafficLoadReductionIndication < 1 || amfTrafficLoadReductionIndication > 99) {
		additionalCause = ngap_metrics.AMF_TRAFFIC_LOAD_REDUCTION_INDICATION_OOO_ERR
		ran.Log.Error("AmfTrafficLoadReductionIndication out of range (should be 1 ~ 99)")
		return
	}

	if overloadStartNSSAIList != nil && len(overloadStartNSSAIList.List) > context.MaxNumOfSlice {
		additionalCause = ngap_metrics.NSSAI_LIST_OOR_ERR
		ran.Log.Error("NSSAI List out of range")
		return
	}

	pkt, err := BuildOverloadStart(amfOverloadResponse, amfTrafficLoadReductionIndication, overloadStartNSSAIList)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build OverloadStart failed : %s", err.Error())
		return
	}

	isOverloadStartSent, additionalCause = SendToRan(ran, pkt)
}

func SendOverloadStop(ran *context.AmfRan) {
	isOverleadStopSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.OVERLOAD_STOP, &isOverleadStopSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Overload Stop")

	pkt, err := BuildOverloadStop()
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build OverloadStop failed : %s", err.Error())
		return
	}

	isOverleadStopSent, additionalCause = SendToRan(ran, pkt)
}

// SONConfigurationTransfer = sONConfigurationTransfer from uplink Ran Configuration Transfer
func SendDownlinkRanConfigurationTransfer(ran *context.AmfRan, transfer *ngapType.SONConfigurationTransfer) {
	isDLRandConfigurationTransferSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.DOWNLINK_RAN_CONFIGURATION_TRANSFER, &isDLRandConfigurationTransferSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send Downlink Ran Configuration Transfer")

	pkt, err := BuildDownlinkRanConfigurationTransfer(transfer)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build DownlinkRanConfigurationTransfer failed : %s", err.Error())
		return
	}

	isDLRandConfigurationTransferSent, additionalCause = SendToRan(ran, pkt)
}

// NRPPa PDU is by pass
// NRPPa PDU is from LMF define in 4.13.5.6
func SendDownlinkNonUEAssociatedNRPPATransport(ue *context.RanUe, nRPPaPDU ngapType.NRPPaPDU) {
	metricsStatus := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.DOWNLINK_NON_UE_ASSOCIATED_NRPPA_TRANSPORT, &metricsStatus, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Downlink Non UE Associated NRPPA Transport")

	if len(nRPPaPDU.Value) == 0 {
		additionalCause = ngap_metrics.NRPPA_LEN_ZERO_ERR
		ue.Log.Error("length of NRPPA-PDU is 0")
		return
	}

	pkt, err := BuildDownlinkNonUEAssociatedNRPPATransport(ue, nRPPaPDU)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build DownlinkNonUEAssociatedNRPPATransport failed : %s", err.Error())
		return
	}

	metricsStatus, additionalCause = SendToRanUe(ue, pkt)
}

func SendDeactivateTrace(amfUe *context.AmfUe, anType models.AccessType) {
	isDeactivateTraceSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.DEACTIVATE_TRACE, &isDeactivateTraceSent, emptyCause, &additionalCause)

	if amfUe == nil {
		additionalCause = ngap_metrics.AMF_UE_NIL_ERR
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	ranUe := amfUe.RanUe[anType]
	if ranUe == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ranUe.Log.Info("Send Deactivate Trace")

	pkt, err := BuildDeactivateTrace(amfUe, anType)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ranUe.Log.Errorf("Build DeactivateTrace failed : %s", err.Error())
		return
	}

	isDeactivateTraceSent, additionalCause = SendToRanUe(ranUe, pkt)
}

// AOI List is from SMF
// The SMF may subscribe to the UE mobility event notification from the AMF
// (e.g. location reporting, UE moving into or out of Area Of Interest) TS 23.502 4.3.2.2.1 Step.17
// The Location Reporting Control message shall identify the UE for which reports are requested and may include
// Reporting Type, Location Reporting Level, Area Of Interest and Request Reference ID
// TS 23.502 4.10 LocationReportingProcedure
// The AMF may request the NG-RAN location reporting with event reporting type (e.g. UE location or UE presence
// in Area of Interest), reporting mode and its related parameters (e.g. number of reporting) TS 23.501 5.4.7
// Location Reference ID To Be Canceled IE shall be present if the Event Type IE is set to "Stop UE presence
// in the area of interest". otherwise set it to 0
func SendLocationReportingControl(
	ue *context.RanUe,
	aoiList *ngapType.AreaOfInterestList,
	locationReportingReferenceIDToBeCancelled int64,
	eventType ngapType.EventType,
) {
	isLocationReportingControlSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.LOCATION_REPORTING_CONTROL, &isLocationReportingControlSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Location Reporting Control")

	if aoiList != nil && len(aoiList.List) > context.MaxNumOfAOI {
		additionalCause = ngap_metrics.AOI_LIST_OOR_ERR
		ue.Log.Error("AOI List out of range")
		return
	}

	if eventType.Value == ngapType.EventTypePresentStopUePresenceInAreaOfInterest {
		if locationReportingReferenceIDToBeCancelled < 1 || locationReportingReferenceIDToBeCancelled > 64 {
			additionalCause = ngap_metrics.LOCATION_REPORTING_REFERENCE_ID_OOR_ERR
			ue.Log.Error("LocationReportingReferenceIDToBeCancelled out of range (should be 1 ~ 64)")
			return
		}
	}

	pkt, err := BuildLocationReportingControl(ue, aoiList, locationReportingReferenceIDToBeCancelled, eventType)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build LocationReportingControl failed : %s", err.Error())
		return
	}
	isLocationReportingControlSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendUETNLABindingReleaseRequest(ue *context.RanUe) {
	isUETNLABindingRelReqSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.UE_TNLA_BINDING_RELEASE_REQUEST, &isUETNLABindingRelReqSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send UE TNLA Binging Release Request")

	pkt, err := BuildUETNLABindingReleaseRequest(ue)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build UETNLABindingReleaseRequest failed : %s", err.Error())
		return
	}

	isUETNLABindingRelReqSent, additionalCause = SendToRanUe(ue, pkt)
}

// Weight Factor associated with each of the TNL association within the AMF
func SendAMFConfigurationUpdate(ran *context.AmfRan, usage ngapType.TNLAssociationUsage,
	weightfactor ngapType.TNLAddressWeightFactor,
) {
	isAMFConfigurationUpdateSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(
		ngap_metrics.AMF_CONFIGURATION_UPDATE, &isAMFConfigurationUpdateSent, emptyCause, &additionalCause)

	if ran == nil {
		additionalCause = ngap_metrics.RAN_NIL_ERR
		logger.NgapLog.Error("Ran is nil")
		return
	}

	ran.Log.Info("Send AMF Configuration Update")

	pkt, err := BuildAMFConfigurationUpdate(usage, weightfactor)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ran.Log.Errorf("Build AMFConfigurationUpdate failed : %s", err.Error())
		return
	}

	isAMFConfigurationUpdateSent, additionalCause = SendToRan(ran, pkt)
}

// NRPPa PDU is a pdu from LMF to RAN defined in TS 23.502 4.13.5.5 step 3
// NRPPa PDU is by pass
func SendDownlinkUEAssociatedNRPPaTransport(ue *context.RanUe, nRPPaPDU ngapType.NRPPaPDU) {
	isDLUEAssociatedNRPPaTransportSent := false
	additionalCause := ""
	defer ngap_metrics.IncrMetricsSentMsg(ngap_metrics.DOWNLINK_UE_ASSOCIATED_NRPPA_TRANSPORT,
		&isDLUEAssociatedNRPPaTransportSent, emptyCause, &additionalCause)

	if ue == nil {
		additionalCause = ngap_metrics.RAN_UE_NIL_ERR
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	ue.Log.Info("Send Downlink UE Associated NRPPa Transport")

	if len(nRPPaPDU.Value) == 0 {
		additionalCause = ngap_metrics.NRPPA_LEN_ZERO_ERR
		ue.Log.Error("length of NRPPA-PDU is 0")
		return
	}

	pkt, err := BuildDownlinkUEAssociatedNRPPaTransport(ue, nRPPaPDU)
	if err != nil {
		additionalCause = ngap_metrics.NGAP_MSG_BUILD_ERR
		ue.Log.Errorf("Build DownlinkUEAssociatedNRPPaTransport failed : %s", err.Error())
		return
	}
	isDLUEAssociatedNRPPaTransportSent, additionalCause = SendToRanUe(ue, pkt)
}

func SendN2Message(
	amfUe *context.AmfUe,
	anType models.AccessType,
	nasPdu []byte,
	pduSessionResourceSetupRequestList *ngapType.PDUSessionResourceSetupListCxtReq,
	rrcInactiveTransitionReportRequest *ngapType.RRCInactiveTransitionReportRequest,
	coreNetworkAssistanceInfo *ngapType.CoreNetworkAssistanceInformation,
	emergencyFallbackIndicator *ngapType.EmergencyFallbackIndicator,
	mobilityRestrictionList *ngapType.MobilityRestrictionList,
) {
	if amfUe == nil {
		logger.NgapLog.Error("AmfUe is nil")
		return
	}

	ranUe := amfUe.RanUe[anType]
	if ranUe == nil {
		logger.NgapLog.Error("RanUe is nil")
		return
	}

	if !ranUe.InitialContextSetup && (ranUe.UeContextRequest ||
		(pduSessionResourceSetupRequestList != nil && len(pduSessionResourceSetupRequestList.List) > 0)) {
		SendInitialContextSetupRequest(amfUe, anType, nasPdu, pduSessionResourceSetupRequestList,
			rrcInactiveTransitionReportRequest, coreNetworkAssistanceInfo, emergencyFallbackIndicator)
	} else if ranUe.InitialContextSetup &&
		(pduSessionResourceSetupRequestList != nil && len(pduSessionResourceSetupRequestList.List) > 0) {
		suList := ConvertPDUSessionResourceSetupListCxtReqToSUReq(pduSessionResourceSetupRequestList)
		SendPDUSessionResourceSetupRequest(ranUe, nasPdu, suList)
	} else {
		SendDownlinkNasTransport(ranUe, nasPdu, mobilityRestrictionList)
	}
}
