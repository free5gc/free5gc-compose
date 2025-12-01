package radius

import (
	"net"
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/radius/handler"
	radius_message "github.com/free5gc/tngf/pkg/radius/message"
)

var radiusLog *logrus.Entry

func init() {
	radiusLog = logger.RadiusLog
}

func Dispatch(udpConn *net.UDPConn, localAddr, remoteAddr *net.UDPAddr, msg []byte) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.RadiusLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	radiusMessage := new(radius_message.RadiusMessage)

	err := radiusMessage.Decode(msg)
	if err != nil {
		radiusLog.Error(err)
		return
	}

	switch radiusMessage.Code {
	case radius_message.AccessRequest:
		handler.HandleRadiusAccessRequest(udpConn, localAddr, remoteAddr, radiusMessage)
	default:
		radiusLog.Warnf("Unimplemented radius message type, exchange type: %d", radiusMessage.Code)
	}
}
