package message

import (
	"fmt"

	"github.com/free5gc/amf/internal/context"
	gmm_common "github.com/free5gc/amf/internal/gmm/common"
	"github.com/free5gc/amf/internal/logger"
	ngap_message "github.com/free5gc/amf/internal/ngap/message"
	callback "github.com/free5gc/amf/internal/sbi/processor/notifier"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	nasMetrics "github.com/free5gc/util/metrics/nas"
)

// backOffTimerUint = 7 means backoffTimer is null
func SendDLNASTransport(ue *context.RanUe, payloadContainerType uint8, nasPdu []byte,
	pduSessionId int32, cause uint8, backOffTimerUint *uint8, backOffTimer uint8,
) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.DL_NAS_TRANSPORT, &isNasMsgSent, cause, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendDLNASTransport: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendDLNASTransport: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	amfUe.GmmLog.Info("Send DL NAS Transport")

	var causePtr *uint8
	if cause != 0 {
		causePtr = &cause
	}
	nasMsg, err := BuildDLNASTransport(amfUe, ue.Ran.AnType, payloadContainerType, nasPdu,
		uint8(pduSessionId), causePtr, backOffTimerUint, backOffTimer)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

func SendNotification(ue *context.RanUe, nasMsg []byte) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.NOTIFICATION, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendNotification: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendNotification: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	amfUe.GmmLog.Info("Send Notification")

	if context.GetSelf().T3565Cfg.Enable {
		cfg := context.GetSelf().T3565Cfg
		amfUe.GmmLog.Infof("Start T3565 timer")
		amfUe.T3565 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3565 expires, retransmit Notification (retry: %d)", expireTimes)
			timerAdditionalCause := "Timer expired, retransmit Notification"
			isNasMsgSent = true
			defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.NOTIFICATION_TIMER, &isNasMsgSent, 0, &timerAdditionalCause)
			ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
		}, func() {
			amfUe.GmmLog.Warnf("T3565 Expires %d times, abort notification procedure", cfg.MaxRetryTimes)
			amfUe.T3565 = nil // clear the timer
			if amfUe.OnGoing(models.AccessType__3_GPP_ACCESS).Procedure != context.OnGoingProcedureN2Handover {
				callback.SendN1N2TransferFailureNotification(amfUe, models.N1N2MessageTransferCause_UE_NOT_RESPONDING)
			}
		})
	}
}

func SendIdentityRequest(ue *context.RanUe, accessType models.AccessType, typeOfIdentity uint8) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.IDENTITY_REQUEST, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendIdentityRequest: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendIdentityRequest: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	amfUe.GmmLog.Info("Send Identity Request")

	nasMsg, err := BuildIdentityRequest(amfUe, accessType, typeOfIdentity)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)

	amfUe.RequestIdentityType = typeOfIdentity

	if context.GetSelf().T3570Cfg.Enable {
		cfg := context.GetSelf().T3570Cfg
		amfUe.T3570 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3570 expires, retransmit Identity Request (retry: %d)", expireTimes)
			timerAdditionalCause := "Timer expired, retransmit Identity Request"
			defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.IDENTITY_REQUEST, &isNasMsgSent, 0, &timerAdditionalCause)
			ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
		}, func() {
			amfUe.GmmLog.Warnf("T3570 Expires %d times, abort identification procedure & ongoing 5GMM procedure",
				cfg.MaxRetryTimes)
			gmm_common.RemoveAmfUe(amfUe, false)
		})
	}
}

