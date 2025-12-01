package main

import (
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/util/version"
	"github.com/free5gc/webconsole/backend/factory"
	"github.com/free5gc/webconsole/backend/logger"
	"github.com/free5gc/webconsole/backend/webui_service"
)

// File system constants
const (
	DefaultDirPermission = 0o775
)

var WEBUI *webui_service.WebuiApp

func main() {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()
	app := cli.NewApp()
	app.Name = "webui"
	app.Usage = "free5GC Web Console"
	app.Action = action
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
		},
		&cli.StringSliceFlag{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "Output NF log to `FILE`",
		},
	}
	if err := app.Run(os.Args); err != nil {
		logger.MainLog.Errorf("WEBUI Run error: %v\n", err)
	}
}

func action(cliCtx *cli.Context) error {
	tlsKeyLogPath, err := initLogFile(cliCtx.StringSlice("log"))
	if err != nil {
		return err
	}

	logger.MainLog.Infoln("WEBUI version: ", version.GetVersion())

	cfg, err := factory.ReadConfig(cliCtx.String("config"))
	if err != nil {
		return err
	}
	factory.WebuiConfig = cfg

	webui, err := webui_service.NewApp(cfg)
	if err != nil {
		return err
	}
	WEBUI = webui

	webui.Start(tlsKeyLogPath)

	return nil
}

func initLogFile(logNfPath []string) (string, error) {
	logTlsKeyPath := ""

	for _, path := range logNfPath {
		if err := logger_util.LogFileHook(logger.Log, path); err != nil {
			return "", err
		}
		if logTlsKeyPath != "" {
			continue
		}

		nfDir, _ := filepath.Split(path)
		tmpDir := filepath.Join(nfDir, "key")
		if err := os.MkdirAll(tmpDir, DefaultDirPermission); err != nil {
			logger.InitLog.Errorf("Make directory %s failed: %+v", tmpDir, err)

			return "", err
		}

		_, name := filepath.Split(factory.WebuiDefaultTLSKeyLogPath)
		logTlsKeyPath = filepath.Join(tmpDir, name)
	}

	return logTlsKeyPath, nil
}
