package handler

import (
	"net"

	radius_message "github.com/free5gc/tngf/pkg/radius/message"
)

func SendRadiusMessageToUE(udpConn *net.UDPConn, srcAddr, dstAddr *net.UDPAddr, message *radius_message.RadiusMessage) {
	radiusLog.Infoln("Send Radius message to UE")
	radiusLog.Infoln("Encoding...")
	pkt, err := message.Encode()
	if err != nil {
		radiusLog.Errorln(err)
		return
	}

	radiusLog.Trace("Sending...")
	n, err := udpConn.WriteToUDP(pkt, dstAddr)
	if err != nil {
		radiusLog.Error(err)
		return
	}
	if n != len(pkt) {
		radiusLog.Errorf("Not all of the data is sent. Total length: %d. Sent: %d.", len(pkt), n)
		return
	}
}