func SendAuthenticationRequest(ue *context.RanUe) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.AUTHENTICATION_REQUEST, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationRequest: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationRequest: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationRequest: Ran is nil")
		return
	}
	ran := ue.Ran
	amfUe.GmmLog.Infof("Send Authentication Request")

	if amfUe.AuthenticationCtx == nil {
		additionalCause = nasMetrics.AUTH_CTX_UE_NIL_ERR
		amfUe.GmmLog.Error("Authentication Context of UE is nil")
		return
	}

	nasMsg, err := BuildAuthenticationRequest(amfUe, ran.AnType)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)

	if context.GetSelf().T3560Cfg.Enable {
		cfg := context.GetSelf().T3560Cfg
		amfUe.GmmLog.Infof("Start T3560 timer")
		amfUe.T3560 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3560 expires, retransmit Authentication Request (retry: %d)", expireTimes)
			timerAdditionalCause := "Timer expired, retry authentication request"
			defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.AUTHENTICATION_REQUEST, &isNasMsgSent, 0, &timerAdditionalCause)
			ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
		}, func() {
			amfUe.Lock.Lock()
			defer amfUe.Lock.Unlock()

			amfUe.GmmLog.Warnf("T3560 Expires %d times, abort authentication procedure & ongoing 5GMM procedure",
				cfg.MaxRetryTimes)
			amfUe.T3560 = nil
			gmm_common.RemoveAmfUe(amfUe, false)
		})
	}
}

func SendServiceAccept(amfUe *context.AmfUe, anType models.AccessType,
	cxtList ngapType.PDUSessionResourceSetupListCxtReq, pDUSessionStatus *[16]bool,
	reactivationResult *[16]bool, errPduSessionId, errCause []uint8,
) error {
	isNasMsgSent := false
	additionalCause := getErrCauseSingleStr(errCause)
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.SERVICE_ACCEPT, &isNasMsgSent, 0, &additionalCause)

	if amfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		return fmt.Errorf("SendServiceAccept: AmfUe is nil")
	}
	if amfUe.RanUe[anType] == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		return fmt.Errorf("SendServiceAccept: RanUe is nil")
	}
	amfUe.GmmLog.Info("Send Service Accept")

	nasMsg, err := BuildServiceAccept(amfUe, anType, pDUSessionStatus, reactivationResult,
		errPduSessionId, errCause)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return err
	}

	if amfUe.RanUe[anType].UeContextRequest ||
		(!amfUe.RanUe[anType].InitialContextSetup && len(cxtList.List) > 0) {
		// update Kgnb/Kn3iwf
		amfUe.UpdateSecurityContext(anType)
	}

	isNasMsgSent = true
	ngap_message.SendN2Message(amfUe, anType, nasMsg, &cxtList, nil, nil, nil, nil)
	return nil
}

func SendConfigurationUpdateCommand(amfUe *context.AmfUe,
	accessType models.AccessType,
	flags *context.ConfigurationUpdateCommandFlags,
) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.CONFIGURATION_UPDATE_COMMAND, &isNasMsgSent, 0, &additionalCause)

	if amfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendConfigurationUpdateCommand: AmfUe is nil")
		return
	}
	if amfUe.RanUe[accessType] == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendConfigurationUpdateCommand: RanUe is nil")
		return
	}

	nasMsg, err, startT3555 := BuildConfigurationUpdateCommand(amfUe, accessType, flags)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Errorf("BuildConfigurationUpdateCommand Error: %+v", err)
		return
	}
	amfUe.GmmLog.Info("Send Configuration Update Command")

	mobilityRestrictionList := ngap_message.BuildIEMobilityRestrictionList(amfUe)
	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(amfUe.RanUe[accessType], nasMsg, &mobilityRestrictionList)

	if startT3555 && context.GetSelf().T3555Cfg.Enable {
		cfg := context.GetSelf().T3555Cfg
		amfUe.GmmLog.Infof("Start T3555 timer")
		amfUe.T3555 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3555 expires, retransmit Configuration Update Command (retry: %d)",
				expireTimes)
			timerAdditionalCause := "Timer expired, retry configuration update command"
			defer nasMetrics.IncrMetricsSentNasMsgs(
				nasMetrics.CONFIGURATION_UPDATE_COMMAND_TIMER, &isNasMsgSent, 0, &timerAdditionalCause)
			ngap_message.SendDownlinkNasTransport(amfUe.RanUe[accessType], nasMsg, &mobilityRestrictionList)
		}, func() {
			amfUe.GmmLog.Warnf("T3555 Expires %d times, abort configuration update procedure",
				cfg.MaxRetryTimes)
		},
		)
	}
}

