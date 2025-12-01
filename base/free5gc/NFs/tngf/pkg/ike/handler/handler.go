package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/tngf/internal/logger"
	ngap_message "github.com/free5gc/tngf/internal/ngap/message"
	"github.com/free5gc/tngf/pkg/context"
	ike_message "github.com/free5gc/tngf/pkg/ike/message"
	"github.com/free5gc/tngf/pkg/ike/xfrm"
	"github.com/free5gc/util/ueauth"
)

// Log
var ikeLog *logrus.Entry

func init() {
	ikeLog = logger.IKELog
}

func HandleIKESAINIT(udpConn *net.UDPConn, tngfAddr, ueAddr *net.UDPAddr, message *ike_message.IKEMessage) {
	ikeLog.Infoln("Handle IKE_SA_INIT")

	// Used to receive value from peer
	var securityAssociation *ike_message.SecurityAssociation
	var keyExchange *ike_message.KeyExchange
	var nonce *ike_message.Nonce
	var notifications []*ike_message.Notification

	tngfSelf := context.TNGFSelf()

	// For response or needed data
	responseIKEMessage := new(ike_message.IKEMessage)
	var sharedKeyData, localNonce, concatenatedNonce []byte
	// Chosen transform from peer's proposal
	var encryptionAlgorithmTransform, pseudorandomFunctionTransform *ike_message.Transform
	var integrityAlgorithmTransform, diffieHellmanGroupTransform *ike_message.Transform
	// For NAT-T
	var ueIsBehindNAT, tngfIsBehindNAT bool

	if message == nil {
		ikeLog.Error("IKE Message is nil")
		return
	}

	// parse IKE header and setup IKE context
	// check major version
	majorVersion := ((message.Version & 0xf0) >> 4)
	if majorVersion > 2 {
		ikeLog.Warn("Received an IKE message with higher major version")
		// send INFORMATIONAL type message with INVALID_MAJOR_VERSION Notify payload
		responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
			ike_message.INFORMATIONAL, ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()
		responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone,
			ike_message.INVALID_MAJOR_VERSION, nil, nil)

		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

		return
	}

	for _, ikePayload := range message.Payloads {
		switch ikePayload.Type() {
		case ike_message.TypeSA:
			securityAssociation = ikePayload.(*ike_message.SecurityAssociation)
		case ike_message.TypeKE:
			keyExchange = ikePayload.(*ike_message.KeyExchange)
		case ike_message.TypeNiNr:
			nonce = ikePayload.(*ike_message.Nonce)
		case ike_message.TypeN:
			notifications = append(notifications, ikePayload.(*ike_message.Notification))
		default:
			ikeLog.Warnf(
				"Get IKE payload (type %d) in IKE_SA_INIT message, this payload will not be handled by IKE handler",
				ikePayload.Type())
		}
	}

	if securityAssociation != nil {
		responseSecurityAssociation := responseIKEMessage.Payloads.BuildSecurityAssociation()

		for _, proposal := range securityAssociation.Proposals {
			// We need ENCR, PRF, INTEG, DH, but not ESN
			encryptionAlgorithmTransform = nil
			pseudorandomFunctionTransform = nil
			integrityAlgorithmTransform = nil
			diffieHellmanGroupTransform = nil

			if len(proposal.EncryptionAlgorithm) > 0 {
				for _, transform := range proposal.EncryptionAlgorithm {
					if is_supported(ike_message.TypeEncryptionAlgorithm, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						encryptionAlgorithmTransform = transform
						break
					}
					ikeLog.Warn("Not supported encryption algorithm")
				}
				if encryptionAlgorithmTransform == nil {
					continue
				}
			} else {
				continue // mandatory
			}
			if len(proposal.PseudorandomFunction) > 0 {
				for _, transform := range proposal.PseudorandomFunction {
					if is_supported(ike_message.TypePseudorandomFunction, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						pseudorandomFunctionTransform = transform
						break
					}
				}
				if pseudorandomFunctionTransform == nil {
					continue
				}
			} else {
				continue // mandatory
			}
			if len(proposal.IntegrityAlgorithm) > 0 {
				for _, transform := range proposal.IntegrityAlgorithm {
					if is_supported(ike_message.TypeIntegrityAlgorithm, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						integrityAlgorithmTransform = transform
						break
					}
				}
				if integrityAlgorithmTransform == nil {
					continue
				}
			} else {
				continue // mandatory
			}
			if len(proposal.DiffieHellmanGroup) > 0 {
				for _, transform := range proposal.DiffieHellmanGroup {
					if is_supported(ike_message.TypeDiffieHellmanGroup, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						diffieHellmanGroupTransform = transform
						break
					}
				}
				if diffieHellmanGroupTransform == nil {
					continue
				}
			} else {
				continue // mandatory
			}
			if len(proposal.ExtendedSequenceNumbers) > 0 {
				continue // No ESN
			}

			// Construct chosen proposal, with ENCR, PRF, INTEG, DH, and each
			// contains one transform expectively
			chosenProposal := responseSecurityAssociation.Proposals.BuildProposal(
				proposal.ProposalNumber, proposal.ProtocolID, nil)
			chosenProposal.EncryptionAlgorithm = append(chosenProposal.EncryptionAlgorithm, encryptionAlgorithmTransform)
			chosenProposal.PseudorandomFunction = append(chosenProposal.PseudorandomFunction, pseudorandomFunctionTransform)
			chosenProposal.IntegrityAlgorithm = append(chosenProposal.IntegrityAlgorithm, integrityAlgorithmTransform)
			chosenProposal.DiffieHellmanGroup = append(chosenProposal.DiffieHellmanGroup, diffieHellmanGroupTransform)

			break
		}

		if len(responseSecurityAssociation.Proposals) == 0 {
			ikeLog.Warn("No proposal chosen")
			// Respond NO_PROPOSAL_CHOSEN to UE
			responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
				ike_message.IKE_SA_INIT, ike_message.ResponseBitCheck, message.MessageID)
			responseIKEMessage.Payloads.Reset()
			responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone, ike_message.NO_PROPOSAL_CHOSEN, nil, nil)

			SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

			return
		}
	} else {
		ikeLog.Error("The security association field is nil")
		// TODO: send error message to UE
		return
	}

	if keyExchange != nil {
		chosenDiffieHellmanGroup := diffieHellmanGroupTransform.TransformID
		if chosenDiffieHellmanGroup != keyExchange.DiffieHellmanGroup {
			ikeLog.Warn("The Diffie-Hellman group defined in key exchange payload not matches the one in chosen proposal")
			// send INVALID_KE_PAYLOAD to UE
			responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
				ike_message.IKE_SA_INIT, ike_message.ResponseBitCheck, message.MessageID)
			responseIKEMessage.Payloads.Reset()

			notificationData := make([]byte, 2)
			binary.BigEndian.PutUint16(notificationData, chosenDiffieHellmanGroup)
			responseIKEMessage.Payloads.BuildNotification(
				ike_message.TypeNone, ike_message.INVALID_KE_PAYLOAD, nil, notificationData)

			SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

			return
		}

		var localPublicValue []byte

		localPublicValue, sharedKeyData = CalculateDiffieHellmanMaterials(GenerateRandomNumber(),
			keyExchange.KeyExchangeData, chosenDiffieHellmanGroup)
		responseIKEMessage.Payloads.BUildKeyExchange(chosenDiffieHellmanGroup, localPublicValue)
	} else {
		ikeLog.Error("The key exchange field is nil")
		// TODO: send error message to UE
		return
	}

	if nonce != nil {
		localNonce = GenerateRandomNumber().Bytes()
		concatenatedNonce = nonce.NonceData
		concatenatedNonce = append(concatenatedNonce, localNonce...)

		responseIKEMessage.Payloads.BuildNonce(localNonce)
	} else {
		ikeLog.Error("The nonce field is nil")
		// TODO: send error message to UE
		return
	}

	if len(notifications) != 0 {
		for _, notification := range notifications {
			switch notification.NotifyMessageType {
			case ike_message.NAT_DETECTION_SOURCE_IP:
				ikeLog.Trace("Received IKE Notify: NAT_DETECTION_SOURCE_IP")
				// Calculate local NAT_DETECTION_SOURCE_IP hash
				// : sha1(ispi | rspi | ueip | ueport)
				localDetectionData := make([]byte, 22)
				binary.BigEndian.PutUint64(localDetectionData[0:8], message.InitiatorSPI)
				binary.BigEndian.PutUint64(localDetectionData[8:16], message.ResponderSPI)
				copy(localDetectionData[16:20], ueAddr.IP.To4())
				binary.BigEndian.PutUint16(localDetectionData[20:22], uint16(ueAddr.Port))

				sha1HashFunction := sha1.New()
				if _, err := sha1HashFunction.Write(localDetectionData); err != nil {
					ikeLog.Errorf("Hash function write error: %+v", err)
					return
				}

				if !bytes.Equal(notification.NotificationData, sha1HashFunction.Sum(nil)) {
					ueIsBehindNAT = true
				}
			case ike_message.NAT_DETECTION_DESTINATION_IP:
				ikeLog.Trace("Received IKE Notify: NAT_DETECTION_DESTINATION_IP")
				// Calculate local NAT_DETECTION_SOURCE_IP hash
				// : sha1(ispi | rspi | tngfip | tngfport)
				localDetectionData := make([]byte, 22)
				binary.BigEndian.PutUint64(localDetectionData[0:8], message.InitiatorSPI)
				binary.BigEndian.PutUint64(localDetectionData[8:16], message.ResponderSPI)
				copy(localDetectionData[16:20], tngfAddr.IP.To4())
				binary.BigEndian.PutUint16(localDetectionData[20:22], uint16(tngfAddr.Port))

				sha1HashFunction := sha1.New()
				if _, err := sha1HashFunction.Write(localDetectionData); err != nil {
					ikeLog.Errorf("Hash function write error: %+v", err)
					return
				}

				if !bytes.Equal(notification.NotificationData, sha1HashFunction.Sum(nil)) {
					tngfIsBehindNAT = true
				}
			default:
			}
		}
	}

	// Create new IKE security association
	ikeSecurityAssociation := tngfSelf.NewIKESecurityAssociation()
	ikeSecurityAssociation.RemoteSPI = message.InitiatorSPI
	ikeSecurityAssociation.InitiatorMessageID = message.MessageID
	ikeSecurityAssociation.UEIsBehindNAT = ueIsBehindNAT
	ikeSecurityAssociation.TNGFIsBehindNAT = tngfIsBehindNAT

	// Record algorithm in context
	ikeSecurityAssociation.EncryptionAlgorithm = encryptionAlgorithmTransform
	ikeSecurityAssociation.IntegrityAlgorithm = integrityAlgorithmTransform
	ikeSecurityAssociation.PseudorandomFunction = pseudorandomFunctionTransform
	ikeSecurityAssociation.DiffieHellmanGroup = diffieHellmanGroupTransform

	// Record concatenated nonce
	ikeSecurityAssociation.ConcatenatedNonce = append(ikeSecurityAssociation.ConcatenatedNonce, concatenatedNonce...)
	// Record Diffie-Hellman shared key
	ikeSecurityAssociation.DiffieHellmanSharedKey = append(ikeSecurityAssociation.DiffieHellmanSharedKey, sharedKeyData...)

	if err := GenerateKeyForIKESA(ikeSecurityAssociation); err != nil {
		ikeLog.Errorf("Generate key for IKE SA failed: %+v", err)
		return
	}

	// IKE response to UE
	responseIKEMessage.BuildIKEHeader(ikeSecurityAssociation.RemoteSPI, ikeSecurityAssociation.LocalSPI,
		ike_message.IKE_SA_INIT, ike_message.ResponseBitCheck, message.MessageID)

	// Calculate NAT_DETECTION_SOURCE_IP for NAT-T
	natDetectionSourceIP := make([]byte, 22)
	binary.BigEndian.PutUint64(natDetectionSourceIP[0:8], ikeSecurityAssociation.RemoteSPI)
	binary.BigEndian.PutUint64(natDetectionSourceIP[8:16], ikeSecurityAssociation.LocalSPI)
	copy(natDetectionSourceIP[16:20], tngfAddr.IP.To4())
	binary.BigEndian.PutUint16(natDetectionSourceIP[20:22], uint16(tngfAddr.Port))

	// Build and append notify payload for NAT_DETECTION_SOURCE_IP
	responseIKEMessage.Payloads.BuildNotification(
		ike_message.TypeNone, ike_message.NAT_DETECTION_SOURCE_IP, nil, natDetectionSourceIP)

	// Calculate NAT_DETECTION_DESTINATION_IP for NAT-T
	natDetectionDestinationIP := make([]byte, 22)
	binary.BigEndian.PutUint64(natDetectionDestinationIP[0:8], ikeSecurityAssociation.RemoteSPI)
	binary.BigEndian.PutUint64(natDetectionDestinationIP[8:16], ikeSecurityAssociation.LocalSPI)
	copy(natDetectionDestinationIP[16:20], ueAddr.IP.To4())
	binary.BigEndian.PutUint16(natDetectionDestinationIP[20:22], uint16(ueAddr.Port))

	// Build and append notify payload for NAT_DETECTION_DESTINATION_IP
	responseIKEMessage.Payloads.BuildNotification(
		ike_message.TypeNone, ike_message.NAT_DETECTION_DESTINATION_IP, nil, natDetectionDestinationIP)

	// Prepare authentication data - InitatorSignedOctet
	// InitatorSignedOctet = RealMessage1 | NonceRData | MACedIDForI
	// MACedIDForI is acquired in IKE_AUTH exchange
	receivedIKEMessageData, err := message.Encode()
	if err != nil {
		ikeLog.Errorln(err)
		ikeLog.Error("Encode message failed.")
		return
	}
	ikeLog.Infoln("NonceRData: ", hex.Dump(localNonce))
	ikeSecurityAssociation.InitiatorSignedOctets = receivedIKEMessageData
	ikeSecurityAssociation.InitiatorSignedOctets = append(ikeSecurityAssociation.InitiatorSignedOctets, localNonce...)

	// Prepare authentication data - ResponderSignedOctet
	// ResponderSignedOctet = RealMessage2 | NonceIData | MACedIDForR
	responseIKEMessageData, err := responseIKEMessage.Encode()
	if err != nil {
		ikeLog.Errorln(err)
		ikeLog.Error("Encoding IKE message failed")
		return
	}
	ikeSecurityAssociation.ResponderSignedOctets = responseIKEMessageData
	ikeSecurityAssociation.ResponderSignedOctets = append(ikeSecurityAssociation.ResponderSignedOctets, nonce.NonceData...)
	// MACedIDForR
	var idPayload ike_message.IKEPayloadContainer
	idPayload.BuildIdentificationResponder(ike_message.ID_FQDN, []byte(tngfSelf.FQDN))
	idPayloadData, err := idPayload.Encode()
	if err != nil {
		ikeLog.Errorln(err)
		ikeLog.Error("Encode IKE payload failed.")
		return
	}
	pseudorandomFunction, ok := NewPseudorandomFunction(ikeSecurityAssociation.SK_pr,
		ikeSecurityAssociation.PseudorandomFunction.TransformID)
	if !ok {
		ikeLog.Error("Get an unsupported pseudorandom funcion. This may imply an unsupported transform is chosen.")
		return
	}
	if _, random_err := pseudorandomFunction.Write(idPayloadData[4:]); random_err != nil {
		ikeLog.Errorf("Pseudorandom function write error: %+v", random_err)
		return
	}
	ikeSecurityAssociation.ResponderSignedOctets = append(ikeSecurityAssociation.ResponderSignedOctets,
		pseudorandomFunction.Sum(nil)...)

	ikeLog.Tracef("Local unsigned authentication data:\n%s", hex.Dump(ikeSecurityAssociation.ResponderSignedOctets))

	// Send response to UE
	SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)
}

