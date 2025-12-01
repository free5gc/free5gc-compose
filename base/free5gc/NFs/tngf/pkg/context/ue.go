package context

import (
	"errors"
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	gtpv1 "github.com/wmnsk/go-gtp/gtpv1"

	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/ngap/ngapType"
	ike_message "github.com/free5gc/tngf/pkg/ike/message"
)

const (
	AmfUeNgapIdUnspecified int64 = 0xffffffffff
)

type RadiusSession struct {
	CallingStationID string
	State            uint8

	// UE context
	ThisUE *TNGFUe

	// RADIUS Info
	Auth  []byte
	PktId uint8
}

type TNGFUe struct {
	/* UE identity */
	RanUeNgapId      int64
	AmfUeNgapId      int64
	IPAddrv4         string
	IPAddrv6         string
	PortNumber       int32
	TNAPID           uint64
	MaskedIMEISV     *ngapType.MaskedIMEISV // TS 38.413 9.3.1.54
	Guti             string
	IPSecInnerIP     net.IP
	IPSecInnerIPAddr *net.IPAddr // Used to send UP packets to UE

	/* Relative Context */
	AMF *TNGFAMF

	/* PDU Session */
	PduSessionList map[int64]*PDUSession // pduSessionId as key

	/* PDU Session Setup Temporary Data */
	TemporaryPDUSessionSetupData *PDUSessionSetupTemporaryData

	/* Temporary cached NAS message */
	// Used when NAS registration accept arrived before
	// UE setup NAS TCP connection with TNGF, and
	// Forward pduSessionEstablishmentAccept to UE after
	// UE send CREATE_CHILD_SA response
	TemporaryCachedNASMessage []byte

	/* Security */
	Ktngf                []uint8                          // 32 bytes (256 bits), value is from NGAP IE "Security Key"
	Ktnap                []uint8                          // 32 bytes (256 bits), value is computed from Ktngf
	Ktipsec              []uint8                          // 32 bytes (256 bits), value is computed from Ktngf
	SecurityCapabilities *ngapType.UESecurityCapabilities // TS 38.413 9.3.1.86

	/* IKE Security Association */
	TNGFIKESecurityAssociation   *IKESecurityAssociation
	TNGFChildSecurityAssociation map[uint32]*ChildSecurityAssociation // inbound SPI as key
	SignallingIPsecSAEstablished bool

	// RADIUS Session
	RadiusSession *RadiusSession

	/* Temporary Mapping of two SPIs */
	// Exchange Message ID(including a SPI) and ChildSA(including a SPI)
	// Mapping of Message ID of exchange in IKE and Child SA when creating new child SA
	TemporaryExchangeMsgIDChildSAMapping map[uint32]*ChildSecurityAssociation // Message ID as a key

	/* NAS IKE Connection */
	IKEConnection *UDPSocketInfo
	/* NAS TCP Connection */
	TCPConnection net.Conn
	// RADIUS Connection
	RadiusConnection *UDPSocketInfo

	/* Others */
	Guami                            *ngapType.GUAMI
	IndexToRfsp                      int64
	Ambr                             *ngapType.UEAggregateMaximumBitRate
	AllowedNssai                     *ngapType.AllowedNSSAI
	RadioCapability                  *ngapType.UERadioCapability                // TODO: This is for RRC, can be deleted
	CoreNetworkAssistanceInformation *ngapType.CoreNetworkAssistanceInformation // TS 38.413 9.3.1.15
	IMSVoiceSupported                int32
	RRCEstablishmentCause            int16
	UserName                         string
	UEIdentity                       *nasType.MobileIdentity5GS
}

type PDUSession struct {
	Id                               int64 // PDU Session ID
	Type                             *ngapType.PDUSessionType
	Ambr                             *ngapType.PDUSessionAggregateMaximumBitRate
	Snssai                           ngapType.SNSSAI
	NetworkInstance                  *ngapType.NetworkInstance
	SecurityCipher                   bool
	SecurityIntegrity                bool
	MaximumIntegrityDataRateUplink   *ngapType.MaximumIntegrityProtectedDataRate
	MaximumIntegrityDataRateDownlink *ngapType.MaximumIntegrityProtectedDataRate
	GTPConnection                    *GTPConnectionInfo
	QFIList                          []uint8
	QosFlows                         map[int64]*QosFlow // QosFlowIdentifier as key
}