func SendAuthenticationReject(ue *context.RanUe, eapMsg string, cause5GMM uint8, otherCause string) {
	isNasMsgSent := false
	additionalCause := otherCause
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.AUTHENTICATION_REJECT, &isNasMsgSent, cause5GMM, &additionalCause)

	if ue == nil {
		// We keep the latest error even if there were a cause added in the argument
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationReject: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationReject: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationReject: Ran is nil")
		return
	}
	ran := ue.Ran
	amfUe.GmmLog.Info("Send Authentication Reject")

	nasMsg, err := BuildAuthenticationReject(amfUe, ran.AnType, eapMsg)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

func SendAuthenticationResult(ue *context.RanUe, eapSuccess bool, eapMsg string) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.AUTHENTICATION_RESULT, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationResult: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationResult: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendAuthenticationResult: Ran is nil")
		return
	}
	ran := ue.Ran
	amfUe.GmmLog.Info("Send Authentication Result")

	nasMsg, err := BuildAuthenticationResult(amfUe, ran.AnType, eapSuccess, eapMsg)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

func SendServiceReject(ue *context.RanUe, pDUSessionStatus *[16]bool, cause uint8) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.SERVICE_REJECT, &isNasMsgSent, cause, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendServiceReject: RanUe is nil")
		return
	}
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendServiceReject: Ran is nil")
		return
	}
	ran := ue.Ran
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		ue.Log.Info("Send Service Reject")
	} else {
		ue.AmfUe.GmmLog.Info("Send Service Reject")
	}

	nasMsg, err := BuildServiceReject(ue.AmfUe, ran.AnType, pDUSessionStatus, cause)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		if ue.AmfUe == nil {
			ue.Log.Error(err.Error())
		} else {
			ue.AmfUe.GmmLog.Error(err.Error())
		}
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

// T3502: This IE may be included to indicate a value for timer T3502 during the initial registration
// eapMessage: if the REGISTRATION REJECT message is used to convey EAP-failure message
func SendRegistrationReject(ue *context.RanUe, cause5GMM uint8, eapMessage string) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.REGISTRATION_REJECT, &isNasMsgSent, cause5GMM,
		&additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendRegistrationReject: RanUe is nil")
		return
	}
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendRegistrationReject: Ran is nil")
		return
	}
	ran := ue.Ran
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		ue.Log.Info("Send Registration Reject")
	} else {
		ue.AmfUe.GmmLog.Info("Send Registration Reject")
	}

	nasMsg, err := BuildRegistrationReject(ue.AmfUe, ran.AnType, cause5GMM, eapMessage)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		if ue.AmfUe == nil {
			ue.Log.Error(err.Error())
		} else {
			ue.AmfUe.GmmLog.Error(err.Error())
		}
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

// eapSuccess: only used when authType is EAP-AKA', set the value to false if authType is not EAP-AKA'
// eapMessage: only used when authType is EAP-AKA', set the value to "" if authType is not EAP-AKA'
func SendSecurityModeCommand(ue *context.RanUe, accessType models.AccessType, eapSuccess bool, eapMessage string) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.SECURITY_MODE_COMMAND, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendSecurityModeCommand: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendSecurityModeCommand: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	amfUe.GmmLog.Info("Send Security Mode Command")

	nasMsg, err := BuildSecurityModeCommand(amfUe, accessType, eapSuccess, eapMessage)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)

	if context.GetSelf().T3560Cfg.Enable {
		cfg := context.GetSelf().T3560Cfg
		amfUe.GmmLog.Infof("Start T3560 timer")
		amfUe.T3560 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3560 expires, retransmit Security Mode Command (retry: %d)", expireTimes)
			timerAdditionalCause := "Retry Security Mode Command"
			defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.SECURITY_MODE_COMMAND, &isNasMsgSent, 0, &timerAdditionalCause)
			ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
		}, func() {
			amfUe.Lock.Lock()
			defer amfUe.Lock.Unlock()

			amfUe.GmmLog.Warnf("T3560 Expires %d times, abort security mode control procedure", cfg.MaxRetryTimes)
			amfUe.T3560 = nil
			gmm_common.RemoveAmfUe(amfUe, false)
		})
	}
}

