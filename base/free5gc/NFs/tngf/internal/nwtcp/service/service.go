package service

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/internal/ngap/message"
	tngf_context "github.com/free5gc/tngf/pkg/context"
)

var nwtcpLog *logrus.Entry

func init() {
	nwtcpLog = logger.NWtCPLog
}

// Run setup TNGF NAS for UE to forward NAS message
// to AMF
func Run(ctx context.Context) error {
	// TNGF context
	tngfSelf := tngf_context.TNGFSelf()
	tcpAddr := fmt.Sprintf("%s:%d", tngfSelf.IPSecGatewayAddress, tngfSelf.TCPPort)

	var lc net.ListenConfig
	tcpListener, err := lc.Listen(ctx, "tcp", tcpAddr)
	if err != nil {
		nwtcpLog.Errorf("Listen TCP address failed: %+v", err)
		return errors.New("listen failed")
	}

	nwtcpLog.Tracef("Successfully listen %+v", tcpAddr)

	go listenAndServe(tcpListener)

	return nil
}

// listenAndServe handle TCP listener and accept incoming
// requests. It also stores accepted connection into UE
// context, and finally, call serveConn() to serve the messages
// received from the connection.
func listenAndServe(tcpListener net.Listener) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NWtCPLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}

		err := tcpListener.Close()
		if err != nil {
			nwtcpLog.Errorf("Error closing tcpListener: %+v", err)
		}
	}()

	for {
		connection, err := tcpListener.Accept()
		if err != nil {
			nwtcpLog.Error("TCP server accept failed. Close the listener...")
			return
		}

		nwtcpLog.Tracef("Accepted one UE from %+v", connection.RemoteAddr())

		// Find UE context and store this connection in to it, then check if
		// there is any cached NAS message for this UE. If yes, send to it.
		tngfSelf := tngf_context.TNGFSelf()

		ueIP := strings.Split(connection.RemoteAddr().String(), ":")[0]
		ue, ok := tngfSelf.AllocatedUEIPAddressLoad(ueIP)
		if !ok {
			nwtcpLog.Errorf("UE context not found for peer %+v", ueIP)
			continue
		}

		// Store connection
		ue.TCPConnection = connection

		if ue.TemporaryCachedNASMessage != nil {
			// Send to UE
			if n, write_err := connection.Write(ue.TemporaryCachedNASMessage); write_err != nil {
				nwtcpLog.Errorf("Writing via IPSec signaling SA failed: %+v", write_err)
			} else {
				nwtcpLog.Trace("Forward NWt <- N2")
				nwtcpLog.Tracef("Wrote %d bytes", n)
			}
			// Clean the cached message
			ue.TemporaryCachedNASMessage = nil
		}

		go serveConn(ue, connection)
	}
}

func decapNasMsgFromEnvelope(envelop []byte) []byte {
	// According to TS 24.502 8.2.4,
	// in order to transport a NAS message over the non-3GPP access between the UE and the TNGF,
	// the NAS message shall be framed in a NAS message envelope as defined in subclause 9.4.
	// According to TS 24.502 9.4,
	// a NAS message envelope = Length | NAS Message

	// Get NAS Message Length
	nasLen := binary.BigEndian.Uint16(envelop[:2])
	nasMsg := make([]byte, nasLen)
	copy(nasMsg, envelop[2:2+nasLen])

	return nasMsg
}

// serveConn handle accepted TCP connection. It reads NAS packets
// from the connection and call forward() to forward NAS messages
// to AMF
func serveConn(ue *tngf_context.TNGFUe, connection net.Conn) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NWtCPLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}

		err := connection.Close()
		if err != nil {
			nwtcpLog.Errorf("Error closing connection: %+v", err)
		}
	}()

	data := make([]byte, 65535)
	for {
		n, err := connection.Read(data)
		if err != nil {
			if err.Error() == "EOF" {
				nwtcpLog.Warn("Connection close by peer")
				ue.TCPConnection = nil
				return
			} else {
				nwtcpLog.Errorf("Read TCP connection failed: %+v", err)
			}
		}
		nwtcpLog.Tracef("Get NAS PDU from UE:\nNAS length: %d\nNAS content:\n%s", n, hex.Dump(data[:n]))

		// Decap Nas envelope
		forwardData := decapNasMsgFromEnvelope(data)

		go forward(ue, forwardData)
	}
}

// forward forwards NAS messages sent from UE to the
// associated AMF
func forward(ue *tngf_context.TNGFUe, packet []byte) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NWtCPLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	nwtcpLog.Trace("Forward NWt -> N2")
	message.SendUplinkNASTransport(ue.AMF, ue, packet)
}