func HandleIKEAUTH(udpConn *net.UDPConn, tngfAddr, ueAddr *net.UDPAddr, message *ike_message.IKEMessage) {
	ikeLog.Infoln("Handle IKE_AUTH")

	var encryptedPayload *ike_message.Encrypted

	tngfSelf := context.TNGFSelf()

	// Used for response
	responseIKEMessage := new(ike_message.IKEMessage)
	var responseIKEPayload ike_message.IKEPayloadContainer

	if message == nil {
		ikeLog.Error("IKE Message is nil")
		return
	}

	// parse IKE header and setup IKE context
	// check major version
	majorVersion := ((message.Version & 0xf0) >> 4)
	if majorVersion > 2 {
		ikeLog.Warn("Received an IKE message with higher major version")
		// send INFORMATIONAL type message with INVALID_MAJOR_VERSION Notify payload ( OUTSIDE IKE SA )
		responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
			ike_message.INFORMATIONAL, ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()
		responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone, ike_message.INVALID_MAJOR_VERSION, nil, nil)

		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

		return
	}

	// Find corresponding IKE security association
	localSPI := message.ResponderSPI
	ikeSecurityAssociation, ok := tngfSelf.IKESALoad(localSPI)
	if !ok {
		ikeLog.Warn("Unrecognized SPI")
		// send INFORMATIONAL type message with INVALID_IKE_SPI Notify payload ( OUTSIDE IKE SA )
		responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, 0, ike_message.INFORMATIONAL,
			ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()
		responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone, ike_message.INVALID_IKE_SPI, nil, nil)

		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

		return
	}

	for _, ikePayload := range message.Payloads {
		switch ikePayload.Type() {
		case ike_message.TypeSK:
			encryptedPayload = ikePayload.(*ike_message.Encrypted)
		default:
			ikeLog.Warnf(
				"Get IKE payload (type %d) in IKE_AUTH message, this payload will not be handled by IKE handler",
				ikePayload.Type())
		}
	}

	decryptedIKEPayload, err := DecryptProcedure(ikeSecurityAssociation, message, encryptedPayload)
	if err != nil {
		ikeLog.Errorf("Decrypt IKE message failed: %+v", err)
		return
	}

	// Parse payloads
	var initiatorID *ike_message.IdentificationInitiator
	// var certificateRequest *ike_message.CertificateRequest
	// var certificate *ike_message.Certificate
	var securityAssociation *ike_message.SecurityAssociation
	var trafficSelectorInitiator *ike_message.TrafficSelectorInitiator
	var trafficSelectorResponder *ike_message.TrafficSelectorResponder
	var authentication *ike_message.Authentication
	var configuration *ike_message.Configuration

	for _, ikePayload := range decryptedIKEPayload {
		switch ikePayload.Type() {
		case ike_message.TypeIDi:
			initiatorID = ikePayload.(*ike_message.IdentificationInitiator)
		case ike_message.TypeSA:
			securityAssociation = ikePayload.(*ike_message.SecurityAssociation)
		case ike_message.TypeTSi:
			trafficSelectorInitiator = ikePayload.(*ike_message.TrafficSelectorInitiator)
		case ike_message.TypeTSr:
			trafficSelectorResponder = ikePayload.(*ike_message.TrafficSelectorResponder)
		case ike_message.TypeAUTH:
			authentication = ikePayload.(*ike_message.Authentication)
		case ike_message.TypeCP:
			configuration = ikePayload.(*ike_message.Configuration)
		default:
			ikeLog.Warnf(
				"Get IKE payload (type %d) in IKE_AUTH message, this payload will not be handled by IKE handler",
				ikePayload.Type())
		}
	}

	// NOTE: tune it
	transformPseudorandomFunction := ikeSecurityAssociation.PseudorandomFunction
	ikeSecurityAssociation.InitiatorMessageID = message.MessageID

	if initiatorID != nil {
		ikeLog.Info("Ecoding initiator for later IKE authentication")
		ikeSecurityAssociation.InitiatorID = initiatorID

		// Record maced identification for authentication
		idPayload := ike_message.IKEPayloadContainer{
			initiatorID,
		}
		idPayloadData, payload_err := idPayload.Encode()
		if payload_err != nil {
			ikeLog.Errorln(payload_err)
			ikeLog.Error("Encoding ID payload message failed.")
			return
		}
		pseudorandomFunction, random_ok := NewPseudorandomFunction(ikeSecurityAssociation.SK_pi,
			transformPseudorandomFunction.TransformID)
		if !random_ok {
			ikeLog.Error("Get an unsupported pseudorandom funcion. This may imply an unsupported transform is chosen.")
			return
		}
		if _, random_err := pseudorandomFunction.Write(idPayloadData[4:]); random_err != nil {
			ikeLog.Errorf("Pseudorandom function write error: %+v", random_err)
			return
		}
		ikeSecurityAssociation.InitiatorSignedOctets = append(
			ikeSecurityAssociation.InitiatorSignedOctets, pseudorandomFunction.Sum(nil)...)

		ikeLog.Infoln("PRF(sk_p,id) :", hex.Dump(pseudorandomFunction.Sum(nil)))
	} else {
		ikeLog.Error("The initiator identification field is nil")
		// TODO: send error message to UE
		return
	}

	if securityAssociation != nil {
		ikeLog.Info("Parsing security association")
		responseSecurityAssociation := new(ike_message.SecurityAssociation)

		for _, proposal := range securityAssociation.Proposals {
			var encryptionAlgorithmTransform *ike_message.Transform = nil
			var integrityAlgorithmTransform *ike_message.Transform = nil
			var diffieHellmanGroupTransform *ike_message.Transform = nil
			var extendedSequenceNumbersTransform *ike_message.Transform = nil

			if len(proposal.SPI) != 4 {
				continue // The SPI of ESP must be 32-bit
			}

			if len(proposal.EncryptionAlgorithm) > 0 {
				for _, transform := range proposal.EncryptionAlgorithm {
					if is_Kernel_Supported(ike_message.TypeEncryptionAlgorithm, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						encryptionAlgorithmTransform = transform
						break
					}
				}
				if encryptionAlgorithmTransform == nil {
					continue
				}
			} else {
				continue // mandatory
			}
			if len(proposal.PseudorandomFunction) > 0 {
				continue // Pseudorandom function is not used by ESP
			}
			if len(proposal.IntegrityAlgorithm) > 0 {
				for _, transform := range proposal.IntegrityAlgorithm {
					if is_Kernel_Supported(ike_message.TypeIntegrityAlgorithm, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						integrityAlgorithmTransform = transform
						break
					}
				}
				if integrityAlgorithmTransform == nil {
					continue
				}
			} // Optional
			if len(proposal.DiffieHellmanGroup) > 0 {
				for _, transform := range proposal.DiffieHellmanGroup {
					if is_Kernel_Supported(ike_message.TypeDiffieHellmanGroup, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						diffieHellmanGroupTransform = transform
						break
					}
				}
				if diffieHellmanGroupTransform == nil {
					continue
				}
			} // Optional
			if len(proposal.ExtendedSequenceNumbers) > 0 {
				for _, transform := range proposal.ExtendedSequenceNumbers {
					if is_Kernel_Supported(ike_message.TypeExtendedSequenceNumbers, transform.TransformID,
						transform.AttributePresent, transform.AttributeValue) {
						extendedSequenceNumbersTransform = transform
						break
					}
				}
				if extendedSequenceNumbersTransform == nil {
					continue
				}
			} else {
				continue // Mandatory
			}

			chosenProposal := responseSecurityAssociation.Proposals.BuildProposal(
				proposal.ProposalNumber, proposal.ProtocolID, proposal.SPI)
			chosenProposal.EncryptionAlgorithm = append(chosenProposal.EncryptionAlgorithm, encryptionAlgorithmTransform)
			chosenProposal.ExtendedSequenceNumbers = append(
				chosenProposal.ExtendedSequenceNumbers, extendedSequenceNumbersTransform)
			if integrityAlgorithmTransform != nil {
				chosenProposal.IntegrityAlgorithm = append(chosenProposal.IntegrityAlgorithm, integrityAlgorithmTransform)
			}
			if diffieHellmanGroupTransform != nil {
				chosenProposal.DiffieHellmanGroup = append(chosenProposal.DiffieHellmanGroup, diffieHellmanGroupTransform)
			}

			break
		}

		if len(responseSecurityAssociation.Proposals) == 0 {
			ikeLog.Warn("No proposal chosen")
			// Respond NO_PROPOSAL_CHOSEN to UE
			// Build IKE message
			responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
				ike_message.IKE_AUTH, ike_message.ResponseBitCheck, message.MessageID)
			responseIKEMessage.Payloads.Reset()

			// Build response
			responseIKEPayload.Reset()

			// Notification
			responseIKEPayload.BuildNotification(ike_message.TypeNone, ike_message.NO_PROPOSAL_CHOSEN, nil, nil)

			if encrypt_err := EncryptProcedure(
				ikeSecurityAssociation, responseIKEPayload, responseIKEMessage); encrypt_err != nil {
				ikeLog.Errorf("Encrypting IKE message failed: %+v", encrypt_err)
				return
			}

			// Send IKE message to UE
			SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

			return
		}

		ikeSecurityAssociation.IKEAuthResponseSA = responseSecurityAssociation
	} else {
		ikeLog.Error("The security association field is nil")
		// TODO: send error message to UE
		return
	}

	if trafficSelectorInitiator != nil {
		ikeLog.Info("Received traffic selector initiator from UE")
		ikeSecurityAssociation.TrafficSelectorInitiator = trafficSelectorInitiator
	} else {
		ikeLog.Error("The initiator traffic selector field is nil")
		// TODO: send error message to UE
		return
	}

	if trafficSelectorResponder != nil {
		ikeLog.Info("Received traffic selector initiator from UE")
		ikeSecurityAssociation.TrafficSelectorResponder = trafficSelectorResponder
	} else {
		ikeLog.Error("The initiator traffic selector field is nil")
		// TODO: send error message to UE
		return
	}

	// Load needed information
	thisUE := tngfSelf.UELoadbyIDi(initiatorID.IDData)
	fmt.Println("initiatorID.IDData: ", string(initiatorID.IDData))
	ikeSecurityAssociation.ThisUE = thisUE
	if thisUE == nil {
		ikeLog.Errorln("UE is nil")
		return
	}

	// Prepare pseudorandom function for calculating/verifying authentication data
	p0 := []byte{0x01}
	thisUE.Ktipsec, err = ueauth.GetKDFValue(thisUE.Ktngf, ueauth.FC_FOR_KTIPSEC_KTNAP_DERIVATION, p0, ueauth.KDFLen(p0))
	if err != nil {
		ikeLog.Error("UE authentication get KDF value error.")
		return
	}
	fmt.Println("ktipsec: ", hex.Dump(thisUE.Ktipsec))
	pseudorandomFunction, ok := NewPseudorandomFunction(thisUE.Ktipsec, transformPseudorandomFunction.TransformID)
	if !ok {
		ikeLog.Error("Get an unsupported pseudorandom funcion. This may imply an unsupported transform is chosen.")
		return
	}
	if _, random_err := pseudorandomFunction.Write([]byte("Key Pad for IKEv2")); random_err != nil {
		ikeLog.Errorf("Pseudorandom function write error: %+v", random_err)
		return
	}
	secret := pseudorandomFunction.Sum(nil)
	ikeLog.Infoln("Using key to authentication:", hex.Dump(secret))
	pseudorandomFunction, ok = NewPseudorandomFunction(secret, transformPseudorandomFunction.TransformID)
	if !ok {
		ikeLog.Error("Get an unsupported pseudorandom funcion. This may imply an unsupported transform is chosen.")
		return
	}

	if authentication != nil {
		// Verifying remote AUTH
		pseudorandomFunction.Reset()
		ikeLog.Infoln("InitoatorSignedOctets: ", hex.Dump(ikeSecurityAssociation.InitiatorSignedOctets))
		if _, random_err := pseudorandomFunction.Write(ikeSecurityAssociation.InitiatorSignedOctets); random_err != nil {
			ikeLog.Errorf("Pseudorandom function write error: %+v", random_err)
			return
		}
		expectedAuthenticationData := pseudorandomFunction.Sum(nil)

		ikeLog.Debugf("Expected Authentication Data:\n%s", hex.Dump(expectedAuthenticationData))
		if !bytes.Equal(authentication.AuthenticationData, expectedAuthenticationData) {
			ikeLog.Debugf("Authentication Data:\n%s", hex.Dump(authentication.AuthenticationData))
			ikeLog.Warn("Peer authentication failed.")
			// Inform UE the authentication has failed
			// Build IKE message
			responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
				ike_message.IKE_AUTH, ike_message.ResponseBitCheck, message.MessageID)
			responseIKEMessage.Payloads.Reset()

			// Notification
			responseIKEPayload.BuildNotification(ike_message.TypeNone, ike_message.AUTHENTICATION_FAILED, nil, nil)

			if encrypt_err := EncryptProcedure(
				ikeSecurityAssociation, responseIKEPayload, responseIKEMessage); encrypt_err != nil {
				ikeLog.Errorf("Encrypting IKE message failed: %+v", encrypt_err)
				return
			}

			// Send IKE message to UE
			SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)
			return
		}
	} else {
		ikeLog.Warn("Peer authentication failed.")
		// Inform UE the authentication has failed
		// Build IKE message
		responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
			ike_message.IKE_AUTH, ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()

		// Notification
		responseIKEPayload.BuildNotification(ike_message.TypeNone, ike_message.AUTHENTICATION_FAILED, nil, nil)

		if encrypt_err := EncryptProcedure(
			ikeSecurityAssociation, responseIKEPayload, responseIKEMessage); encrypt_err != nil {
			ikeLog.Errorf("Encrypting IKE message failed: %+v", encrypt_err)
			return
		}

		// Send IKE message to UE
		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)
		return
	}

	// Parse configuration request to get if the UE has requested internal address,
	// and prepare configuration payload to UE
	addrRequest := false

	if configuration != nil {
		ikeLog.Tracef("Received configuration payload with type: %d", configuration.ConfigurationType)

		var attribute *ike_message.IndividualConfigurationAttribute
		for _, attribute = range configuration.ConfigurationAttribute {
			switch attribute.Type {
			case ike_message.INTERNAL_IP4_ADDRESS:
				addrRequest = true
				if len(attribute.Value) != 0 {
					ikeLog.Tracef("Got client requested address: %d.%d.%d.%d",
						attribute.Value[0], attribute.Value[1], attribute.Value[2], attribute.Value[3])
				}
			default:
				ikeLog.Warnf("Receive other type of configuration request: %d", attribute.Type)
			}
		}
	} else {
		ikeLog.Warn("Configuration is nil. UE did not sent any configuration request.")
	}

	// Build response IKE message
	responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
		ike_message.IKE_AUTH, ike_message.ResponseBitCheck, message.MessageID)
	responseIKEMessage.Payloads.Reset()

	responseIKEPayload.BuildIdentificationResponder(ike_message.ID_FQDN, []byte(tngfSelf.FQDN))

	// Calculate local AUTH
	pseudorandomFunction.Reset()
	if _, random_err := pseudorandomFunction.Write(ikeSecurityAssociation.ResponderSignedOctets); random_err != nil {
		ikeLog.Errorf("Pseudorandom function write error: %+v", random_err)
		return
	}

	// Authentication
	responseIKEPayload.BuildAuthentication(
		ike_message.SharedKeyMesageIntegrityCode, pseudorandomFunction.Sum(nil))

	// Prepare configuration payload and traffic selector payload for initiator and responder
	var ueIPAddr, tngfIPAddr net.IP
	if addrRequest {
		// IP addresses (IPSec)
		ueIPAddr = tngfSelf.NewInternalUEIPAddr(thisUE).To4()
		tngfIPAddr = net.ParseIP(tngfSelf.IPSecGatewayAddress).To4()

		responseConfiguration := responseIKEPayload.BuildConfiguration(ike_message.CFG_REPLY)
		responseConfiguration.ConfigurationAttribute.BuildConfigurationAttribute(ike_message.INTERNAL_IP4_ADDRESS, ueIPAddr)
		responseConfiguration.ConfigurationAttribute.BuildConfigurationAttribute(
			ike_message.INTERNAL_IP4_NETMASK, tngfSelf.Subnet.Mask)

		thisUE.IPSecInnerIP = ueIPAddr
		if ipsecInnerIPAddr, resolve_err := net.ResolveIPAddr("ip", ueIPAddr.String()); resolve_err != nil {
			ikeLog.Errorf("Resolve UE inner IP address failed: %+v", resolve_err)
			return
		} else {
			thisUE.IPSecInnerIPAddr = ipsecInnerIPAddr
		}
		ikeLog.Tracef("ueIPAddr: %+v", ueIPAddr)
	} else {
		ikeLog.Error("UE did not send any configuration request for its IP address.")
		return
	}

	// Security Association
	responseIKEPayload = append(responseIKEPayload, ikeSecurityAssociation.IKEAuthResponseSA)

	// Traffic Selectors initiator/responder
	responseTrafficSelectorInitiator := responseIKEPayload.BuildTrafficSelectorInitiator()
	responseTrafficSelectorInitiator.TrafficSelectors.BuildIndividualTrafficSelector(
		ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll, 0, 65535, ueIPAddr.To4(), ueIPAddr.To4())
	responseTrafficSelectorResponder := responseIKEPayload.BuildTrafficSelectorResponder()
	responseTrafficSelectorResponder.TrafficSelectors.BuildIndividualTrafficSelector(
		ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll, 0, 65535, tngfIPAddr.To4(), tngfIPAddr.To4())

	// Record traffic selector to IKE security association
	ikeSecurityAssociation.TrafficSelectorInitiator = responseTrafficSelectorInitiator
	ikeSecurityAssociation.TrafficSelectorResponder = responseTrafficSelectorResponder

	// Get data needed by xfrm

	// Allocate TNGF inbound SPI
	var inboundSPI uint32
	inboundSPIByte := make([]byte, 4)
	for {
		randomUint64 := GenerateRandomNumber().Uint64()
		// check if the inbound SPI havn't been allocated by TNGF
		if _, load_ok := tngfSelf.ChildSA.Load(uint32(randomUint64)); !load_ok {
			inboundSPI = uint32(randomUint64)
			break
		}
	}
	binary.BigEndian.PutUint32(inboundSPIByte, inboundSPI)

	outboundSPI := binary.BigEndian.Uint32(ikeSecurityAssociation.IKEAuthResponseSA.Proposals[0].SPI)
	ikeLog.Infof("Inbound SPI: %+v, Outbound SPI: %+v", inboundSPI, outboundSPI)

	// SPI field of IKEAuthResponseSA is used to save outbound SPI temporarily.
	// After TNGF produced its inbound SPI, the field will be overwritten with the SPI.
	ikeSecurityAssociation.IKEAuthResponseSA.Proposals[0].SPI = inboundSPIByte

	// Consider 0x01 as the speicified index for IKE_AUTH exchange
	thisUE.CreateHalfChildSA(0x01, inboundSPI, -1)
	childSecurityAssociationContext, err := thisUE.CompleteChildSA(
		0x01, outboundSPI, ikeSecurityAssociation.IKEAuthResponseSA)
	if err != nil {
		ikeLog.Errorf("Create child security association context failed: %+v", err)
		return
	}
	err = parseIPAddressInformationToChildSecurityAssociation(childSecurityAssociationContext, ueAddr.IP,
		ikeSecurityAssociation.TrafficSelectorResponder.TrafficSelectors[0],
		ikeSecurityAssociation.TrafficSelectorInitiator.TrafficSelectors[0])
	if err != nil {
		ikeLog.Errorf("Parse IP address to child security association failed: %+v", err)
		return
	}
	// Select TCP traffic
	childSecurityAssociationContext.SelectedIPProtocol = unix.IPPROTO_TCP

	if errGen := GenerateKeyForChildSA(ikeSecurityAssociation, childSecurityAssociationContext); errGen != nil {
		ikeLog.Errorf("Generate key for child SA failed: %+v", errGen)
		return
	}
	// NAT-T concern
	if ikeSecurityAssociation.UEIsBehindNAT || ikeSecurityAssociation.TNGFIsBehindNAT {
		childSecurityAssociationContext.EnableEncapsulate = true
		childSecurityAssociationContext.TNGFPort = tngfAddr.Port
		childSecurityAssociationContext.NATPort = ueAddr.Port
	}

	// Notification(NAS_IP_ADDRESS)
	responseIKEPayload.BuildNotifyNAS_IP4_ADDRESS(tngfSelf.IPSecGatewayAddress)

	// Notification(NSA_TCP_PORT)
	responseIKEPayload.BuildNotifyNAS_TCP_PORT(tngfSelf.TCPPort)

	if errEncrypt := EncryptProcedure(ikeSecurityAssociation, responseIKEPayload, responseIKEMessage); errEncrypt != nil {
		ikeLog.Errorf("Encrypting IKE message failed: %+v", errEncrypt)
		return
	}

	// Aplly XFRM rules
	// IPsec for CP always use default XFRM interface
	if err = xfrm.ApplyXFRMRule(false, tngfSelf.XfrmIfaceId, childSecurityAssociationContext); err != nil {
		ikeLog.Errorf("Applying XFRM rules failed: %+v", err)
		return
	}

	// Send IKE message to UE
	SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

	// After this, TNGF will forward NAS with Child SA (IPSec SA)
	thisUE.SignallingIPsecSAEstablished = true

	// Store SA into tngfUE
	thisUE.TNGFIKESecurityAssociation = ikeSecurityAssociation

	// Store IKE Connection
	UDPSocket := context.UDPSocketInfo{
		Conn:     udpConn,
		TNGFAddr: tngfAddr,
		UEAddr:   ueAddr,
	}

	thisUE.IKEConnection = &UDPSocket

	// If needed, setup PDU session
	if thisUE.TemporaryPDUSessionSetupData != nil {
		for {
			if len(thisUE.TemporaryPDUSessionSetupData.UnactivatedPDUSession) != 0 {
				pduSessionID := thisUE.TemporaryPDUSessionSetupData.UnactivatedPDUSession[0]
				pduSession := thisUE.PduSessionList[pduSessionID]

				// Add MessageID for IKE security association
				ikeSecurityAssociation.InitiatorMessageID++

				// Send CREATE_CHILD_SA to UE
				ikeMessage := new(ike_message.IKEMessage)
				var ikePayload ike_message.IKEPayloadContainer

				// Build IKE message
				ikeMessage.BuildIKEHeader(ikeSecurityAssociation.LocalSPI,
					ikeSecurityAssociation.RemoteSPI, ike_message.CREATE_CHILD_SA,
					ike_message.InitiatorBitCheck, ikeSecurityAssociation.InitiatorMessageID)
				ikeMessage.Payloads.Reset()

				// Build SA
				requestSA := ikePayload.BuildSecurityAssociation()

				// Allocate SPI
				var spi uint32
				spiByte := make([]byte, 4)
				for {
					randomUint64 := GenerateRandomNumber().Uint64()
					if _, load_ok := tngfSelf.ChildSA.Load(uint32(randomUint64)); !load_ok {
						spi = uint32(randomUint64)
						break
					}
				}
				binary.BigEndian.PutUint32(spiByte, spi)

				// First Proposal - Proposal No.1
				proposal := requestSA.Proposals.BuildProposal(1, ike_message.TypeESP, spiByte)

				// Encryption transform
				proposal.EncryptionAlgorithm.BuildTransform(ike_message.TypeEncryptionAlgorithm,
					ike_message.ENCR_NULL, nil, nil, nil)
				// Integrity transform
				if pduSession.SecurityIntegrity {
					proposal.IntegrityAlgorithm.BuildTransform(
						ike_message.TypeIntegrityAlgorithm, ike_message.AUTH_HMAC_SHA1_96, nil, nil, nil)
				}
				// ESN transform
				proposal.ExtendedSequenceNumbers.BuildTransform(
					ike_message.TypeExtendedSequenceNumbers, ike_message.ESN_NO, nil, nil, nil)

				// Build Nonce
				nonceData := GenerateRandomNumber().Bytes()
				ikePayload.BuildNonce(nonceData)

				// Store nonce into context
				ikeSecurityAssociation.ConcatenatedNonce = nonceData

				// TSi
				tsi_ueIPAddr := thisUE.IPSecInnerIP
				tsi := ikePayload.BuildTrafficSelectorInitiator()
				tsi.TrafficSelectors.BuildIndividualTrafficSelector(ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
					0, 65535, tsi_ueIPAddr, tsi_ueIPAddr)
				// TSr
				tsr_tngfIPAddr := net.ParseIP(tngfSelf.IPSecGatewayAddress)
				tsr := ikePayload.BuildTrafficSelectorResponder()
				tsr.TrafficSelectors.BuildIndividualTrafficSelector(ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
					0, 65535, tsr_tngfIPAddr, tsr_tngfIPAddr)

				// Notify-Qos
				ikePayload.BuildNotify5G_QOS_INFO(uint8(pduSessionID), pduSession.QFIList, true, false, 0)

				// Notify-UP_IP_ADDRESS
				ikePayload.BuildNotifyUP_IP4_ADDRESS(tngfSelf.IPSecGatewayAddress)

				if encrypt_err := EncryptProcedure(
					thisUE.TNGFIKESecurityAssociation, ikePayload, ikeMessage); encrypt_err != nil {
					ikeLog.Errorf("Encrypting IKE message failed: %+v", encrypt_err)
					thisUE.TemporaryPDUSessionSetupData.UnactivatedPDUSession = thisUE.
						TemporaryPDUSessionSetupData.UnactivatedPDUSession[1:]
					cause := ngapType.Cause{
						Present: ngapType.CausePresentTransport,
						Transport: &ngapType.CauseTransport{
							Value: ngapType.CauseTransportPresentTransportResourceUnavailable,
						},
					}
					transfer, pdu_err := ngap_message.BuildPDUSessionResourceSetupUnsuccessfulTransfer(cause, nil)
					if pdu_err != nil {
						ikeLog.Errorf("Build PDU Session Resource Setup Unsuccessful Transfer Failed: %+v", pdu_err)
						continue
					}
					ngap_message.AppendPDUSessionResourceFailedToSetupListCxtRes(
						thisUE.TemporaryPDUSessionSetupData.FailedListCxtRes, pduSessionID, transfer)
					continue
				}

				SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)
				break
			} else {
				// Send Initial Context Setup Response to AMF
				ngap_message.SendInitialContextSetupResponse(thisUE.AMF, thisUE,
					thisUE.TemporaryPDUSessionSetupData.SetupListCxtRes,
					thisUE.TemporaryPDUSessionSetupData.FailedListCxtRes, nil)
				break
			}
		}
	} else {
		// Send Initial Context Setup Response to AMF
		ngap_message.SendInitialContextSetupResponse(thisUE.AMF, thisUE, nil, nil, nil)
	}
}