func SendDeregistrationRequest(ue *context.RanUe, accessType uint8, reRegistrationRequired bool, cause5GMM uint8) {
	if ue == nil {
		logger.GmmLog.Error("SendDeregistrationRequest: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		logger.GmmLog.Error("SendDeregistrationRequest: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	amfUe.GmmLog.Info("Send Deregistration Request")

	nasMsg, err := BuildDeregistrationRequest(ue, accessType, reRegistrationRequired, cause5GMM)
	if err != nil {
		amfUe.GmmLog.Error(err.Error())
		return
	}
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)

	if context.GetSelf().T3522Cfg.Enable {
		cfg := context.GetSelf().T3522Cfg
		amfUe.GmmLog.Infof("Start T3522 timer")
		amfUe.T3522 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			amfUe.GmmLog.Warnf("T3522 expires, retransmit Deregistration Request (retry: %d)", expireTimes)
			ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
		}, func() {
			amfUe.GmmLog.Warnf("T3522 Expires %d times, abort deregistration procedure", cfg.MaxRetryTimes)
			amfUe.T3522 = nil // clear the timer
			switch accessType {
			case nasMessage.AccessType3GPP:
				amfUe.GmmLog.Warnln("UE accessType[3GPP] transfer to Deregistered state")
				amfUe.State[models.AccessType__3_GPP_ACCESS].Set(context.Deregistered)
			case nasMessage.AccessTypeNon3GPP:
				amfUe.GmmLog.Warnln("UE accessType[Non3GPP] transfer to Deregistered state")
				amfUe.State[models.AccessType_NON_3_GPP_ACCESS].Set(context.Deregistered)
			default:
				amfUe.GmmLog.Warnln("UE accessType[3GPP] transfer to Deregistered state")
				amfUe.State[models.AccessType__3_GPP_ACCESS].Set(context.Deregistered)
				amfUe.GmmLog.Warnln("UE accessType[Non3GPP] transfer to Deregistered state")
				amfUe.State[models.AccessType_NON_3_GPP_ACCESS].Set(context.Deregistered)
			}
		})
	}
}

func SendDeregistrationAccept(ue *context.RanUe) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(
		nasMetrics.DEREGISTRATION_ACCEPT_UE_ORIGINATING_DEREGISTRATION, &isNasMsgSent, 0, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendDeregistrationAccept: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendDeregistrationAccept: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendDeregistrationAccept: Ran is nil")
		return
	}
	ran := ue.Ran
	amfUe.GmmLog.Info("Send Deregistration Accept")

	nasMsg, err := BuildDeregistrationAccept(ue.AmfUe, ran.AnType)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

