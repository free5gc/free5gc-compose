package handler

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/tngf/internal/logger"
	ngap_message "github.com/free5gc/tngf/internal/ngap/message"
	"github.com/free5gc/tngf/pkg/context"
	radius_message "github.com/free5gc/tngf/pkg/radius/message"
	"github.com/free5gc/util/ueauth"
)

// Log
var radiusLog *logrus.Entry

func init() {
	radiusLog = logger.RadiusLog
}

// Radius state
const (
	EAP5GStart = iota
	EAP5GNAS
	InitialContextSetup
)

func HandleRadiusAccessRequest(udpConn *net.UDPConn, tngfAddr, ueAddr *net.UDPAddr,
	message *radius_message.RadiusMessage,
) {
	radiusLog.Infoln("Handle Radius Access Request")
	responseRadiusMessage := new(radius_message.RadiusMessage)
	var responseRadiusPayload radius_message.RadiusPayloadContainer

	tngfSelf := context.TNGFSelf()

	// var userName string
	var callingStationId string
	var calledStationId string
	var eapMessage []byte
	var requestMessageAuthenticator []byte
	var tnapId uint64
	var err error
	var userName string
	for i, radiusPayload := range message.Payloads {
		switch radiusPayload.Type {
		case radius_message.TypeUserName:
			userName = string(radiusPayload.Val)
		case radius_message.TypeCallingStationId:
			callingStationId = string(radiusPayload.Val)
		case radius_message.TypeEAPMessage:
			eapMessage = radiusPayload.Val
		case radius_message.TypeCalledStationId:
			calledStationId = string(radiusPayload.Val)
			calledStationId = strings.ReplaceAll(calledStationId[:17], "-", "")
			tnapId, err = strconv.ParseUint(calledStationId, 16, 64)
			if err != nil {
				radiusLog.Errorln("Request Message CalledStationId error", err)
				return
			}
		case radius_message.TypeMessageAuthenticator:
			requestMessageAuthenticator = radiusPayload.Val
			radiusLog.Debugln("Message Authenticator:\n", hex.Dump(requestMessageAuthenticator))

			message.Payloads[i].Val = make([]byte, 16)
			exRequestMessageAuthenticator := GetMessageAuthenticator(message)
			radiusLog.Debugln("expected authenticator:\n", hex.Dump(exRequestMessageAuthenticator))
			if !bytes.Equal(requestMessageAuthenticator, exRequestMessageAuthenticator) {
				radiusLog.Errorln("Request Message Authenticator error")
				return
			}
		}
	}

	var session *context.RadiusSession
	session, ok := tngfSelf.RadiusSessionPoolLoad(callingStationId)
	if !ok {
		session = tngfSelf.NewRadiusSession(callingStationId)
	}

	switch session.State {
	case EAP5GStart:
		// EAP expanded 5G-Start
		radiusLog.Infoln("Handle EAP-Res/Identity")
		var identifier uint8

		identifier, err = GenerateRandomUint8()
		if err != nil {
			radiusLog.Errorf("Random number failed: %+v", err)
			return
		}

		responseRadiusPayload.BuildEAP5GStart(identifier)

		responseRadiusMessage.BuildRadiusHeader(radius_message.AccessChallenge, message.PktID, message.Auth)

		if requestMessageAuthenticator != nil {
			tmpRadiusMessage := *responseRadiusMessage
			payload := new(radius_message.RadiusPayload)
			payload.Type = radius_message.TypeMessageAuthenticator
			payload.Length = uint8(18)
			payload.Val = make([]byte, 16)

			tmpResponseRadiusPayload := responseRadiusPayload
			tmpResponseRadiusPayload = append(tmpResponseRadiusPayload, *payload)

			tmpRadiusMessage.Payloads = tmpResponseRadiusPayload

			payload.Val = GetMessageAuthenticator(&tmpRadiusMessage)
			responseRadiusPayload = append(responseRadiusPayload, *payload)
		}
		responseRadiusMessage.Payloads = responseRadiusPayload
		SendRadiusMessageToUE(udpConn, tngfAddr, ueAddr, responseRadiusMessage)
		session.State = EAP5GNAS

	case EAP5GNAS:
		radiusLog.Infoln("Handle EAP-Res/5G-NAS")
		eap := &radius_message.EAP{}
		err = eap.Unmarshal(eapMessage)
		if err != nil {
			radiusLog.Errorf("[EAP] EAP5GNAS unmarshal error: %+v", err)
			return
		}
		if eap.Code != radius_message.EAPCodeResponse {
			radiusLog.Error("[EAP] Received an EAP payload with code other than response. Drop the payload.")
			return
		}

		eapTypeData := eap.EAPTypeData[0]
		var eapExpanded *radius_message.EAPExpanded

		switch eapTypeData.Type() {
		case radius_message.EAPTypeExpanded:
			eapExpanded = eapTypeData.(*radius_message.EAPExpanded)
		default:
			radiusLog.Errorf("[EAP] Received EAP packet with type other than EAP expanded type: %d", eapTypeData.Type())
			return
		}

		if eapExpanded.VendorID != radius_message.VendorID3GPP {
			radiusLog.Error("The peer sent EAP expended packet with wrong vendor ID. Drop the packet.")
			return
		}
		if eapExpanded.VendorType != radius_message.VendorTypeEAP5G {
			radiusLog.Error("The peer sent EAP expanded packet with wrong vendor type. Drop the packet.")
			return
		}

		eap5GMessageID, anParameters, nasPDU, unmarshal_err := UnmarshalEAP5GData(eapExpanded.VendorData)
		if unmarshal_err != nil {
			radiusLog.Errorf("Unmarshalling EAP-5G packet failed: %+v", unmarshal_err)
			return
		}

		if eap5GMessageID == radius_message.EAP5GType5GStop {
			// Send EAP failure
			// Build Radius message
			responseRadiusMessage.BuildRadiusHeader(radius_message.AccessChallenge, message.PktID, message.Auth)
			responseRadiusMessage.Payloads.Reset()

			// EAP
			identifier, random_err := GenerateRandomUint8()
			if random_err != nil {
				radiusLog.Errorf("Generate random uint8 failed: %+v", random_err)
				return
			}
			responseRadiusPayload.BuildEAPfailure(identifier)

			if requestMessageAuthenticator != nil {
				tmpRadiusMessage := *responseRadiusMessage
				payload := new(radius_message.RadiusPayload)
				payload.Type = radius_message.TypeMessageAuthenticator
				payload.Length = uint8(18)
				payload.Val = make([]byte, 16)

				tmpResponseRadiusPayload := responseRadiusPayload
				tmpResponseRadiusPayload = append(tmpResponseRadiusPayload, *payload)

				tmpRadiusMessage.Payloads = tmpResponseRadiusPayload

				payload.Val = GetMessageAuthenticator(&tmpRadiusMessage)
				responseRadiusPayload = append(responseRadiusPayload, *payload)
			}
			responseRadiusMessage.Payloads = responseRadiusPayload

			// Send Radius message to UE
			SendRadiusMessageToUE(udpConn, tngfAddr, ueAddr, responseRadiusMessage)
			return
		}

		// Send Initial UE Message or Uplink NAS Transport
		if session.ThisUE == nil {
			// print AN parameters
			radiusLog.Debug("Select AMF with the following AN parameters:")
			if anParameters.GUAMI == nil {
				radiusLog.Debug("\tGUAMI: nil")
			} else {
				radiusLog.Debugf("\tGUAMI: PLMNIdentity[% x], "+
					"AMFRegionID[% x], AMFSetID[% x], AMFPointer[% x]",
					anParameters.GUAMI.PLMNIdentity, anParameters.GUAMI.AMFRegionID,
					anParameters.GUAMI.AMFSetID, anParameters.GUAMI.AMFPointer)
			}
			if anParameters.SelectedPLMNID == nil {
				radiusLog.Debug("\tSelectedPLMNID: nil")
			} else {
				radiusLog.Debugf("\tSelectedPLMNID: % v", anParameters.SelectedPLMNID.Value)
			}
			if anParameters.RequestedNSSAI == nil {
				radiusLog.Debug("\tRequestedNSSAI: nil")
			} else {
				radiusLog.Debugf("\tRequestedNSSAI:")
				for i := 0; i < len(anParameters.RequestedNSSAI.List); i++ {
					radiusLog.Debugf("\tRequestedNSSAI:")
					radiusLog.Debugf("\t\tSNSSAI %d:", i+1)
					radiusLog.Debugf("\t\t\tSST: % x", anParameters.RequestedNSSAI.List[i].SNSSAI.SST.Value)
					sd := anParameters.RequestedNSSAI.List[i].SNSSAI.SD
					if sd == nil {
						radiusLog.Debugf("\t\t\tSD: nil")
					} else {
						radiusLog.Debugf("\t\t\tSD: % x", sd.Value)
					}
				}
			}
			if anParameters.UEIdentity == nil {
				radiusLog.Debug("\tUEIdentity: nil")
			} else {
				radiusLog.Debugf("\tUEIdentity %v", anParameters.UEIdentity)
			}

			// AMF selection
			selectedAMF := tngfSelf.AMFSelection(anParameters.GUAMI, anParameters.SelectedPLMNID)
			if selectedAMF == nil {
				radiusLog.Warn("No avalible AMF for this UE")

				// Send EAP failure
				// Build Radius message
				responseRadiusMessage.BuildRadiusHeader(radius_message.AccessChallenge, message.PktID, message.Auth)
				responseRadiusMessage.Payloads.Reset()

				// EAP
				identifier, random_err := GenerateRandomUint8()
				if random_err != nil {
					radiusLog.Errorf("Generate random uint8 failed: %+v", random_err)
					return
				}
				responseRadiusPayload.BuildEAPfailure(identifier)

				if requestMessageAuthenticator != nil {
					tmpRadiusMessage := *responseRadiusMessage
					payload := new(radius_message.RadiusPayload)
					payload.Type = radius_message.TypeMessageAuthenticator
					payload.Length = uint8(18)
					payload.Val = make([]byte, 16)

					tmpResponseRadiusPayload := responseRadiusPayload
					tmpResponseRadiusPayload = append(tmpResponseRadiusPayload, *payload)
					tmpRadiusMessage.Payloads = tmpResponseRadiusPayload

					payload.Val = GetMessageAuthenticator(&tmpRadiusMessage)
					responseRadiusPayload = append(responseRadiusPayload, *payload)
				}
				responseRadiusMessage.Payloads = responseRadiusPayload

				// Send Radius message to UE
				SendRadiusMessageToUE(udpConn, tngfAddr, ueAddr, responseRadiusMessage)
				return
			}
			radiusLog.Infof("Selected AMF Name: %s", selectedAMF.AMFName.Value)

			// Create UE context
			ue := tngfSelf.NewTngfUe()

			// Relative context
			session.ThisUE = ue
			session.Auth = message.Auth
			session.PktId = message.PktID
			ue.RadiusSession = session
			ue.AMF = selectedAMF
			ue.UEIdentity = anParameters.UEIdentity

			ue.RadiusConnection = &context.UDPSocketInfo{
				Conn:     udpConn,
				TNGFAddr: tngfAddr,
				UEAddr:   ueAddr,
			}

			// Store some information in conext
			ue.IPAddrv4 = ueAddr.IP.To4().String()
			ue.PortNumber = int32(ueAddr.Port)
			if anParameters.EstablishmentCause != nil {
				ue.RRCEstablishmentCause = int16(anParameters.EstablishmentCause.Value)
			}
			ue.TNAPID = tnapId
			ue.UserName = userName

			// Send Initial UE Message
			ngap_message.SendInitialUEMessage(selectedAMF, ue, nasPDU)
		} else {
			session.Auth = message.Auth
			session.PktId = message.PktID
			ue := session.ThisUE
			amf := ue.AMF

			ue.RadiusConnection = &context.UDPSocketInfo{
				Conn:     udpConn,
				TNGFAddr: tngfAddr,
				UEAddr:   ueAddr,
			}

			// Send Uplink NAS Transport
			ngap_message.SendUplinkNASTransport(amf, ue, nasPDU)
		}

	case InitialContextSetup:
		radiusLog.Infoln("Handle EAP-Res/5G-Notification")
		identifier := eapMessage[1]
		responseRadiusMessage.BuildRadiusHeader(radius_message.AccessAccept, message.PktID, message.Auth)
		responseRadiusPayload.BuildEAPSuccess(identifier)

		// Derivative Ktnap
		thisUE := session.ThisUE
		p0 := []byte{0x2}
		thisUE.Ktnap, err = ueauth.GetKDFValue(thisUE.Ktngf, ueauth.FC_FOR_KTIPSEC_KTNAP_DERIVATION, p0, ueauth.KDFLen(p0))
		if err != nil {
			radiusLog.Errorf("Initial Context Setup GetKDFValue(): %+v", err)
		}
		salt, salt_err := GenerateSalt()
		if salt_err != nil {
			radiusLog.Errorf("Initial Context Setup GenerateSalt(): %+v", salt_err)
		}

		mppeRecvKey, encrypt_err := EncryptMppeKey(thisUE.Ktnap, []byte(tngfSelf.RadiusSecret), message.Auth, salt)
		if encrypt_err != nil {
			radiusLog.Errorf("Initial Context Setup EncryptMppeKey(): %+v", err)
		}
		vendorSpecificData := make([]byte, 2)
		binary.BigEndian.PutUint16(vendorSpecificData, salt)
		vendorSpecificData = append(vendorSpecificData, mppeRecvKey...)

		responseRadiusPayload.BuildMicrosoftVendorSpecific(0x11, vendorSpecificData)
		responseRadiusPayload.BuildMicrosoftVendorSpecific(0x10, vendorSpecificData)
		responseRadiusPayload.BuildTLVPayload(1, []byte(thisUE.UserName))

		if requestMessageAuthenticator != nil {
			tmpRadiusMessage := *responseRadiusMessage
			payload := new(radius_message.RadiusPayload)
			payload.Type = radius_message.TypeMessageAuthenticator
			payload.Length = uint8(18)
			payload.Val = make([]byte, 16)

			tmpResponseRadiusPayload := responseRadiusPayload
			tmpResponseRadiusPayload = append(tmpResponseRadiusPayload, *payload)
			tmpRadiusMessage.Payloads = tmpResponseRadiusPayload

			payload.Val = GetMessageAuthenticator(&tmpRadiusMessage)
			responseRadiusPayload = append(responseRadiusPayload, *payload)
		}
		responseRadiusMessage.Payloads = responseRadiusPayload
		SendRadiusMessageToUE(udpConn, tngfAddr, ueAddr, responseRadiusMessage)
	}
}
