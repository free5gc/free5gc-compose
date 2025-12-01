package message

import (
	"encoding/hex"
	"errors"

	gtpMessage "github.com/wmnsk/go-gtp/gtpv1/message"

	"github.com/free5gc/tngf/internal/logger"
)

// [TS 38.415] 5.5.2 Frame format for the PDU Session user plane protocol
const (
	DL_PDU_SESSION_INFORMATION_TYPE = 0x00
	UL_PDU_SESSION_INFORMATION_TYPE = 0x10
)

type QoSTPDUPacket struct {
	tPDU *gtpMessage.TPDU
	qos  bool
	rqi  bool
	qfi  uint8
}

func (p *QoSTPDUPacket) GetPayload() []byte {
	return p.tPDU.Payload
}

func (p *QoSTPDUPacket) GetTEID() uint32 {
	return p.tPDU.TEID()
}

func (p *QoSTPDUPacket) GetExtensionHeader() []*gtpMessage.ExtensionHeader {
	return p.tPDU.ExtensionHeaders
}

func (p *QoSTPDUPacket) HasQoS() bool {
	return p.qos
}

func (p *QoSTPDUPacket) GetQoSParameters() (uint8, bool) {
	return p.qfi, p.rqi
}

func (p *QoSTPDUPacket) Unmarshal(pdu *gtpMessage.TPDU) error {
	p.tPDU = pdu
	if p.tPDU.HasExtensionHeader() {
		if err := p.unmarshalExtensionHeader(); err != nil {
			return err
		}
	}

	return nil
}

// [TS 29.281] [TS 38.415]
// Define GTP extension header
// [TS 38.415]
// Define PDU Session User Plane protocol
func (p *QoSTPDUPacket) unmarshalExtensionHeader() error {
	for _, eh := range p.tPDU.ExtensionHeaders {
		switch eh.Type {
		case gtpMessage.ExtHeaderTypePDUSessionContainer:
			p.qos = true
			p.rqi = ((int(eh.Content[1]) >> 6) & 0x1) == 1
			p.qfi = eh.Content[1] & 0x3F
			logger.GTPLog.Tracef("Parsed Extension Header: Len=%d, Next Type=%d, Content Dump:\n%s",
				eh.Length, eh.NextType, hex.Dump(eh.Content))
		default:
			logger.GTPLog.Warningf("Unsupported Extension Header Field Value: %x", eh.Type)
		}
	}

	if !p.qos {
		return errors.New("unmarshalExtensionHeader err: no PDUSessionContainer in ExtensionHeaders")
	}

	return nil
}