func SendRegistrationAccept(
	amfUe *context.AmfUe, anType models.AccessType,
	pDUSessionStatus *[16]bool, reactivationResult *[16]bool,
	errPduSessionId, errCause []uint8,
	cxtList *ngapType.PDUSessionResourceSetupListCxtReq,
) {
	isNasMsgSent := false
	additionalCause := getErrCauseSingleStr(errCause)

	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.REGISTRATION_ACCEPT, &isNasMsgSent, 0, &additionalCause)

	if amfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendRegistrationAccept: AmfUe is nil")
		return
	}
	if amfUe.RanUe[anType] == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendRegistrationAccept: RanUe is nil")
		return
	}
	amfUe.GmmLog.Info("Send Registration Accept")

	nasMsg, err := BuildRegistrationAccept(amfUe, anType, pDUSessionStatus, reactivationResult, errPduSessionId, errCause)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}

	isNasMsgSent = true
	if anType == models.AccessType_NON_3_GPP_ACCESS {
		// TS 23.502 4.12.2.2 10a ~ 13: if non-3gpp, AMF should send initial context setup request to N3IWF first,
		// and send registration accept after receiving initial context setup response
		amfUe.RegistrationAcceptForNon3GPPAccess = nasMsg
		ngap_message.SendInitialContextSetupRequest(amfUe, anType, nil, cxtList, nil, nil, nil)
	} else {
		// anType is 3GPP_ACCESS
		ngap_message.SendN2Message(amfUe, anType, nasMsg, cxtList, nil, nil, nil, nil)
	}

	if context.GetSelf().T3550Cfg.Enable {
		cfg := context.GetSelf().T3550Cfg
		amfUe.GmmLog.Infof("Start T3550 timer")
		amfUe.T3550 = context.NewTimer(cfg.ExpireTime, cfg.MaxRetryTimes, func(expireTimes int32) {
			if amfUe.RanUe[anType] == nil {
				amfUe.GmmLog.Warnf("[NAS] UE Context released, abort retransmission of Registration Accept")
				amfUe.T3550 = nil
			} else {
				amfUe.GmmLog.Warnf("T3550 expires, retransmit Registration Accept (retry: %d)", expireTimes)
				timerAdditionalCause := "Retry Registration Accept"
				defer nasMetrics.IncrMetricsSentNasMsgs(
					nasMetrics.REGISTRATION_ACCEPT_TIMER, &isNasMsgSent, 0, &timerAdditionalCause)
				ngap_message.SendN2Message(amfUe, anType, nasMsg, cxtList, nil, nil, nil, nil)
			}
		}, func() {
			amfUe.GmmLog.Warnf("T3550 Expires %d times, abort retransmission of Registration Accept", cfg.MaxRetryTimes)
			amfUe.T3550 = nil // clear the timer
			// TS 24.501 5.5.1.2.8 case c, 5.5.1.3.8 case c
			amfUe.State[anType].Set(context.Registered)
			amfUe.ClearRegistrationRequestData(anType)
		})
	}
}

func SendStatus5GMM(ue *context.RanUe, cause uint8) {
	isNasMsgSent := false
	additionalCause := ""
	defer nasMetrics.IncrMetricsSentNasMsgs(nasMetrics.STATUS_5GMM, &isNasMsgSent, cause, &additionalCause)

	if ue == nil {
		additionalCause = nasMetrics.RAN_UE_NIL_ERR
		logger.GmmLog.Error("SendStatus5GMM: RanUe is nil")
		return
	}
	if ue.AmfUe == nil {
		additionalCause = nasMetrics.AMF_UE_NIL_ERR
		logger.GmmLog.Error("SendStatus5GMM: AmfUe is nil")
		return
	}
	amfUe := ue.AmfUe
	if ue.Ran == nil {
		additionalCause = nasMetrics.RAN_NIL_ERR
		logger.GmmLog.Error("SendStatus5GMM: Ran is nil")
		return
	}
	ran := ue.Ran
	amfUe.GmmLog.Info("Send Status 5GMM")

	nasMsg, err := BuildStatus5GMM(ue.AmfUe, ran.AnType, cause)
	if err != nil {
		additionalCause = nasMetrics.NAS_MSG_BUILD_ERR
		amfUe.GmmLog.Error(err.Error())
		return
	}
	ngap_message.SendDownlinkNasTransport(ue, nasMsg, nil)
}

func getErrCauseSingleStr(errCause []uint8) string {
	result := ""
	// transform errCause into a single string
	if len(errCause) > 1 {
		result += "Multiple Causes : "
		for i, c := range errCause {
			result += nasMessage.Cause5GMMToString(c)
			if i < len(errCause)-1 {
				result += "; "
			}
		}
	} else if len(errCause) == 1 {
		result = nasMessage.Cause5GMMToString(errCause[0])
	}
	return result
}
