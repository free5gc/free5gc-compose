/*
 * PCF Configuration Factory
 */

package factory

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asaskevich/govalidator"

	"github.com/free5gc/pcf/internal/logger"
)

const (
	PcfTimeFormatLayout         = "2006-01-02 15:04:05"
	PcfDefaultTLSKeyLogPath     = "./log/pcfsslkey.log"
	PcfDefaultCertPemPath       = "./cert/pcf.pem"
	PcfDefaultPrivateKeyPath    = "./cert/pcf.key"
	PcfDefaultConfigPath        = "./config/pcfcfg.yaml"
	PcfSbiDefaultIPv4           = "127.0.0.7"
	PcfSbiDefaultPort           = 8000
	PcfSbiDefaultScheme         = "https"
	PcfMetricsDefaultEnabled    = false
	PcfMetricsDefaultPort       = 9091
	PcfMetricsDefaultScheme     = "https"
	PcfMetricsDefaultNamespace  = "free5gc"
	PcfDefaultNrfUri            = "https://127.0.0.10:8000"
	PcfPolicyAuthResUriPrefix   = "/npcf-policyauthorization/v1"
	PcfAMpolicyCtlResUriPrefix  = "/npcf-am-policy-control/v1"
	PcfCallbackResUriPrefix     = "/npcf-callback/v1"
	PcfSMpolicyCtlResUriPrefix  = "/npcf-smpolicycontrol/v1"
	PcfBdtPolicyCtlResUriPrefix = "/npcf-bdtpolicycontrol/v1"
	PcfOamResUriPrefix          = "/npcf-oam/v1"
	PcfUePolicyCtlResUriPrefix  = "/npcf-ue-policy-control/v1/"
)

type Config struct {
	Info          *Info          `yaml:"info" valid:"required"`
	Configuration *Configuration `yaml:"configuration" valid:"required"`
	Logger        *Logger        `yaml:"logger" valid:"required"`
	sync.RWMutex
}

