/*
 * UDR Configuration Factory
 */

package factory

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asaskevich/govalidator"

	"github.com/free5gc/udr/internal/logger"
)

const (
	UdrDefaultTLSKeyLogPath    = "./log/udrsslkey.log"
	UdrDefaultCertPemPath      = "./cert/udr.pem"
	UdrDefaultPrivateKeyPath   = "./cert/udr.key"
	UdrDefaultConfigPath       = "./config/udrcfg.yaml"
	UdrSbiDefaultIPv4          = "127.0.0.9"
	UdrSbiDefaultPort          = 8000
	UdrSbiDefaultScheme        = "https"
	UdrMetricsDefaultEnabled   = false
	UdrMetricsDefaultPort      = 9091
	UdrMetricsDefaultScheme    = "https"
	UdrMetricsDefaultNamespace = "free5gc"
	UdrDefaultNrfUri           = "https://127.0.0.10:8000"
	UdrDrResUriPrefix          = "/nudr-dr/v2"
	UdrGroupIdResUriPrefix     = "/nudr-group-id-map/v1"
	HSSIsmSDMUriPrefix         = "/nhss-ims-sdm/v1"
)

type DbType string

type Config struct {
	Info          *Info          `yaml:"info" valid:"required"`
	Configuration *Configuration `yaml:"configuration" valid:"required"`
	Logger        *Logger        `yaml:"logger" valid:"required"`
	sync.RWMutex
}

func (c *Config) Validate() (bool, error) {
	if configuration := c.Configuration; configuration != nil {
		if result, err := configuration.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty" valid:"required,in(1.1.0)"`
	Description string `yaml:"description,omitempty" valid:"type(string),optional"`
}

const (
	UDR_DEFAULT_IPV4     = "127.0.0.4"
	UDR_DEFAULT_PORT     = "8000"
	UDR_DEFAULT_PORT_INT = 8000
)

type Configuration struct {
	Sbi             *Sbi     `yaml:"sbi" valid:"required"`
	Metrics         *Metrics `yaml:"metrics,omitempty" valid:"optional"`
	DbConnectorType DbType   `yaml:"dbConnectorType" valid:"required,in(mongodb)"`
	Mongodb         *Mongodb `yaml:"mongodb" valid:"optional"`
	NrfUri          string   `yaml:"nrfUri" valid:"url,required"`
	NrfCertPem      string   `yaml:"nrfCertPem,omitempty" valid:"optional"`
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

func (c *Configuration) validate() (bool, error) {
	govalidator.TagMap["scheme"] = govalidator.Validator(func(str string) bool {
		return str == "https" || str == "http"
	})

	if c.Metrics != nil {
		if _, err := c.Metrics.validate(); err != nil {
			return false, err
		}

		if c.Sbi != nil && c.Metrics.Port == c.Sbi.Port && c.Sbi.BindingIPv4 == c.Metrics.BindingIPv4 {
			var errs govalidator.Errors
			err := fmt.Errorf("sbi and metrics bindings IPv4: %s and port: %d cannot be the same, "+
				"please provide at least another port for the metrics", c.Sbi.BindingIPv4, c.Sbi.Port)
			errs = append(errs, err)
			return false, error(errs)
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"scheme,required"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,optional"` // IP that is registered at NRF.
	// IPv6Addr string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"host,optional"` // IP used to run the server in the node.
	Port        int    `yaml:"port" valid:"port,required"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
}

type Tls struct {
	Pem string `yaml:"pem,omitempty" valid:"type(string),minstringlength(1),required"`
	Key string `yaml:"key,omitempty" valid:"type(string),minstringlength(1),required"`
}

func (t *Tls) validate() (bool, error) {
	result, err := govalidator.ValidateStruct(t)
	return result, appendInvalid(err)
}

type Metrics struct {
	Enable      bool   `yaml:"enable" valid:"optional"`
	Scheme      string `yaml:"scheme" valid:"required,scheme"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"required,host"` // IP used to run the server in the node.
	Port        int    `yaml:"port,omitempty" valid:"optional,port"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
	Namespace   string `yaml:"namespace" valid:"optional"`
}

// This function is the mirror of the SBI one, I decided not to factor the code as it could in the future diverge.
// And it will reduce the cognitive overload when reading the function by not hiding the logic elsewhere.
func (m *Metrics) validate() (bool, error) {
	var errs govalidator.Errors

	if tls := m.Tls; tls != nil {
		if _, err := tls.validate(); err != nil {
			errs = append(errs, err)
		}
	}

	if _, err := govalidator.ValidateStruct(m); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return false, error(errs)
	}
	return true, nil
}

type Mongodb struct {
	Name string `yaml:"name" valid:"type(string),required"`
	Url  string `yaml:"url" valid:"requrl,required"`
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

func (c *Config) GetVersion() string {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if c.Info.Version != "" {
		return c.Info.Version
	}
	return ""
}

func (c *Config) SetLogEnable(enable bool) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

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
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

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
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

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
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.Enable
}

func (c *Config) GetLogLevel() string {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return "info"
	}
	return c.Logger.Level
}

func (c *Config) GetLogReportCaller() bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.ReportCaller
}

func (c *Config) GetCertPemPath() string {
	c.RLock()
	defer c.RUnlock()
	return c.Configuration.Sbi.Tls.Pem
}

func (c *Config) GetCertKeyPath() string {
	c.RLock()
	defer c.RUnlock()
	return c.Configuration.Sbi.Tls.Key
}

func (c *Config) AreMetricsEnabled() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil {
		return c.Configuration.Metrics.Enable
	}
	return UdrMetricsDefaultEnabled
}

func (c *Config) GetMetricsScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Scheme != "" {
		return c.Configuration.Metrics.Scheme
	}
	return UdrMetricsDefaultScheme
}

func (c *Config) GetMetricsPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Port != 0 {
		return c.Configuration.Metrics.Port
	}
	return UdrMetricsDefaultPort
}

func (c *Config) GetMetricsBindingIP() string {
	c.RLock()
	defer c.RUnlock()
	bindIP := "0.0.0.0"

	if c.Configuration == nil || c.Configuration.Metrics == nil {
		return bindIP
	}

	if c.Configuration.Metrics.BindingIPv4 != "" {
		if bindIP = os.Getenv(c.Configuration.Metrics.BindingIPv4); bindIP != "" {
			logger.CfgLog.Infof("Parsing ServerIPv4 [%s] from ENV Variable", bindIP)
		} else {
			bindIP = c.Configuration.Metrics.BindingIPv4
		}
	}
	return bindIP
}

func (c *Config) GetMetricsBindingAddr() string {
	c.RLock()
	defer c.RUnlock()
	return c.GetMetricsBindingIP() + ":" + strconv.Itoa(c.GetMetricsPort())
}

func (c *Config) GetMetricsCertPemPath() string {
	// We can see if there is a benefit to factor this tls key/pem with the sbi ones
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Metrics != nil && c.Configuration.Metrics.Tls != nil {
		return c.Configuration.Metrics.Tls.Pem
	}
	return ""
}

func (c *Config) GetMetricsCertKeyPath() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Metrics != nil && c.Configuration.Metrics.Tls != nil {
		return c.Configuration.Metrics.Tls.Key
	}
	return ""
}

func (c *Config) GetMetricsNamespace() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration.Metrics != nil && c.Configuration.Metrics.Namespace != "" {
		return c.Configuration.Metrics.Namespace
	}
	return UdrMetricsDefaultNamespace
}