func HandleCREATECHILDSA(udpConn *net.UDPConn, tngfAddr, ueAddr *net.UDPAddr, message *ike_message.IKEMessage) {
	ikeLog.Infoln("Handle CREATE_CHILD_SA")

	var encryptedPayload *ike_message.Encrypted

	tngfSelf := context.TNGFSelf()

	responseIKEMessage := new(ike_message.IKEMessage)

	if message == nil {
		ikeLog.Error("IKE Message is nil")
		return
	}

	// parse IKE header and setup IKE context
	// check major version
	majorVersion := ((message.Version & 0xf0) >> 4)
	if majorVersion > 2 {
		ikeLog.Warn("Received an IKE message with higher major version")
		// send INFORMATIONAL type message with INVALID_MAJOR_VERSION Notify payload ( OUTSIDE IKE SA )
		responseIKEMessage.BuildIKEHeader(message.InitiatorSPI, message.ResponderSPI,
			ike_message.INFORMATIONAL, ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()
		responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone, ike_message.INVALID_MAJOR_VERSION, nil, nil)

		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

		return
	}

	// Find corresponding IKE security association
	responderSPI := message.ResponderSPI

	ikeLog.Warnf("CREATE_CHILD_SA responderSPI: %+v", responderSPI)
	ikeSecurityAssociation, ok := tngfSelf.IKESALoad(responderSPI)
	if !ok {
		ikeLog.Warn("Unrecognized SPI")
		// send INFORMATIONAL type message with INVALID_IKE_SPI Notify payload ( OUTSIDE IKE SA )
		responseIKEMessage.BuildIKEHeader(0, message.ResponderSPI, ike_message.INFORMATIONAL,
			ike_message.ResponseBitCheck, message.MessageID)
		responseIKEMessage.Payloads.Reset()
		responseIKEMessage.Payloads.BuildNotification(ike_message.TypeNone, ike_message.INVALID_IKE_SPI, nil, nil)

		SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)

		return
	}

	for _, ikePayload := range message.Payloads {
		switch ikePayload.Type() {
		case ike_message.TypeSK:
			encryptedPayload = ikePayload.(*ike_message.Encrypted)
		default:
			ikeLog.Warnf(
				"Get IKE payload (type %d) in CREATE_CHILD_SA message, this payload will not be handled by IKE handler",
				ikePayload.Type())
		}
	}

	decryptedIKEPayload, err := DecryptProcedure(ikeSecurityAssociation, message, encryptedPayload)
	if err != nil {
		ikeLog.Errorf("Decrypt IKE message failed: %+v", err)
		return
	}

	// Parse payloads
	var securityAssociation *ike_message.SecurityAssociation
	var nonce *ike_message.Nonce
	var trafficSelectorInitiator *ike_message.TrafficSelectorInitiator
	var trafficSelectorResponder *ike_message.TrafficSelectorResponder

	for _, ikePayload := range decryptedIKEPayload {
		switch ikePayload.Type() {
		case ike_message.TypeSA:
			securityAssociation = ikePayload.(*ike_message.SecurityAssociation)
		case ike_message.TypeNiNr:
			nonce = ikePayload.(*ike_message.Nonce)
		case ike_message.TypeTSi:
			trafficSelectorInitiator = ikePayload.(*ike_message.TrafficSelectorInitiator)
		case ike_message.TypeTSr:
			trafficSelectorResponder = ikePayload.(*ike_message.TrafficSelectorResponder)
		default:
			ikeLog.Warnf(
				"Get IKE payload (type %d) in CREATE_CHILD_SA message, this payload will not be handled by IKE handler",
				ikePayload.Type())
		}
	}

	// Record message ID
	ikeSecurityAssociation.ResponderMessageID = message.MessageID

	// UE context
	thisUE := ikeSecurityAssociation.ThisUE
	if thisUE == nil {
		ikeLog.Error("UE context is nil")
		return
	}
	// PDU session information
	if thisUE.TemporaryPDUSessionSetupData == nil {
		ikeLog.Error("No PDU session information")
		return
	}
	temporaryPDUSessionSetupData := thisUE.TemporaryPDUSessionSetupData
	if len(temporaryPDUSessionSetupData.UnactivatedPDUSession) == 0 {
		ikeLog.Error("No unactivated PDU session information")
		return
	}
	pduSessionID := temporaryPDUSessionSetupData.UnactivatedPDUSession[0]
	pduSession, ok := thisUE.PduSessionList[pduSessionID]
	if !ok {
		ikeLog.Errorf("No such PDU session [PDU session ID: %d]", pduSessionID)
		return
	}

	// Check received message
	if securityAssociation == nil {
		ikeLog.Error("The security association field is nil")
		return
	}

	if trafficSelectorInitiator == nil {
		ikeLog.Error("The traffic selector initiator field is nil")
		return
	}

	if trafficSelectorResponder == nil {
		ikeLog.Error("The traffic selector responder field is nil")
		return
	}

	// Nonce
	if nonce != nil {
		ikeSecurityAssociation.ConcatenatedNonce = append(ikeSecurityAssociation.ConcatenatedNonce, nonce.NonceData...)
	} else {
		ikeLog.Error("The nonce field is nil")
		// TODO: send error message to UE
		return
	}

	// Get xfrm needed data
	// As specified in RFC 7296, ESP negotiate two child security association (pair) in one exchange
	// Message ID is used to be a index to pair two SPI in serveral IKE messages.
	outboundSPI := binary.BigEndian.Uint32(securityAssociation.Proposals[0].SPI)
	childSecurityAssociationContext, err := thisUE.CompleteChildSA(
		ikeSecurityAssociation.ResponderMessageID, outboundSPI, securityAssociation)
	if err != nil {
		ikeLog.Errorf("Create child security association context failed: %+v", err)
		return
	}

	// Build TSi if there is no one in the response
	if len(trafficSelectorInitiator.TrafficSelectors) == 0 {
		ikeLog.Warnf("There is no TSi in CREATE_CHILD_SA response.")
		tngfIPAddr := net.ParseIP(tngfSelf.IPSecGatewayAddress)
		trafficSelectorInitiator.TrafficSelectors.BuildIndividualTrafficSelector(
			ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
			0, 65535, tngfIPAddr, tngfIPAddr)
	}

	// Build TSr if there is no one in the response
	if len(trafficSelectorResponder.TrafficSelectors) == 0 {
		ikeLog.Warnf("There is no TSr in CREATE_CHILD_SA response.")
		ueIPAddr := thisUE.IPSecInnerIP
		trafficSelectorResponder.TrafficSelectors.BuildIndividualTrafficSelector(
			ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
			0, 65535, ueIPAddr, ueIPAddr)
	}

	err = parseIPAddressInformationToChildSecurityAssociation(childSecurityAssociationContext, ueAddr.IP,
		trafficSelectorInitiator.TrafficSelectors[0], trafficSelectorResponder.TrafficSelectors[0])
	if err != nil {
		ikeLog.Errorf("Parse IP address to child security association failed: %+v", err)
		return
	}
	// Select GRE traffic
	childSecurityAssociationContext.SelectedIPProtocol = unix.IPPROTO_GRE

	if errGen := GenerateKeyForChildSA(ikeSecurityAssociation, childSecurityAssociationContext); errGen != nil {
		ikeLog.Errorf("Generate key for child SA failed: %+v", errGen)
		return
	}
	// NAT-T concern
	if ikeSecurityAssociation.UEIsBehindNAT || ikeSecurityAssociation.TNGFIsBehindNAT {
		childSecurityAssociationContext.EnableEncapsulate = true
		childSecurityAssociationContext.TNGFPort = tngfAddr.Port
		childSecurityAssociationContext.NATPort = ueAddr.Port
	}

	newXfrmiId := tngfSelf.XfrmIfaceId

	// The additional PDU session will be separated from default xfrm interface
	// to avoid SPD entry collision
	if len(thisUE.PduSessionList) > 1 {
		// Setup XFRM interface for ipsec
		var linkIPSec netlink.Link
		tngfIPAddr := net.ParseIP(tngfSelf.IPSecGatewayAddress).To4()
		tngfIPAddrAndSubnet := net.IPNet{IP: tngfIPAddr, Mask: tngfSelf.Subnet.Mask}
		newXfrmiId += tngfSelf.XfrmIfaceId + tngfSelf.XfrmIfaceIdOffsetForUP
		newXfrmiName := fmt.Sprintf("%s-%d", tngfSelf.XfrmIfaceName, newXfrmiId)

		if linkIPSec, err = xfrm.SetupIPsecXfrmi(newXfrmiName, tngfSelf.XfrmParentIfaceName,
			newXfrmiId, tngfIPAddrAndSubnet); err != nil {
			ikeLog.Errorf("Setup XFRM interface %s fail: %+v", newXfrmiName, err)
			return
		}

		tngfSelf.XfrmIfaces.LoadOrStore(newXfrmiId, linkIPSec)
		childSecurityAssociationContext.XfrmIface = linkIPSec
		tngfSelf.XfrmIfaceIdOffsetForUP++
	} else {
		if linkIPSec, load_ok := tngfSelf.XfrmIfaces.Load(newXfrmiId); load_ok {
			childSecurityAssociationContext.XfrmIface = linkIPSec.(netlink.Link)
		} else {
			ikeLog.Warnf("Cannot find the XFRM interface with if_id: %d", newXfrmiId)
		}
	}

	// Aplly XFRM rules
	if err = xfrm.ApplyXFRMRule(true, newXfrmiId, childSecurityAssociationContext); err != nil {
		ikeLog.Errorf("Applying XFRM rules failed: %+v", err)
		return
	} else {
		// Forward PDU Seesion Establishment Accept to UE
		if n, ikeErr := thisUE.TCPConnection.Write(thisUE.TemporaryCachedNASMessage); ikeErr != nil {
			ikeLog.Errorf("Writing via IPSec signaling SA failed: %+v", err)
		} else {
			ikeLog.Tracef("Forward PDU Seesion Establishment Accept to UE. Wrote %d bytes", n)
		}
	}

	// Append NGAP PDU session resource setup response transfer
	transfer, err := ngap_message.BuildPDUSessionResourceSetupResponseTransfer(pduSession)
	if err != nil {
		ikeLog.Errorf("Build PDU session resource setup response transfer failed: %+v", err)
		return
	}
	ngap_message.AppendPDUSessionResourceSetupListSURes(
		temporaryPDUSessionSetupData.SetupListSURes, pduSessionID, transfer)

	// Remove handled PDU session setup request from queue
	temporaryPDUSessionSetupData.UnactivatedPDUSession = temporaryPDUSessionSetupData.UnactivatedPDUSession[1:]

	for {
		if len(temporaryPDUSessionSetupData.UnactivatedPDUSession) != 0 {
			ngapProcedure := temporaryPDUSessionSetupData.NGAPProcedureCode.Value
			tmp_pduSessionID := temporaryPDUSessionSetupData.UnactivatedPDUSession[0]
			tmp_pduSession := thisUE.PduSessionList[tmp_pduSessionID]

			// Add MessageID for IKE security association
			ikeSecurityAssociation.ResponderMessageID++

			// Send CREATE_CHILD_SA to UE
			ikeMessage := new(ike_message.IKEMessage)
			var ikePayload ike_message.IKEPayloadContainer

			// Build IKE message
			ikeMessage.BuildIKEHeader(ikeSecurityAssociation.RemoteSPI,
				ikeSecurityAssociation.LocalSPI, ike_message.CREATE_CHILD_SA,
				ike_message.InitiatorBitCheck, ikeSecurityAssociation.ResponderMessageID)
			ikeMessage.Payloads.Reset()

			// Build SA
			requestSA := ikePayload.BuildSecurityAssociation()

			// Allocate SPI
			var spi uint32
			spiByte := make([]byte, 4)
			for {
				randomUint64 := GenerateRandomNumber().Uint64()
				if _, load_ok := tngfSelf.ChildSA.Load(uint32(randomUint64)); !load_ok {
					spi = uint32(randomUint64)
					break
				}
			}
			binary.BigEndian.PutUint32(spiByte, spi)

			// First Proposal - Proposal No.1
			proposal := requestSA.Proposals.BuildProposal(1, ike_message.TypeESP, spiByte)

			// Encryption transform
			proposal.EncryptionAlgorithm.BuildTransform(ike_message.TypeEncryptionAlgorithm,
				ike_message.ENCR_NULL, nil, nil, nil)
			// Integrity transform
			if tmp_pduSession.SecurityIntegrity {
				proposal.IntegrityAlgorithm.BuildTransform(ike_message.TypeIntegrityAlgorithm,
					ike_message.AUTH_HMAC_MD5_96, nil, nil, nil)
			}
			// ESN transform
			proposal.ExtendedSequenceNumbers.BuildTransform(ike_message.TypeExtendedSequenceNumbers,
				ike_message.ESN_NO, nil, nil, nil)

			// Build Nonce
			nonceData := GenerateRandomNumber().Bytes()
			ikePayload.BuildNonce(nonceData)

			// Store nonce into context
			ikeSecurityAssociation.ConcatenatedNonce = nonceData

			// TSi
			ueIPAddr := thisUE.IPSecInnerIP
			tsi := ikePayload.BuildTrafficSelectorInitiator()
			tsi.TrafficSelectors.BuildIndividualTrafficSelector(ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
				0, 65535, ueIPAddr, ueIPAddr)
			// TSr
			tngfIPAddr := net.ParseIP(tngfSelf.IPSecGatewayAddress)
			tsr := ikePayload.BuildTrafficSelectorResponder()
			tsr.TrafficSelectors.BuildIndividualTrafficSelector(ike_message.TS_IPV4_ADDR_RANGE, ike_message.IPProtocolAll,
				0, 65535, tngfIPAddr, tngfIPAddr)

			// Notify-Qos
			ikePayload.BuildNotify5G_QOS_INFO(uint8(tmp_pduSessionID), tmp_pduSession.QFIList, true, false, 0)

			// Notify-UP_IP_ADDRESS
			ikePayload.BuildNotifyUP_IP4_ADDRESS(tngfSelf.IPSecGatewayAddress)

			if encrypt_err := EncryptProcedure(thisUE.TNGFIKESecurityAssociation, ikePayload, ikeMessage); encrypt_err != nil {
				ikeLog.Errorf("Encrypting IKE message failed: %+v", encrypt_err)
				temporaryPDUSessionSetupData.UnactivatedPDUSession = temporaryPDUSessionSetupData.UnactivatedPDUSession[1:]
				cause := ngapType.Cause{
					Present: ngapType.CausePresentTransport,
					Transport: &ngapType.CauseTransport{
						Value: ngapType.CauseTransportPresentTransportResourceUnavailable,
					},
				}
				resource_transfer, pdusetup_err := ngap_message.BuildPDUSessionResourceSetupUnsuccessfulTransfer(cause, nil)
				if pdusetup_err != nil {
					ikeLog.Errorf("Build PDU Session Resource Setup Unsuccessful Transfer Failed: %+v", pdusetup_err)
					continue
				}
				if ngapProcedure == ngapType.ProcedureCodeInitialContextSetup {
					ngap_message.AppendPDUSessionResourceFailedToSetupListCxtRes(
						temporaryPDUSessionSetupData.FailedListCxtRes, tmp_pduSessionID, resource_transfer)
				} else {
					ngap_message.AppendPDUSessionResourceFailedToSetupListSURes(
						temporaryPDUSessionSetupData.FailedListSURes, tmp_pduSessionID, resource_transfer)
				}
				continue
			}

			SendIKEMessageToUE(udpConn, tngfAddr, ueAddr, responseIKEMessage)
			break
		} else {
			// Send Response to AMF
			ngapProcedure := temporaryPDUSessionSetupData.NGAPProcedureCode.Value
			if ngapProcedure == ngapType.ProcedureCodeInitialContextSetup {
				ngap_message.SendInitialContextSetupResponse(thisUE.AMF, thisUE,
					temporaryPDUSessionSetupData.SetupListCxtRes,
					temporaryPDUSessionSetupData.FailedListCxtRes, nil)
			} else {
				ngap_message.SendPDUSessionResourceSetupResponse(thisUE.AMF, thisUE,
					temporaryPDUSessionSetupData.SetupListSURes,
					temporaryPDUSessionSetupData.FailedListSURes, nil)
			}
			break
		}
	}
}

