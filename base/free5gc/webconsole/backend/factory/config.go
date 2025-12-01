/*

 * WebUi Configuration Factory

 */

package factory

import (
	"fmt"
	"sync"

	"github.com/asaskevich/govalidator"

	"github.com/free5gc/webconsole/backend/logger"
)

const (
	WebuiDefaultTLSKeyLogPath  = "./log/webuisslkey.log"
	WebuiDefaultCertPemPath    = "./cert/webui.pem"
	WebuiDefaultPrivateKeyPath = "./cert/webui.key"
	WebuiDefaultConfigPath     = "./config/webuicfg.yaml"
)

type Config struct {
	Info          *Info          `yaml:"info" valid:"required"`
	Configuration *Configuration `yaml:"configuration" valid:"required"`
	Logger        *Logger        `yaml:"logger" valid:"required"`
	sync.RWMutex
}

func (c *Config) Validate() (bool, error) {
	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty" valid:"required,in(1.0.3)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	WebServer     *WebServer     `yaml:"webServer,omitempty" valid:"optional"`
	Mongodb       *Mongodb       `yaml:"mongodb" valid:"required"`
	NrfUri        string         `yaml:"nrfUri" valid:"required"`
	BillingServer *BillingServer `yaml:"billingServer,omitempty" valid:"required"`
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

type WebServer struct {
	Scheme string `yaml:"scheme" valid:"required"`
	IP     string `yaml:"ipv4Address,omitempty"`
	PORT   string `yaml:"port" valid:"required"`
}

type Cert struct {
	Pem string `yaml:"pem,omitempty" valid:"type(string),minstringlength(1),required"`
	Key string `yaml:"key,omitempty" valid:"type(string),minstringlength(1),required"`
}

type PortRange struct {
	Start int `yaml:"start,omitempty" valid:"required" json:"start"`
	End   int `yaml:"end,omitempty" valid:"required" json:"end"`
}

type BillingServer struct {
	Enable     bool      `yaml:"enable,omitempty" valid:"required,type(bool)"`
	HostIPv4   string    `yaml:"hostIPv4,omitempty" valid:"required,host"`
	ListenPort int       `yaml:"listenPort,omitempty" valid:"required,port"`
	PortRange  PortRange `yaml:"portRange,omitempty" valid:"required"`
	BastPath   string    `yaml:"basePath,omitempty" valid:"type(string),required"`
	Cert       *Cert     `yaml:"cert,omitempty" valid:"optional"`
	Port       int       `yaml:"port,omitempty" valid:"optional,port"`
}

type Mongodb struct {
	Name string `yaml:"name" valid:"required"`
	Url  string `yaml:"url"  valid:"required"`
}

func appendInvalid(err error) error {
	var errs govalidator.Errors

	if err == nil {
		return nil
	}

	es := err.(govalidator.Errors).Errors()
	for _, e := range es {
		errs = append(errs, fmt.Errorf("invalid %w", e))
	}

	return error(errs)
}

func (c *Config) SetLogEnable(enable bool) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Enable: enable,
			Level:  "info",
		}
	} else {
		c.Logger.Enable = enable
	}
}

func (c *Config) SetLogLevel(level string) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level: level,
		}
	} else {
		c.Logger.Level = level
	}
}

func (c *Config) SetLogReportCaller(reportCaller bool) {
	c.Lock()
	defer c.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level:        "info",
			ReportCaller: reportCaller,
		}
	} else {
		c.Logger.ReportCaller = reportCaller
	}
}

func (c *Config) GetLogEnable() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.Enable
}

func (c *Config) GetLogLevel() string {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return "info"
	}
	return c.Logger.Level
}

func (c *Config) GetLogReportCaller() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.ReportCaller
}
