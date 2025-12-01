package service

import (
	"errors"
	"net"
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/context"
	"github.com/free5gc/tngf/pkg/ike"
)

var ikeLog *logrus.Entry

func init() {
	// init logger
	ikeLog = logger.IKELog
}

func Run() error {
	// Resolve UDP addresses
	ip := context.TNGFSelf().IKEBindAddress
	udpAddrPort500, err := net.ResolveUDPAddr("udp", ip+":500")
	if err != nil {
		ikeLog.Errorf("Resolve UDP address failed: %+v", err)
		return errors.New("IKE service run failed")
	}
	udpAddrPort4500, err := net.ResolveUDPAddr("udp", ip+":4500")
	if err != nil {
		ikeLog.Errorf("Resolve UDP address failed: %+v", err)
		return errors.New("IKE service run failed")
	}

	// Listen and serve
	var errChan chan error

	// Port 500
	errChan = make(chan error)
	go listenAndServe(udpAddrPort500, errChan)
	if chan_err, ok := <-errChan; ok {
		ikeLog.Errorln(chan_err)
		return errors.New("IKE service run failed")
	}

	// Port 4500
	errChan = make(chan error)
	go listenAndServe(udpAddrPort4500, errChan)
	if chan_err, ok := <-errChan; ok {
		ikeLog.Errorln(chan_err)
		return errors.New("IKE service run failed")
	}

	return nil
}

func listenAndServe(localAddr *net.UDPAddr, errChan chan<- error) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.IKELog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	listener, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		ikeLog.Errorf("Listen UDP failed: %+v", err)
		errChan <- errors.New("listenAndServe failed")
		return
	}

	close(errChan)

	data := make([]byte, 65535)

	for {
		n, remoteAddr, udpread_err := listener.ReadFromUDP(data)
		if udpread_err != nil {
			ikeLog.Errorf("ReadFromUDP failed: %+v", udpread_err)
			continue
		}

		forwardData := make([]byte, n)
		copy(forwardData, data[:n])

		go ike.Dispatch(listener, localAddr, remoteAddr, forwardData)
	}
}
