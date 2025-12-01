package handler

import (
	"net"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	gtp "github.com/wmnsk/go-gtp/gtpv1"
	gtpMsg "github.com/wmnsk/go-gtp/gtpv1/message"

	"github.com/free5gc/tngf/internal/gre"
	gtpQoSMsg "github.com/free5gc/tngf/internal/gtp/message"
	"github.com/free5gc/tngf/internal/logger"
	tngfContext "github.com/free5gc/tngf/pkg/context"
)

var gtpLog *logrus.Entry

func init() {
	gtpLog = logger.GTPLog
}

// Parse the fields not supported by go-gtp and forward data to UE.
func HandleQoSTPDU(c gtp.Conn, senderAddr net.Addr, msg gtpMsg.Message) error {
	pdu := gtpQoSMsg.QoSTPDUPacket{}
	if err := pdu.Unmarshal(msg.(*gtpMsg.TPDU)); err != nil {
		return err
	}

	forward(pdu)
	return nil
}

// Forward user plane packets from N3 to UE with GRE header and new IP header encapsulated
func forward(packet gtpQoSMsg.QoSTPDUPacket) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			gtpLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	// TNGF context
	self := tngfContext.TNGFSelf()

	// Nwu connection in IPv4
	NWtConn := self.NWtIPv4PacketConn

	// Find UE information
	ue, ok := self.AllocatedUETEIDLoad(packet.GetTEID())
	if !ok {
		gtpLog.Error("UE context not found")
		return
	}

	// UE inner IP in IPSec
	ueInnerIPAddr := ue.IPSecInnerIPAddr

	var (
		qfi uint8
		rqi bool
	)

	// QoS Related Parameter
	if packet.HasQoS() {
		qfi, rqi = packet.GetQoSParameters()
		gtpLog.Tracef("QFI: %v, RQI: %v", qfi, rqi)
	}

	// Encasulate IPv4 packet with GRE header before forward to UE through IPsec
	grePacket := gre.GREPacket{}

	// TODO:[24.502(v15.7) 9.3.3 ] The Protocol Type field should be set to zero
	grePacket.SetPayload(packet.GetPayload(), gre.IPv4)
	grePacket.SetQoS(qfi, rqi)
	forwardData := grePacket.Marshal()

	// Send to UE through Nwu
	if n, err := NWtConn.WriteTo(forwardData, nil, ueInnerIPAddr); err != nil {
		gtpLog.Errorf("Write to UE failed: %+v", err)
		return
	} else {
		gtpLog.Trace("Forward NWt <- N3")
		gtpLog.Tracef("Wrote %d bytes", n)
	}
}
