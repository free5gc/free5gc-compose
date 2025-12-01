package message

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/context"
)

// Log
var radiusLog *logrus.Entry

func init() {
	radiusLog = logger.RadiusLog
}

type RadiusMessage struct {
	Code     uint8
	PktID    uint8
	Length   uint16
	Auth     []byte
	Payloads RadiusPayloadContainer
}

func GetResponseAuth(message []byte) []byte {
	tngfSelf := context.TNGFSelf()
	catData := message
	catData = append(catData, []byte(tngfSelf.RadiusSecret)...)
	responseAuth := md5.Sum(catData)
	return responseAuth[:]
}

func (radiusMessage *RadiusMessage) Encode() ([]byte, error) {
	radiusLog.Debugln("Encoding Radius message")

	radiusMessageData := make([]byte, 4)

	radiusMessageData[0] = radiusMessage.Code
	radiusMessageData[1] = radiusMessage.PktID
	radiusMessageData = append(radiusMessageData, radiusMessage.Auth...)

	radiusMessagePayloadData, err := radiusMessage.Payloads.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode(): EncodePayload failed: %+v", err)
	}
	radiusMessageData = append(radiusMessageData, radiusMessagePayloadData...)
	binary.BigEndian.PutUint16(radiusMessageData[2:4], uint16(len(radiusMessageData)))

	radiusMessage.Auth = GetResponseAuth(radiusMessageData)
	copy(radiusMessageData[4:], radiusMessage.Auth)

	radiusLog.Tracef("Encoded %d bytes", len(radiusMessageData))
	radiusLog.Tracef("Radius message data:\n%s", hex.Dump(radiusMessageData))

	return radiusMessageData, nil
}

func (radiusMessage *RadiusMessage) Decode(rawData []byte) error {
	// Radius message packet format this implementation referenced is
	// defined in RFC 2865, Section 3
	radiusLog.Debugln("Decoding Radius message")
	radiusLog.Tracef("Received Radius message:\n%s", hex.Dump(rawData))

	// bounds checking
	if len(rawData) < 20 {
		return errors.New("decode(): Received broken Radius header")
	}

	radiusMessage.Code = rawData[0]
	radiusMessage.PktID = rawData[1]
	radiusMessage.Length = binary.BigEndian.Uint16(rawData[2:4])
	radiusMessage.Auth = rawData[4:20]

	// fmt.Println(rawData[20:])
	err := radiusMessage.Payloads.Decode(rawData[20:])
	if err != nil {
		return fmt.Errorf("Decode(): DecodePayload failed: %+v", err)
	}

	return nil
}

type RadiusPayloadContainer []RadiusPayload

type RadiusPayload struct {
	Type   uint8
	Length uint8
	Val    []byte
}

func (container *RadiusPayloadContainer) Encode() ([]byte, error) {
	radiusLog.Debugln("Encoding radius payload")
	payloadData := make([]byte, 0)
	for _, payload := range *container {
		data, err := payload.marshal()
		if err != nil {
			return nil, fmt.Errorf("EncodePayload(): Failed to marshal payload: %+v", err)
		}

		payloadData = append(payloadData, data...)
	}
	return payloadData, nil
}

func (container *RadiusPayloadContainer) Decode(rawData []byte) error {
	radiusLog.Debugln("Decoding Radius payloads")

	for len(rawData) > 0 {
		// bounds checking
		radiusLog.Trace("DecodePayload(): Decode 1 payload")
		if len(rawData) <= 2 {
			return errors.New("decodePayload(): No sufficient bytes to decode next payload")
		}
		payloadType := rawData[0]
		payloadLength := rawData[1]
		// fmt.Println(payloadType, payloadLength)

		if payloadLength < 4 {
			return fmt.Errorf("DecodePayload(): Illegal payload length %d < header length 4", payloadLength)
		}
		if len(rawData) < int(payloadLength) {
			return errors.New("decodePayload(): The length of received message not matchs the length specified in header")
		}

		var payload RadiusPayload

		payload.Type = payloadType
		payload.Length = payloadLength
		payload.Val = rawData[2:payloadLength]

		*container = append(*container, payload)

		rawData = rawData[payloadLength:]
	}

	return nil
}

func (radiusPayload *RadiusPayload) marshal() ([]byte, error) {
	radiusLog.Debugln("Radius payload marshal")

	payloadData := make([]byte, 2)
	payloadData[0] = radiusPayload.Type
	payloadData = append(payloadData, radiusPayload.Val...)
	payloadData[1] = uint8(len(payloadData))

	return payloadData, nil
}

