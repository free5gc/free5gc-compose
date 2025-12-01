package webui_service

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/sirupsen/logrus"

	"github.com/free5gc/util/mongoapi"
	"github.com/free5gc/webconsole/backend/WebUI"
	"github.com/free5gc/webconsole/backend/billing"
	"github.com/free5gc/webconsole/backend/factory"
	"github.com/free5gc/webconsole/backend/logger"
	"github.com/free5gc/webconsole/backend/webui_context"
)

// Configuration constants
const (
	CorsMaxAge            = 86400
	ServerShutdownTimeout = 3 * time.Second
)

type WebuiApp struct {
	cfg      *factory.Config
	webuiCtx *webui_context.WEBUIContext

	wg            *sync.WaitGroup
	server        *http.Server
	billingServer *billing.BillingDomain
}

func NewApp(cfg *factory.Config) (*WebuiApp, error) {
	webui := &WebuiApp{
		cfg: cfg,
		wg:  &sync.WaitGroup{},
	}
	webui.SetLogEnable(cfg.GetLogEnable())
	webui.SetLogLevel(cfg.GetLogLevel())
	webui.SetReportCaller(cfg.GetLogReportCaller())

	webui_context.Init()
	webui.webuiCtx = webui_context.GetSelf()
	return webui, nil
}

func (a *WebuiApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("Log enable is set to [%v]", enable)
	if enable && logger.Log.Out == os.Stderr {
		return
	} else if !enable && logger.Log.Out == io.Discard {
		return
	}
	a.cfg.SetLogEnable(enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(io.Discard)
	}
}

func (a *WebuiApp) SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.MainLog.Warnf("Log level [%s] is invalid", level)
		return
	}
	logger.MainLog.Infof("Log level is set to [%s]", level)
	if lvl == logger.Log.GetLevel() {
		return
	}
	a.cfg.SetLogLevel(level)
	logger.Log.SetLevel(lvl)
}

func (a *WebuiApp) SetReportCaller(reportCaller bool) {
	logger.MainLog.Infof("Report Caller is set to [%v]", reportCaller)
	if reportCaller == logger.Log.ReportCaller {
		return
	}
	a.cfg.SetLogReportCaller(reportCaller)
	logger.Log.SetReportCaller(reportCaller)
}

func (a *WebuiApp) Start(tlsKeyLogPath string) {
	// get config file info from WebUIConfig
	mongodb := factory.WebuiConfig.Configuration.Mongodb
	webServer := factory.WebuiConfig.Configuration.WebServer
	billingServer := factory.WebuiConfig.Configuration.BillingServer

	// Connect to MongoDB
	if err := mongoapi.SetMongoDB(mongodb.Name, mongodb.Url); err != nil {
		logger.InitLog.Errorf("Server start err: %+v", err)
		return
	}

	logger.InitLog.Infoln("Server started")

	a.wg.Add(1)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				// Print stack for panic to log. Fatalf() will let program exit.
				logger.InitLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
			}
		}()

		<-signalChannel
		a.Terminate()
		a.wg.Done()
	}()

	go func() {
		err := webui_context.SendNFRegistration()
		if err != nil {
			retry_err := webui_context.RetrySendNFRegistration(1)
			if retry_err != nil {
				logger.InitLog.Errorln(retry_err)
				logger.InitLog.Warningln("The registration to NRF failed, resulting in limited functionalities.")
			}
		} else {
			a.webuiCtx.IsRegistered = true
		}
	}()

	router := WebUI.NewRouter()
	WebUI.SetAdmin()
	if err := WebUI.InitJwtKey(); err != nil {
		logger.InitLog.Errorln(err)
		return
	}

	router.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "User-Agent",
			"Referrer", "Host", "Authorization", "Token", "X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           CorsMaxAge,
	}))

	self := webui_context.GetSelf()
	self.UpdateNfProfiles()

	if billingServer.Enable {
		a.wg.Add(1)
		a.billingServer = billing.OpenServer(a.wg)
		if a.billingServer == nil {
			logger.InitLog.Errorln("Billing Server open error.")
		}
	}

	router.NoRoute(ReturnPublic())

	var addr string
	if webServer != nil {
		addr = webServer.IP + ":" + webServer.PORT
	} else {
		addr = ":5000"
	}

	a.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		logger.MainLog.Infof("Http server listening on %+v", addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.MainLog.Fatalf("listen: %s\n", err)
		}
	}()

	logger.MainLog.Infoln("wait all routine stopped")
	a.wg.Wait()
}

func (a *WebuiApp) Terminate() {
	logger.MainLog.Infoln("Terminating WebUI-AF...")

	if a.billingServer != nil {
		a.billingServer.Stop()
	}

	if a.server != nil {
		logger.MainLog.Infoln("stopping HTTP server")

		ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			logger.MainLog.Fatal("HTTP server forced to shutdown: ", err)
		}
	}

	// Deregister with NRF
	if a.webuiCtx.IsRegistered {
		problemDetails, err := webui_context.SendDeregisterNFInstance()
		if problemDetails != nil {
			logger.InitLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
		} else if err != nil {
			logger.InitLog.Errorf("Deregister NF instance Error[%+v]", err)
		} else {
			logger.InitLog.Infof("Deregister from NRF successfully")
		}
	}

	logger.MainLog.Infoln("WebUI-AF Terminated")
}
