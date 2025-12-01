package sbi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/nef/internal/sbi/processor"
	"github.com/free5gc/nef/pkg/app"
	"github.com/free5gc/nef/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/util/metrics"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	CorsConfigMaxAge = 86400
)

type nef interface {
	app.App
	Processor() *processor.Processor
}

type Server struct {
	nef

	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(nef nef, tlsKeyLogPath string) (*Server, error) {
	s := &Server{
		nef: nef,
	}

	s.router = logger_util.NewGinWithLogrus(logger.GinLog)

	s.router.Use(metrics.InboundMetrics())

	endpoints := s.getTrafficInfluenceRoutes()
	group := s.router.Group(factory.TraffInfluResUriPrefix)
	applyRoutes(group, endpoints)

	endpoints = s.getPFDManagementRoutes()
	group = s.router.Group(factory.PfdMngResUriPrefix)
	applyRoutes(group, endpoints)

	endpoints = s.getPFDFRoutes()
	group = s.router.Group(factory.NefPfdMngResUriPrefix)
	applyRoutes(group, endpoints)

	endpoints = s.getOamRoutes()
	group = s.router.Group(factory.NefOamResUriPrefix)
	applyRoutes(group, endpoints)

	endpoints = s.getCallbackRoutes()
	group = s.router.Group(factory.NefCallbackResUriPrefix)
	applyRoutes(group, endpoints)

	s.router.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "User-Agent",
			"Referrer", "Host", "Token", "X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           CorsConfigMaxAge,
	}))

	bindAddr := s.Config().SbiBindingAddr()
	logger.SBILog.Infof("Binding addr: [%s]", bindAddr)
	var err error
	if s.httpServer, err = httpwrapper.NewHttp2Server(bindAddr, tlsKeyLogPath, s.router); err != nil {
		logger.InitLog.Errorf("Initialize HTTP server failed: %+v", err)
		return nil, err
	}

	return s, nil
}

func (s *Server) Run(wg *sync.WaitGroup) error {
	wg.Add(1)
	go s.startServer(wg)
	return nil
}

func (s *Server) Terminate() {
	const defaultShutdownTimeout time.Duration = 2 * time.Second

	if s.httpServer != nil {
		logger.SBILog.Infof("Stop SBI server (listen on %s)", s.httpServer.Addr)
		toCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(toCtx); err != nil {
			logger.SBILog.Errorf("Could not close SBI server: %#v", err)
		}
	}
}

func (s *Server) startServer(wg *sync.WaitGroup) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.SBILog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
			s.Terminate()
		}

		wg.Done()
	}()

	logger.SBILog.Infof("Start SBI server (listen on %s)", s.httpServer.Addr)

	var err error

	scheme := s.Config().SbiScheme()
	switch scheme {
	case "http":
		err = s.httpServer.ListenAndServe()
	case "https":
		// TODO: use config file to config path
		err = s.httpServer.ListenAndServeTLS(s.Config().GetCertPemPath(), s.Config().GetCertKeyPath())
	default:
		err = fmt.Errorf("scheme [%s] is not supported", scheme)
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.SBILog.Errorf("SBI server error: %+v", err)
	}
	logger.SBILog.Warnf("SBI server (listen on %s) stopped", s.httpServer.Addr)
}