// Definition of EAP
type EAP struct {
	Code        uint8
	Identifier  uint8
	EAPTypeData EAPTypeDataContainer
}

func (eap *EAP) Marshal() ([]byte, error) {
	radiusLog.Debugln("[EAP] marshal(): Start marshaling")

	eapData := make([]byte, 4)

	eapData[0] = eap.Code
	eapData[1] = eap.Identifier

	if len(eap.EAPTypeData) > 0 {
		eapTypeData, err := eap.EAPTypeData[0].marshal()
		if err != nil {
			return nil, fmt.Errorf("EAP: EAP type data marshal failed: %+v", err)
		}

		eapData = append(eapData, eapTypeData...)
	}

	binary.BigEndian.PutUint16(eapData[2:4], uint16(len(eapData)))

	return eapData, nil
}

func (eap *EAP) Unmarshal(rawData []byte) error {
	radiusLog.Debugln("[EAP] unmarshal(): Start unmarshalling received bytes")
	radiusLog.Tracef("[EAP] unmarshal(): Payload length %d bytes", len(rawData))

	if len(rawData) > 0 {
		radiusLog.Trace("[EAP] unmarshal(): Unmarshal 1 EAP")
		// bounds checking
		if len(rawData) < 4 {
			return errors.New("EAP: No sufficient bytes to decode next EAP payload")
		}
		eapPayloadLength := binary.BigEndian.Uint16(rawData[2:4])
		if eapPayloadLength < 4 {
			return errors.New("EAP: Payload length specified in the header is too small for EAP")
		}
		if len(rawData) != int(eapPayloadLength) {
			return errors.New("EAP: Received payload length not matches the length specified in header")
		}

		eap.Code = rawData[0]
		eap.Identifier = rawData[1]

		// EAP Success or Failed
		if eapPayloadLength == 4 {
			return nil
		}

		eapType := rawData[4]
		var eapTypeData EAPTypeFormat

		switch eapType {
		case EAPTypeIdentity:
			eapTypeData = new(EAPIdentity)
		case EAPTypeNotification:
			eapTypeData = new(EAPNotification)
		case EAPTypeNak:
			eapTypeData = new(EAPNak)
		case EAPTypeExpanded:
			eapTypeData = new(EAPExpanded)
		default:
			// TODO: Create unsupprted type to handle it
			return errors.New("EAP: Not supported EAP type")
		}

		if err := eapTypeData.unmarshal(rawData[4:]); err != nil {
			return fmt.Errorf("EAP: Unamrshal EAP type data failed: %+v", err)
		}

		eap.EAPTypeData = append(eap.EAPTypeData, eapTypeData)
	}

	return nil
}

type EAPTypeDataContainer []EAPTypeFormat

type EAPTypeFormat interface {
	// Type specifies EAP types
	Type() EAPType

	// Called by EAP.marshal() or EAP.unmarshal()
	marshal() ([]byte, error)
	unmarshal(rawData []byte) error
}

// Definition of EAP Identity

var _ EAPTypeFormat = &EAPIdentity{}

type EAPIdentity struct {
	IdentityData []byte
}

func (eapIdentity *EAPIdentity) Type() EAPType { return EAPTypeIdentity }

func (eapIdentity *EAPIdentity) marshal() ([]byte, error) {
	radiusLog.Debugln("[EAP][Identity] marshal(): Start marshaling")

	if len(eapIdentity.IdentityData) == 0 {
		return nil, errors.New("EAPIdentity: EAP identity is empty")
	}

	eapIdentityData := []byte{EAPTypeIdentity}
	eapIdentityData = append(eapIdentityData, eapIdentity.IdentityData...)

	return eapIdentityData, nil
}

func (eapIdentity *EAPIdentity) unmarshal(rawData []byte) error {
	radiusLog.Debugln("[EAP][Identity] unmarshal(): Start unmarshalling received bytes")
	radiusLog.Tracef("[EAP][Identity] unmarshal(): Payload length %d bytes", len(rawData))

	if len(rawData) > 1 {
		eapIdentity.IdentityData = append(eapIdentity.IdentityData, rawData[1:]...)
	}

	return nil
}

// Definition of EAP Notification

var _ EAPTypeFormat = &EAPNotification{}

type EAPNotification struct {
	NotificationData []byte
}

func (eapNotification *EAPNotification) Type() EAPType { return EAPTypeNotification }

