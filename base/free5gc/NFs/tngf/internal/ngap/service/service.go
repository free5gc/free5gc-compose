package service

import (
	"errors"
	"io"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"

	lib_ngap "github.com/free5gc/ngap"
	"github.com/free5gc/sctp"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/internal/ngap"
	"github.com/free5gc/tngf/internal/ngap/message"
	"github.com/free5gc/tngf/pkg/context"
)

var ngapLog *logrus.Entry

func init() {
	ngapLog = logger.NgapLog
}

// Run start the TNGF SCTP process.
func Run() error {
	// tngf context
	tngfSelf := context.TNGFSelf()
	// load amf SCTP address slice
	amfSCTPAddresses := tngfSelf.AMFSCTPAddresses

	localAddr := new(sctp.SCTPAddr)

	for _, remoteAddr := range amfSCTPAddresses {
		errChan := make(chan error)
		go listenAndServe(localAddr, remoteAddr, errChan)
		if err, ok := <-errChan; ok {
			ngapLog.Errorln(err)
			return errors.New("NGAP service run failed")
		}
	}

	return nil
}

func listenAndServe(localAddr, remoteAddr *sctp.SCTPAddr, errChan chan<- error) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.NgapLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	var conn *sctp.SCTPConn
	var err error
	// Connect the session
	for i := 0; i < 3; i++ {
		conn, err = sctp.DialSCTP("sctp", localAddr, remoteAddr)
		if err != nil {
			ngapLog.Errorf("[SCTP] DialSCTP(): %+v", err)
		} else {
			break
		}

		if i != 2 {
			ngapLog.Info("Retry to connect AMF after 1 second...")
			time.Sleep(1 * time.Second)
		} else {
			ngapLog.Debugf("[SCTP] AMF SCTP address: %+v", remoteAddr.String())
			errChan <- errors.New("failed to connect to AMF")
			return
		}
	}

	// Set default sender SCTP information sinfo_ppid = NGAP_PPID = 60
	info, err := conn.GetDefaultSentParam()
	if err != nil {
		ngapLog.Errorf("[SCTP] GetDefaultSentParam(): %+v", err)
		errConn := conn.Close()
		if errConn != nil {
			ngapLog.Errorf("conn close error in GetDefaultSentParam(): %+v", errConn)
		}
		errChan <- errors.New("get socket information failed")
		return
	}
	info.PPID = lib_ngap.PPID
	err = conn.SetDefaultSentParam(info)
	if err != nil {
		ngapLog.Errorf("[SCTP] SetDefaultSentParam(): %+v", err)
		errConn := conn.Close()
		if errConn != nil {
			ngapLog.Errorf("conn close error in SetDefaultSentParam(): %+v", errConn)
		}
		errChan <- errors.New("set socket parameter failed")
		return
	}

	// Subscribe receiver SCTP information
	err = conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)
	if err != nil {
		ngapLog.Errorf("[SCTP] SubscribeEvents(): %+v", err)
		errConn := conn.Close()
		if errConn != nil {
			ngapLog.Errorf("conn close error in SubscribeEvents(): %+v", errConn)
		}
		errChan <- errors.New("subscribe SCTP event failed")
		return
	}

	// Send NG setup request
	go message.SendNGSetupRequest(conn)

	close(errChan)

	data := make([]byte, 65535)

	for {
		n, sctp_info, _, sctpread_err := conn.SCTPRead(data)

		if sctpread_err != nil {
			ngapLog.Debugf("[SCTP] AMF SCTP address: %s", remoteAddr.String())
			if sctpread_err == io.EOF || sctpread_err == io.ErrUnexpectedEOF {
				ngapLog.Warn("[SCTP] Close connection.")
				errConn := conn.Close()
				if errConn != nil {
					ngapLog.Errorf("conn close error: %+v", errConn)
				}
				return
			}
			ngapLog.Errorf("[SCTP] Read from SCTP connection failed: %+v", sctpread_err)
		} else {
			ngapLog.Tracef("[SCTP] Successfully read %d bytes.", n)

			if sctp_info == nil || sctp_info.PPID != lib_ngap.PPID {
				ngapLog.Warn("Received SCTP PPID != 60")
				continue
			}

			forwardData := make([]byte, n)
			copy(forwardData, data[:n])

			go ngap.Dispatch(conn, forwardData)
		}
	}
}