type PDUSessionSetupTemporaryData struct {
	// Slice of unactivated PDU session
	UnactivatedPDUSession []int64 // PDUSessionID as content
	// NGAPProcedureCode is used to identify which type of
	// response shall be used
	NGAPProcedureCode ngapType.ProcedureCode
	// PDU session setup list response
	SetupListCxtRes  *ngapType.PDUSessionResourceSetupListCxtRes
	FailedListCxtRes *ngapType.PDUSessionResourceFailedToSetupListCxtRes
	SetupListSURes   *ngapType.PDUSessionResourceSetupListSURes
	FailedListSURes  *ngapType.PDUSessionResourceFailedToSetupListSURes
}

type QosFlow struct {
	Identifier int64
	Parameters ngapType.QosFlowLevelQosParameters
}

type GTPConnectionInfo struct {
	UPFIPAddr           string
	UPFUDPAddr          net.Addr
	IncomingTEID        uint32
	OutgoingTEID        uint32
	UserPlaneConnection *gtpv1.UPlaneConn
}

type IKESecurityAssociation struct {
	// SPI
	RemoteSPI uint64
	LocalSPI  uint64

	// Message ID
	InitiatorMessageID uint32
	ResponderMessageID uint32

	// Transforms for IKE SA
	EncryptionAlgorithm    *ike_message.Transform
	PseudorandomFunction   *ike_message.Transform
	IntegrityAlgorithm     *ike_message.Transform
	DiffieHellmanGroup     *ike_message.Transform
	ExpandedSequenceNumber *ike_message.Transform

	// Used for key generating
	ConcatenatedNonce      []byte
	DiffieHellmanSharedKey []byte

	// Keys
	SK_d  []byte // used for child SA key deriving
	SK_ai []byte // used by initiator for integrity checking
	SK_ar []byte // used by responder for integrity checking
	SK_ei []byte // used by initiator for encrypting
	SK_er []byte // used by responder for encrypting
	SK_pi []byte // used by initiator for IKE authentication
	SK_pr []byte // used by responder for IKE authentication

	// State for IKE_AUTH
	State uint8

	// Temporary data stored for the use in later exchange
	InitiatorID              *ike_message.IdentificationInitiator
	InitiatorCertificate     *ike_message.Certificate
	IKEAuthResponseSA        *ike_message.SecurityAssociation
	TrafficSelectorInitiator *ike_message.TrafficSelectorInitiator
	TrafficSelectorResponder *ike_message.TrafficSelectorResponder
	LastEAPIdentifier        uint8

	// Authentication data
	ResponderSignedOctets []byte
	InitiatorSignedOctets []byte

	// NAT detection
	// If UEIsBehindNAT == true, TNGF should enable NAT traversal and
	// TODO: should support dynamic updating network address (MOBIKE)
	UEIsBehindNAT bool
	// If TNGFIsBehindNAT == true, TNGF should send UDP keepalive periodically
	TNGFIsBehindNAT bool

	// UE context
	ThisUE *TNGFUe
}

type ChildSecurityAssociation struct {
	// SPI
	InboundSPI  uint32 // TNGF Specify
	OutboundSPI uint32 // Non-3GPP UE Specify

	// Associated XFRM interface
	XfrmIface netlink.Link

	// IP address
	PeerPublicIPAddr  net.IP
	LocalPublicIPAddr net.IP

	// Traffic selector
	SelectedIPProtocol    uint8
	TrafficSelectorLocal  net.IPNet
	TrafficSelectorRemote net.IPNet

	// Security
	EncryptionAlgorithm               uint16
	InitiatorToResponderEncryptionKey []byte
	ResponderToInitiatorEncryptionKey []byte
	IntegrityAlgorithm                uint16
	InitiatorToResponderIntegrityKey  []byte
	ResponderToInitiatorIntegrityKey  []byte
	ESN                               bool

	// Encapsulate
	EnableEncapsulate bool
	TNGFPort          int
	NATPort           int

	// PDU Session IDs associated with this child SA
	PDUSessionIds []int64

	// UE context
	ThisUE *TNGFUe
}

type UDPSocketInfo struct {
	Conn     *net.UDPConn
	TNGFAddr *net.UDPAddr
	UEAddr   *net.UDPAddr
}

func (ue *TNGFUe) init(ranUeNgapId int64) {
	ue.RanUeNgapId = ranUeNgapId
	ue.AmfUeNgapId = AmfUeNgapIdUnspecified
	ue.PduSessionList = make(map[int64]*PDUSession)
	ue.TNGFChildSecurityAssociation = make(map[uint32]*ChildSecurityAssociation)
	ue.TemporaryExchangeMsgIDChildSAMapping = make(map[uint32]*ChildSecurityAssociation)
}

