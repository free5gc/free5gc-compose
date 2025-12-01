// ftpserver allows to create your own FTP(S) server
package billing

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/fclairamb/ftpserver/config"
	ftpconf "github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/server"
	ftpserver "github.com/fclairamb/ftpserverlib"

	"github.com/free5gc/webconsole/backend/factory"
	"github.com/free5gc/webconsole/backend/logger"
)

// File permission constants
const (
	DefaultFilePermission = 0o600
)

type BillingDomain struct {
	ftpServer *ftpserver.FtpServer
	driver    *server.Server
	wg        *sync.WaitGroup
}

// The ftp server is for CDR Push method, that is the CHF will send the CDR file to the FTP server
func OpenServer(wg *sync.WaitGroup) *BillingDomain {
	b := &BillingDomain{
		wg: wg,
	}

	billingConfig := factory.WebuiConfig.Configuration.BillingServer

	basePath := billingConfig.BastPath
	confFile := path.Join(basePath, "ftpserver.json")

	if _, err := os.Stat(basePath); err != nil {
		if err_mk := os.Mkdir(basePath, os.ModePerm); err_mk != nil {
			logger.BillingLog.Error(err_mk)
		}
	}

	addr := billingConfig.HostIPv4 + ":" + strconv.Itoa(billingConfig.ListenPort)

	params := map[string]string{
		"basePath": basePath,
	}

	logger.BillingLog.Infof("Open BillingServer on %+v", basePath)

	if billingConfig.Cert != nil {
		params["cert"] = billingConfig.Cert.Pem
		params["key"] = billingConfig.Cert.Key
		logger.BillingLog.Infof("Use tls: %+v, %+v", params["cert"], params["key"])
	}

	ftpConfig := ftpconf.Content{
		Version: 1,
		Accesses: []*ftpconf.Access{
			{
				User:   "admin",
				Pass:   "free5gc",
				Fs:     "os",
				Params: params,
			},
		},
		PassiveTransferPortRange: &ftpconf.PortRange{
			Start: billingConfig.PortRange.Start,
			End:   billingConfig.PortRange.End,
		},
		ListenAddress: addr,
	}

	file, err := json.MarshalIndent(ftpConfig, "", " ")
	if err != nil {
		logger.BillingLog.Errorf("Couldn't Marshal conf file %v", err)
		return nil
	}

	if err = os.WriteFile(confFile, file, DefaultFilePermission); err != nil {
		logger.BillingLog.Errorf("Couldn't create conf file %v, err: %+v", confFile, err)
		return nil
	}

	conf, errConfig := config.NewConfig(confFile, logger.FtpServerLog)
	if errConfig != nil {
		logger.BillingLog.Error("Can't load conf", "Err", errConfig)
		return nil
	}

	// Loading the driver
	var errNewServer error
	b.driver, errNewServer = server.NewServer(conf, logger.FtpServerLog)

	if errNewServer != nil {
		logger.BillingLog.Error("Could not load the driver", "err", errNewServer)

		return nil
	}

	// Instantiating the server by passing our driver implementation
	b.ftpServer = ftpserver.NewFtpServer(b.driver)

	// Setting up the ftpserver logger
	b.ftpServer.Logger = logger.FtpServerLog

	go b.Serve()
	logger.BillingLog.Info("Billing server started")

	return b
}

func (b *BillingDomain) Serve() {
	if err := b.ftpServer.ListenAndServe(); err != nil {
		logger.BillingLog.Error("Problem listening ", "err", err)
	}
}

func (b *BillingDomain) Stop() {
	logger.BillingLog.Infoln("Stop BillingDomain server")

	b.driver.Stop()
	if err := b.ftpServer.Stop(); err != nil {
		logger.BillingLog.Error("Problem stopping server", "Err", err)
	}

	logger.BillingLog.Infoln("BillingDomain server stopped")
	b.wg.Done()
}