func is_supported(transformType uint8, transformID uint16, attributePresent bool, attributeValue uint16) bool {
	switch transformType {
	case ike_message.TypeEncryptionAlgorithm:
		switch transformID {
		case ike_message.ENCR_DES_IV64:
			return false
		case ike_message.ENCR_DES:
			return false
		case ike_message.ENCR_3DES:
			return false
		case ike_message.ENCR_RC5:
			return false
		case ike_message.ENCR_IDEA:
			return false
		case ike_message.ENCR_CAST:
			return false
		case ike_message.ENCR_BLOWFISH:
			return false
		case ike_message.ENCR_3IDEA:
			return false
		case ike_message.ENCR_DES_IV32:
			return false
		case ike_message.ENCR_NULL:
			return true
		case ike_message.ENCR_AES_CBC:
			if attributePresent {
				switch attributeValue {
				case 128:
					return true
				case 192:
					return true
				case 256:
					return true
				default:
					return false
				}
			} else {
				return false
			}
		case ike_message.ENCR_AES_CTR:
			return false
		default:
			return false
		}
	case ike_message.TypePseudorandomFunction:
		switch transformID {
		case ike_message.PRF_HMAC_MD5:
			return true
		case ike_message.PRF_HMAC_SHA1:
			return true
		case ike_message.PRF_HMAC_TIGER:
			return false
		default:
			return false
		}
	case ike_message.TypeIntegrityAlgorithm:
		switch transformID {
		case ike_message.AUTH_NONE:
			return false
		case ike_message.AUTH_HMAC_MD5_96:
			return true
		case ike_message.AUTH_HMAC_SHA1_96:
			return true
		case ike_message.AUTH_DES_MAC:
			return false
		case ike_message.AUTH_KPDK_MD5:
			return false
		case ike_message.AUTH_AES_XCBC_96:
			return false
		default:
			return false
		}
	case ike_message.TypeDiffieHellmanGroup:
		switch transformID {
		case ike_message.DH_NONE:
			return false
		case ike_message.DH_768_BIT_MODP:
			return false
		case ike_message.DH_1024_BIT_MODP:
			return true
		case ike_message.DH_1536_BIT_MODP:
			return false
		case ike_message.DH_2048_BIT_MODP:
			return true
		case ike_message.DH_3072_BIT_MODP:
			return false
		case ike_message.DH_4096_BIT_MODP:
			return false
		case ike_message.DH_6144_BIT_MODP:
			return false
		case ike_message.DH_8192_BIT_MODP:
			return false
		default:
			return false
		}
	default:
		return false
	}
}

