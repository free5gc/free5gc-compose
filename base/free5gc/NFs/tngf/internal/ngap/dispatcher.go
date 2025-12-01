package ngap

import (
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/ngap"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/sctp"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/internal/ngap/handler"
	"github.com/free5gc/tngf/pkg/context"
)

var Ngaplog *logrus.Entry

func init() {
	Ngaplog = logger.NgapLog
}

func Dispatch(conn *sctp.SCTPConn, msg []byte) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NgapLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	// AMF SCTP address
	sctpAddr := conn.RemoteAddr().String()
	// AMF context
	amf, _ := context.TNGFSelf().AMFPoolLoad(sctpAddr)
	// Decode
	pdu, err := ngap.Decoder(msg)
	if err != nil {
		Ngaplog.Errorf("NGAP decode error: %+v\n", err)
		return
	}

	switch pdu.Present {
	case ngapType.NGAPPDUPresentInitiatingMessage:
		initiatingMessage := pdu.InitiatingMessage
		if initiatingMessage == nil {
			Ngaplog.Errorln("Initiating Message is nil")
			return
		}

		switch initiatingMessage.ProcedureCode.Value {
		case ngapType.ProcedureCodeNGReset:
			handler.HandleNGReset(amf, pdu)
		case ngapType.ProcedureCodeInitialContextSetup:
			handler.HandleInitialContextSetupRequest(amf, pdu)
		case ngapType.ProcedureCodeUEContextModification:
			handler.HandleUEContextModificationRequest(amf, pdu)
		case ngapType.ProcedureCodeUEContextRelease:
			handler.HandleUEContextReleaseCommand(amf, pdu)
		case ngapType.ProcedureCodeDownlinkNASTransport:
			handler.HandleDownlinkNASTransport(amf, pdu)
		case ngapType.ProcedureCodePDUSessionResourceSetup:
			handler.HandlePDUSessionResourceSetupRequest(amf, pdu)
		case ngapType.ProcedureCodePDUSessionResourceModify:
			handler.HandlePDUSessionResourceModifyRequest(amf, pdu)
		case ngapType.ProcedureCodePDUSessionResourceRelease:
			handler.HandlePDUSessionResourceReleaseCommand(amf, pdu)
		case ngapType.ProcedureCodeErrorIndication:
			handler.HandleErrorIndication(amf, pdu)
		case ngapType.ProcedureCodeUERadioCapabilityCheck:
			handler.HandleUERadioCapabilityCheckRequest(amf, pdu)
		case ngapType.ProcedureCodeAMFConfigurationUpdate:
			handler.HandleAMFConfigurationUpdate(amf, pdu)
		case ngapType.ProcedureCodeDownlinkRANConfigurationTransfer:
			handler.HandleDownlinkRANConfigurationTransfer(pdu)
		case ngapType.ProcedureCodeDownlinkRANStatusTransfer:
			handler.HandleDownlinkRANStatusTransfer(pdu)
		case ngapType.ProcedureCodeAMFStatusIndication:
			handler.HandleAMFStatusIndication(pdu)
		case ngapType.ProcedureCodeLocationReportingControl:
			handler.HandleLocationReportingControl(pdu)
		case ngapType.ProcedureCodeUETNLABindingRelease:
			handler.HandleUETNLAReleaseRequest(pdu)
		case ngapType.ProcedureCodeOverloadStart:
			handler.HandleOverloadStart(amf, pdu)
		case ngapType.ProcedureCodeOverloadStop:
			handler.HandleOverloadStop(amf, pdu)
		default:
			Ngaplog.Warnf("Not implemented NGAP message(initiatingMessage), procedureCode:%d]\n",
				initiatingMessage.ProcedureCode.Value)
		}
	case ngapType.NGAPPDUPresentSuccessfulOutcome:
		successfulOutcome := pdu.SuccessfulOutcome
		if successfulOutcome == nil {
			Ngaplog.Errorln("Successful Outcome is nil")
			return
		}

		switch successfulOutcome.ProcedureCode.Value {
		case ngapType.ProcedureCodeNGSetup:
			handler.HandleNGSetupResponse(sctpAddr, conn, pdu)
		case ngapType.ProcedureCodeNGReset:
			handler.HandleNGResetAcknowledge(amf, pdu)
		case ngapType.ProcedureCodePDUSessionResourceModifyIndication:
			handler.HandlePDUSessionResourceModifyConfirm(amf, pdu)
		case ngapType.ProcedureCodeRANConfigurationUpdate:
			handler.HandleRANConfigurationUpdateAcknowledge(amf, pdu)
		default:
			Ngaplog.Warnf("Not implemented NGAP message(successfulOutcome), procedureCode:%d]\n",
				successfulOutcome.ProcedureCode.Value)
		}
	case ngapType.NGAPPDUPresentUnsuccessfulOutcome:
		unsuccessfulOutcome := pdu.UnsuccessfulOutcome
		if unsuccessfulOutcome == nil {
			Ngaplog.Errorln("Unsuccessful Outcome is nil")
			return
		}

		switch unsuccessfulOutcome.ProcedureCode.Value {
		case ngapType.ProcedureCodeNGSetup:
			handler.HandleNGSetupFailure(sctpAddr, conn, pdu)
		case ngapType.ProcedureCodeRANConfigurationUpdate:
			handler.HandleRANConfigurationUpdateFailure(amf, pdu)
		default:
			Ngaplog.Warnf("Not implemented NGAP message(unsuccessfulOutcome), procedureCode:%d]\n",
				unsuccessfulOutcome.ProcedureCode.Value)
		}
	}
}
