package service

import (
	"context"
	"errors"
	"net"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	gtpv1 "github.com/wmnsk/go-gtp/gtpv1"
	gtpMsg "github.com/wmnsk/go-gtp/gtpv1/message"
	"golang.org/x/net/ipv4"

	"github.com/free5gc/tngf/internal/gre"
	gtpQoSMsg "github.com/free5gc/tngf/internal/gtp/message"
	"github.com/free5gc/tngf/internal/logger"
	tngf_context "github.com/free5gc/tngf/pkg/context"
)

var nwtupLog *logrus.Entry

func init() {
	nwtupLog = logger.NWtUPLog
}

// Run bind and listen IPv4 packet connection on TNGF NWt interface
// with UP_IP_ADDRESS, catching GRE encapsulated packets and forward
// to N3 interface.
func Run(ctx context.Context) error {
	// Local IPSec address
	tngfSelf := tngf_context.TNGFSelf()
	listenAddr := tngfSelf.IPSecGatewayAddress

	// Setup IPv4 packet connection socket
	// This socket will only capture GRE encapsulated packet
	var lc net.ListenConfig
	connection, err := lc.ListenPacket(ctx, "ip4:gre", listenAddr)
	if err != nil {
		nwtupLog.Errorf("Error setting listen socket on %s: %+v", listenAddr, err)
		return errors.New("listenPacket failed")
	}
	ipv4PacketConn := ipv4.NewPacketConn(connection)

	tngfSelf.NWtIPv4PacketConn = ipv4PacketConn
	go listenAndServe(ipv4PacketConn)

	return nil
}

// listenAndServe read from socket and call forward() to
// forward packet.
func listenAndServe(ipv4PacketConn *ipv4.PacketConn) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NWtUPLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}

		err := ipv4PacketConn.Close()
		if err != nil {
			nwtupLog.Errorf("Error closing raw socket: %+v", err)
		}
	}()

	buffer := make([]byte, 65535)

	if err := ipv4PacketConn.SetControlMessage(ipv4.FlagInterface|ipv4.FlagTTL, true); err != nil {
		nwtupLog.Errorf("Set control message visibility for IPv4 packet connection fail: %+v", err)
		return
	}

	for {
		n, cm, src, err := ipv4PacketConn.ReadFrom(buffer)
		nwtupLog.Tracef("Read %d bytes, %s", n, cm)
		if err != nil {
			nwtupLog.Errorf("Error read from IPv4 packet connection: %+v", err)
			return
		}

		forwardData := make([]byte, n)
		copy(forwardData, buffer)

		go forward(src.String(), cm.IfIndex, forwardData)
	}
}

// forward forwards user plane packets from NWt to UPF
// with GTP header encapsulated
func forward(ueInnerIP string, ifIndex int, rawData []byte) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NWtUPLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	// Find UE information
	self := tngf_context.TNGFSelf()
	ue, ok := self.AllocatedUEIPAddressLoad(ueInnerIP)
	if !ok {
		nwtupLog.Error("UE context not found")
		return
	}

	var pduSession *tngf_context.PDUSession

	for _, childSA := range ue.TNGFChildSecurityAssociation {
		// Check which child SA the packet come from with interface index,
		// and find the corresponding PDU session
		if childSA.XfrmIface != nil && childSA.XfrmIface.Attrs().Index == ifIndex {
			pduSession = ue.PduSessionList[childSA.PDUSessionIds[0]]
		}
	}

	if pduSession == nil {
		nwtupLog.Error("This UE doesn't have any available PDU session")
		return
	}

	gtpConnection := pduSession.GTPConnection

	userPlaneConnection := gtpConnection.UserPlaneConnection

	// Decapsulate GRE header and extract QoS Parameters if exist
	grePacket := gre.GREPacket{}
	if err := grePacket.Unmarshal(rawData); err != nil {
		nwtupLog.Errorf("gre Unmarshal err: %+v", err)
		return
	}

	var (
		n        int
		writeErr error
	)

	payload, _ := grePacket.GetPayload()

	// Encapsulate UL PDU SESSION INFORMATION with extension header if the QoS parameters exist
	if grePacket.GetKeyFlag() {
		qfi := grePacket.GetQFI()
		gtpPacket, err := buildQoSGTPPacket(gtpConnection.OutgoingTEID, qfi, payload)
		if err != nil {
			nwtupLog.Errorf("buildQoSGTPPacket err: %+v", err)
			return
		}

		n, writeErr = userPlaneConnection.WriteTo(gtpPacket, gtpConnection.UPFUDPAddr)
	} else {
		nwtupLog.Warnf("Receive GRE header without key field specifying QFI and RQI.")
		n, writeErr = userPlaneConnection.WriteToGTP(gtpConnection.OutgoingTEID, payload, gtpConnection.UPFUDPAddr)
	}

	if writeErr != nil {
		nwtupLog.Errorf("Write to UPF failed: %+v", writeErr)
		if writeErr == gtpv1.ErrConnNotOpened {
			nwtupLog.Error("The connection has been closed")
			// TODO: Release the GTP resource
		}
		return
	} else {
		nwtupLog.Trace("Forward NWt -> N3")
		nwtupLog.Tracef("Wrote %d bytes", n)
		return
	}
}

func buildQoSGTPPacket(teid uint32, qfi uint8, payload []byte) ([]byte, error) {
	header := gtpMsg.NewHeader(0x34, gtpMsg.MsgTypeTPDU, teid, 0x00, payload).WithExtensionHeaders(
		gtpMsg.NewExtensionHeader(
			gtpMsg.ExtHeaderTypePDUSessionContainer,
			[]byte{gtpQoSMsg.UL_PDU_SESSION_INFORMATION_TYPE, qfi},
			gtpMsg.ExtHeaderTypeNoMoreExtensionHeaders,
		),
	)

	b := make([]byte, header.MarshalLen())

	if err := header.MarshalTo(b); err != nil {
		nwtupLog.Errorf("go-gtp MarshalTo err: %+v", err)
		return nil, err
	}

	return b, nil
}