func is_Kernel_Supported(
	transformType uint8, transformID uint16, attributePresent bool, attributeValue uint16,
) bool {
	switch transformType {
	case ike_message.TypeEncryptionAlgorithm:
		switch transformID {
		case ike_message.ENCR_DES_IV64:
			return false
		case ike_message.ENCR_DES:
			return true
		case ike_message.ENCR_3DES:
			return true
		case ike_message.ENCR_RC5:
			return false
		case ike_message.ENCR_IDEA:
			return false
		case ike_message.ENCR_CAST:
			if attributePresent {
				switch attributeValue {
				case 128:
					return true
				case 256:
					return false
				default:
					return false
				}
			} else {
				return false
			}
		case ike_message.ENCR_BLOWFISH:
			return true
		case ike_message.ENCR_3IDEA:
			return false
		case ike_message.ENCR_DES_IV32:
			return false
		case ike_message.ENCR_NULL:
			return true
		case ike_message.ENCR_AES_CBC:
			if attributePresent {
				switch attributeValue {
				case 128:
					return true
				case 192:
					return true
				case 256:
					return true
				default:
					return false
				}
			} else {
				return false
			}
		case ike_message.ENCR_AES_CTR:
			if attributePresent {
				switch attributeValue {
				case 128:
					return true
				case 192:
					return true
				case 256:
					return true
				default:
					return false
				}
			} else {
				return false
			}
		default:
			return false
		}
	case ike_message.TypeIntegrityAlgorithm:
		switch transformID {
		case ike_message.AUTH_NONE:
			return false
		case ike_message.AUTH_HMAC_MD5_96:
			return true
		case ike_message.AUTH_HMAC_SHA1_96:
			return true
		case ike_message.AUTH_DES_MAC:
			return false
		case ike_message.AUTH_KPDK_MD5:
			return false
		case ike_message.AUTH_AES_XCBC_96:
			return true
		default:
			return false
		}
	case ike_message.TypeDiffieHellmanGroup:
		switch transformID {
		case ike_message.DH_NONE:
			return false
		case ike_message.DH_768_BIT_MODP:
			return false
		case ike_message.DH_1024_BIT_MODP:
			return false
		case ike_message.DH_1536_BIT_MODP:
			return false
		case ike_message.DH_2048_BIT_MODP:
			return false
		case ike_message.DH_3072_BIT_MODP:
			return false
		case ike_message.DH_4096_BIT_MODP:
			return false
		case ike_message.DH_6144_BIT_MODP:
			return false
		case ike_message.DH_8192_BIT_MODP:
			return false
		default:
			return false
		}
	case ike_message.TypeExtendedSequenceNumbers:
		switch transformID {
		case ike_message.ESN_NO:
			return true
		case ike_message.ESN_NEED:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func parseIPAddressInformationToChildSecurityAssociation(
	childSecurityAssociation *context.ChildSecurityAssociation,
	uePublicIPAddr net.IP,
	trafficSelectorLocal *ike_message.IndividualTrafficSelector,
	trafficSelectorRemote *ike_message.IndividualTrafficSelector,
) error {
	if childSecurityAssociation == nil {
		return errors.New("childSecurityAssociation is nil")
	}

	childSecurityAssociation.PeerPublicIPAddr = uePublicIPAddr
	childSecurityAssociation.LocalPublicIPAddr = net.ParseIP(context.TNGFSelf().IKEBindAddress)

	ikeLog.Tracef("Local TS: %+v", trafficSelectorLocal.StartAddress)
	ikeLog.Tracef("Remote TS: %+v", trafficSelectorRemote.StartAddress)

	childSecurityAssociation.TrafficSelectorLocal = net.IPNet{
		IP:   trafficSelectorLocal.StartAddress,
		Mask: []byte{255, 255, 255, 255},
	}

	childSecurityAssociation.TrafficSelectorRemote = net.IPNet{
		IP:   trafficSelectorRemote.StartAddress,
		Mask: []byte{255, 255, 255, 255},
	}

	return nil
}