func (c *Config) Validate() (bool, error) {
	govalidator.TagMap["scheme"] = func(str string) bool {
		return str == "https" || str == "http"
	}

	if configuration := c.Configuration; configuration != nil {
		if result, err := configuration.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty" valid:"required,in(1.0.2)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	PcfName         string    `yaml:"pcfName,omitempty" valid:"required, type(string)"`
	Sbi             *Sbi      `yaml:"sbi,omitempty" valid:"required"`
	Metrics         *Metrics  `yaml:"metrics,omitempty" valid:"optional"`
	TimeFormat      string    `yaml:"timeFormat,omitempty" valid:"required"`
	DefaultBdtRefId string    `yaml:"defaultBdtRefId,omitempty" valid:"required, type(string)"`
	NrfUri          string    `yaml:"nrfUri,omitempty" valid:"required, url"`
	NrfCertPem      string    `yaml:"nrfCertPem,omitempty" valid:"optional"`
	Bsf             *Bsf      `yaml:"bsf,omitempty" valid:"optional"`
	ServiceList     []Service `yaml:"serviceList,omitempty" valid:"required"`
	Mongodb         *Mongodb  `yaml:"mongodb" valid:"required"`
	Locality        string    `yaml:"locality,omitempty" valid:"-"`
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

type Bsf struct {
	Enable bool `yaml:"enable" valid:"type(bool)"`
}

func (c *Configuration) validate() (bool, error) {
	if c.Sbi != nil {
		if _, err := c.Sbi.validate(); err != nil {
			return false, err
		}
	}

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

	if result := govalidator.IsTime(c.TimeFormat, PcfTimeFormatLayout); !result {
		err := fmt.Errorf("invalid TimeFormat: %s, should be in 2019-01-02 15:04:05 format", c.TimeFormat)
		return false, err
	}

	if c.ServiceList != nil {
		var errs govalidator.Errors
		for _, v := range c.ServiceList {
			if _, err := v.validate(); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return false, error(errs)
		}
	}

	if c.Mongodb != nil {
		if _, err := c.Mongodb.validate(); err != nil {
			return false, err
		}
	}

	if _, err := govalidator.ValidateStruct(c); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
}

type Service struct {
	ServiceName string `yaml:"serviceName" valid:"required, service"`
	SuppFeat    string `yaml:"suppFeat,omitempty" valid:"-"`
}

func (s *Service) validate() (bool, error) {
	govalidator.TagMap["service"] = govalidator.Validator(func(str string) bool {
		switch str {
		case "npcf-am-policy-control":
		case "npcf-smpolicycontrol":
		case "npcf-bdtpolicycontrol":
		case "npcf-policyauthorization":
		case "npcf-eventexposure":
		case "npcf-ue-policy-control":
		default:
			return false
		}
		return true
	})

	switch s.ServiceName {
	case "npcf-smpolicycontrol":
		if sf, e := strconv.ParseUint(s.SuppFeat, 16, 40); e != nil {
			err := fmt.Errorf("invalid SuppFeat: %s, range of the value should be 0~3fff", s.SuppFeat)
			return false, err
		} else {
			if sf2, e := strconv.ParseUint("3fff", 16, 20); e == nil {
				if sf > sf2 {
					err := fmt.Errorf("invalid SuppFeat: %s, range of the value should be 0~3fff", s.SuppFeat)
					return false, err
				}
			}
		}
	case "npcf-policyauthorization":
		if s.SuppFeat != "0" && s.SuppFeat != "1" && s.SuppFeat != "2" && s.SuppFeat != "3" {
			err := fmt.Errorf("invalid SuppFeat: %s, range of the value should be 0~3", s.SuppFeat)
			return false, err
		}
	}

	if _, err := govalidator.ValidateStruct(s); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
}

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"required,scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"required,host"` // IP that is registered at NRF.
	// IPv6Addr  string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"required,host"` // IP used to run the server in the node.
	Port        int    `yaml:"port,omitempty" valid:"required,port"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
}

func (s *Sbi) validate() (bool, error) {
	if tls := s.Tls; tls != nil {
		if result, err := tls.validate(); err != nil {
			return result, err
		}
	}

	if _, err := govalidator.ValidateStruct(s); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
}

type Tls struct {
	Pem string `yaml:"pem,omitempty" valid:"type(string),minstringlength(1),required"`
	Key string `yaml:"key,omitempty" valid:"type(string),minstringlength(1),required"`
}

func (t *Tls) validate() (bool, error) {
	result, err := govalidator.ValidateStruct(t)
	return result, err
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
	Name string `yaml:"name" valid:"required, type(string)"`
	Url  string `yaml:"url" valid:"required"`
}

func (m *Mongodb) validate() (bool, error) {
	pattern := `[-a-zA-Z0-9@:%._\+~#=]{1,256}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`
	if result := govalidator.StringMatches(m.Url, pattern); !result {
		err := fmt.Errorf("invalid Url: %s", m.Url)
		return false, err
	}
	if _, err := govalidator.ValidateStruct(m); err != nil {
		return false, appendInvalid(err)
	}

	return true, nil
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

func (c *Config) GetSbiBindingIP() string {
	c.RLock()
	defer c.RUnlock()
	bindIP := "0.0.0.0"
	if c.Configuration == nil || c.Configuration.Sbi == nil {
		return bindIP
	}
	if c.Configuration.Sbi.BindingIPv4 != "" {
		if bindIP = os.Getenv(c.Configuration.Sbi.BindingIPv4); bindIP != "" {
			logger.CfgLog.Infof("Parsing ServerIPv4 [%s] from ENV Variable", bindIP)
		} else {
			bindIP = c.Configuration.Sbi.BindingIPv4
		}
	}
	return bindIP
}

func (c *Config) GetSbiPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Port != 0 {
		return c.Configuration.Sbi.Port
	}
	return PcfSbiDefaultPort
}

func (c *Config) GetSbiBindingAddr() string {
	c.RLock()
	defer c.RUnlock()
	return c.GetSbiBindingIP() + ":" + strconv.Itoa(c.GetSbiPort())
}

func (c *Config) GetSbiScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Scheme != "" {
		return c.Configuration.Sbi.Scheme
	}
	return PcfSbiDefaultScheme
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

func (c *Config) GetVersion() string {
	c.RLock()
	defer c.RUnlock()

	if c.Info.Version != "" {
		return c.Info.Version
	}
	return ""
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

func (c *Config) AreMetricsEnabled() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil {
		return c.Configuration.Metrics.Enable
	}
	return PcfMetricsDefaultEnabled
}

func (c *Config) GetMetricsScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Scheme != "" {
		return c.Configuration.Metrics.Scheme
	}
	return PcfMetricsDefaultScheme
}

func (c *Config) GetMetricsPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Port != 0 {
		return c.Configuration.Metrics.Port
	}
	return PcfMetricsDefaultPort
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
	return PcfMetricsDefaultNamespace
}