func (eapNotification *EAPNotification) marshal() ([]byte, error) {
	radiusLog.Debugln("[EAP][Notification] marshal(): Start marshaling")

	if len(eapNotification.NotificationData) == 0 {
		return nil, errors.New("EAPNotification: EAP notification is empty")
	}

	eapNotificationData := []byte{EAPTypeNotification}
	eapNotificationData = append(eapNotificationData, eapNotification.NotificationData...)

	return eapNotificationData, nil
}

func (eapNotification *EAPNotification) unmarshal(rawData []byte) error {
	radiusLog.Debugln("[EAP][Notification] unmarshal(): Start unmarshalling received bytes")
	radiusLog.Tracef("[EAP][Notification] unmarshal(): Payload length %d bytes", len(rawData))

	if len(rawData) > 1 {
		eapNotification.NotificationData = append(eapNotification.NotificationData, rawData[1:]...)
	}

	return nil
}

// Definition of EAP Nak

var _ EAPTypeFormat = &EAPNak{}

type EAPNak struct {
	NakData []byte
}

func (eapNak *EAPNak) Type() EAPType { return EAPTypeNak }

func (eapNak *EAPNak) marshal() ([]byte, error) {
	radiusLog.Debugln("[EAP][Nak] marshal(): Start marshaling")

	if len(eapNak.NakData) == 0 {
		return nil, errors.New("EAPNak: EAP nak is empty")
	}

	eapNakData := []byte{EAPTypeNak}
	eapNakData = append(eapNakData, eapNak.NakData...)

	return eapNakData, nil
}

func (eapNak *EAPNak) unmarshal(rawData []byte) error {
	radiusLog.Debugln("[EAP][Nak] unmarshal(): Start unmarshalling received bytes")
	radiusLog.Tracef("[EAP][Nak] unmarshal(): Payload length %d bytes", len(rawData))

	if len(rawData) > 1 {
		eapNak.NakData = append(eapNak.NakData, rawData[1:]...)
	}

	return nil
}

// Definition of EAP expanded

var _ EAPTypeFormat = &EAPExpanded{}

type EAPExpanded struct {
	VendorID   uint32
	VendorType uint32
	VendorData []byte
}

func (eapExpanded *EAPExpanded) Type() EAPType { return EAPTypeExpanded }

func (eapExpanded *EAPExpanded) marshal() ([]byte, error) {
	radiusLog.Debugln("[EAP][Expanded] marshal(): Start marshaling")

	eapExpandedData := make([]byte, 8)

	vendorID := eapExpanded.VendorID & 0x00ffffff
	typeAndVendorID := (uint32(EAPTypeExpanded)<<24 | vendorID)

	binary.BigEndian.PutUint32(eapExpandedData[0:4], typeAndVendorID)
	binary.BigEndian.PutUint32(eapExpandedData[4:8], eapExpanded.VendorType)

	if len(eapExpanded.VendorData) == 0 {
		radiusLog.Warn("[EAP][Expanded] marshal(): EAP vendor data field is empty")
		return eapExpandedData, nil
	}

	eapExpandedData = append(eapExpandedData, eapExpanded.VendorData...)

	return eapExpandedData, nil
}

func (eapExpanded *EAPExpanded) unmarshal(rawData []byte) error {
	radiusLog.Debugln("[EAP][Expanded] unmarshal(): Start unmarshalling received bytes")
	radiusLog.Tracef("[EAP][Expanded] unmarshal(): Payload length %d bytes", len(rawData))

	if len(rawData) > 0 {
		if len(rawData) < 8 {
			return errors.New("EAPExpanded: No sufficient bytes to decode the EAP expanded type")
		}

		typeAndVendorID := binary.BigEndian.Uint32(rawData[0:4])
		eapExpanded.VendorID = typeAndVendorID & 0x00ffffff

		eapExpanded.VendorType = binary.BigEndian.Uint32(rawData[4:8])

		if len(rawData) > 8 {
			eapExpanded.VendorData = append(eapExpanded.VendorData, rawData[8:]...)
		}
	}

	return nil
}

// Definition of EAP
type RadiusMicrosoftVendorSpecific struct {
	Type   uint8
	Length uint8
	String []byte
}

func (vendorSpecific *RadiusMicrosoftVendorSpecific) marshal() ([]byte, error) {
	radiusLog.Debugln("[RADIUS][VendorSpecific] marshal(): Start marshaling")

	vendorSpecificData := make([]byte, 2)

	vendorSpecificData[0] = vendorSpecific.Type
	vendorSpecificData = append(vendorSpecificData, vendorSpecific.String...)
	vendorSpecificData[1] = uint8(len(vendorSpecificData))

	return vendorSpecificData, nil
}