func (ue *TNGFUe) Remove() {
	// remove from AMF context
	ue.DetachAMF()
	// remove from TNGF context
	tngfSelf := TNGFSelf()
	tngfSelf.DeleteTngfUe(ue.RanUeNgapId)
	tngfSelf.DeleteIKESecurityAssociation(ue.TNGFIKESecurityAssociation.LocalSPI)
	tngfSelf.DeleteInternalUEIPAddr(ue.IPSecInnerIP.String())
	for _, pduSession := range ue.PduSessionList {
		tngfSelf.DeleteTEID(pduSession.GTPConnection.IncomingTEID)
	}
}

func (ue *TNGFUe) FindPDUSession(pduSessionID int64) *PDUSession {
	if pduSession, ok := ue.PduSessionList[pduSessionID]; ok {
		return pduSession
	} else {
		return nil
	}
}

func (ue *TNGFUe) CreatePDUSession(pduSessionID int64, snssai ngapType.SNSSAI) (*PDUSession, error) {
	if _, exists := ue.PduSessionList[pduSessionID]; exists {
		return nil, fmt.Errorf("PDU Session[ID:%d] is already exists", pduSessionID)
	}
	pduSession := &PDUSession{
		Id:       pduSessionID,
		Snssai:   snssai,
		QosFlows: make(map[int64]*QosFlow),
	}
	ue.PduSessionList[pduSessionID] = pduSession
	return pduSession, nil
}

// When TNGF send CREATE_CHILD_SA request to N3UE, the inbound SPI of childSA will be only stored first until
// receive response and call CompleteChildSAWithProposal to fill the all data of childSA
func (ue *TNGFUe) CreateHalfChildSA(msgID, inboundSPI uint32, pduSessionID int64) {
	childSA := new(ChildSecurityAssociation)
	childSA.InboundSPI = inboundSPI
	childSA.PDUSessionIds = append(childSA.PDUSessionIds, pduSessionID)
	// Link UE context
	childSA.ThisUE = ue
	// Map Exchange Message ID and Child SA data until get paired response
	ue.TemporaryExchangeMsgIDChildSAMapping[msgID] = childSA
}

func (ue *TNGFUe) CompleteChildSA(msgID uint32, outboundSPI uint32,
	chosenSecurityAssociation *ike_message.SecurityAssociation,
) (*ChildSecurityAssociation, error) {
	childSA, ok := ue.TemporaryExchangeMsgIDChildSAMapping[msgID]

	if !ok {
		return nil, fmt.Errorf("there's not a half child SA created by the exchange with message ID %d", msgID)
	}

	// Remove mapping of exchange msg ID and child SA
	delete(ue.TemporaryExchangeMsgIDChildSAMapping, msgID)

	if chosenSecurityAssociation == nil {
		return nil, errors.New("chosenSecurityAssociation is nil")
	}

	if len(chosenSecurityAssociation.Proposals) == 0 {
		return nil, errors.New("no proposal")
	}

	childSA.OutboundSPI = outboundSPI

	if len(chosenSecurityAssociation.Proposals[0].EncryptionAlgorithm) != 0 {
		childSA.EncryptionAlgorithm = chosenSecurityAssociation.Proposals[0].EncryptionAlgorithm[0].TransformID
	}
	if len(chosenSecurityAssociation.Proposals[0].IntegrityAlgorithm) != 0 {
		childSA.IntegrityAlgorithm = chosenSecurityAssociation.Proposals[0].IntegrityAlgorithm[0].TransformID
	}
	if len(chosenSecurityAssociation.Proposals[0].ExtendedSequenceNumbers) != 0 {
		if chosenSecurityAssociation.Proposals[0].ExtendedSequenceNumbers[0].TransformID == 0 {
			childSA.ESN = false
		} else {
			childSA.ESN = true
		}
	}

	// Record to UE context with inbound SPI as key
	ue.TNGFChildSecurityAssociation[childSA.InboundSPI] = childSA
	// Record to TNGF context with inbound SPI as key
	tngfContext.ChildSA.Store(childSA.InboundSPI, childSA)

	return childSA, nil
}

func (ue *TNGFUe) AttachAMF(sctpAddr string) bool {
	if amf, ok := tngfContext.AMFPoolLoad(sctpAddr); ok {
		amf.TngfUeList[ue.RanUeNgapId] = ue
		ue.AMF = amf
		return true
	} else {
		return false
	}
}

func (ue *TNGFUe) DetachAMF() {
	if ue.AMF == nil {
		return
	}
	delete(ue.AMF.TngfUeList, ue.RanUeNgapId)
}
