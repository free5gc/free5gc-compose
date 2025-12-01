/*
 * UDM Configuration Factory
 */

package factory

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asaskevich/govalidator"

	"github.com/free5gc/udm/internal/logger"
	"github.com/free5gc/udm/pkg/suci"
)

const (
	UdmDefaultTLSKeyLogPath       = "./log/udmsslkey.log"
	UdmDefaultCertPemPath         = "./cert/udm.pem"
	UdmDefaultPrivateKeyPath      = "./cert/udm.key"
	UdmDefaultConfigPath          = "./config/udmcfg.yaml"
	UdmSbiDefaultIPv4             = "127.0.0.3"
	UdmSbiDefaultPort             = 8000
	UdmSbiDefaultScheme           = "https"
	UdmMetricsDefaultEnabled      = false
	UdmMetricsDefaultPort         = 9091
	UdmMetricsDefaultScheme       = "https"
	UdmMetricsDefaultNamespace    = "free5gc"
	UdmDefaultNrfUri              = "https://127.0.0.10:8000"
	UdmSorprotectionResUriPrefix  = "/nudm-sorprotection/v1"
	UdmAuthResUriPrefix           = "/nudm-auth/v1"
	UdmfUpuprotectionResUriPrefix = "/nudm-upuprotection/v1"
	UdmEcmResUriPrefix            = "/nudm-ecm/v1"
	UdmSdmResUriPrefix            = "/nudm-sdm/v2"
	UdmEeResUriPrefix             = "/nudm-ee/v1"
	UdmDrResUriPrefix             = "/nudr-dr/v1"
	UdmUecmResUriPrefix           = "/nudm-uecm/v1"
	UdmPpResUriPrefix             = "/nudm-pp/v1"
	UdmUeauResUriPrefix           = "/nudm-ueau/v1"
	UdmMtResUrdPrefix             = "/nudm-mt/v1"
	UdmNiddauResUriPrefix         = "/nudm-niddau/v1"
	UdmRsdsResUriPrefix           = "/nudm-rsds/v1"
	UdmSsauResUriPrefix           = "/nudm-ssau/v1"
	UdmUeidResUriPrefix           = "/nudm-ueid/v1"
)

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
	Version     string `yaml:"version,omitempty" valid:"required,in(1.0.3)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	Sbi             *Sbi               `yaml:"sbi,omitempty"  valid:"required"`
	Metrics         *Metrics           `yaml:"metrics,omitempty" valid:"optional"`
	ServiceNameList []string           `yaml:"serviceNameList,omitempty"  valid:"required"`
	NrfUri          string             `yaml:"nrfUri,omitempty"  valid:"required, url"`
	NrfCertPem      string             `yaml:"nrfCertPem,omitempty" valid:"optional"`
	SuciProfiles    []suci.SuciProfile `yaml:"SuciProfile,omitempty"`
}
type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

func (c *Configuration) validate() (bool, error) {
	govalidator.TagMap["scheme"] = func(str string) bool {
		return str == "https" || str == "http"
	}

	if sbi := c.Sbi; sbi != nil {
		if result, err := sbi.validate(); err != nil {
			return result, err
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

	if c.ServiceNameList != nil {
		var errs govalidator.Errors
		for _, v := range c.ServiceNameList {
			if v != "nudm-sdm" && v != "nudm-uecm" && v != "nudm-ueau" && v != "nudm-ee" && v != "nudm-pp" {
				err := fmt.Errorf("invalid ServiceNameList: [%s],"+
					" value should be nudm-sdm or nudm-uecm or nudm-ueau or nudm-ee or nudm-pp", v)
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return false, error(errs)
		}
	}

	if c.SuciProfiles != nil {
		var errs govalidator.Errors
		for _, s := range c.SuciProfiles {
			protectScheme := s.ProtectionScheme
			if result := govalidator.StringMatches(protectScheme, "^[A-F0-9]{1}$"); !result {
				err := fmt.Errorf("invalid ProtectionScheme: %s, should be a single hexadecimal digit", protectScheme)
				errs = append(errs, err)
			}

			privateKey := s.PrivateKey
			if result := govalidator.StringMatches(privateKey, "^[A-Fa-f0-9]{64}$"); !result {
				err := fmt.Errorf("invalid PrivateKey: %s, should be 64 hexadecimal digits", privateKey)
				errs = append(errs, err)
			}

			publicKey := s.PublicKey
			if result := govalidator.StringMatches(publicKey, "^[A-Fa-f0-9]{64,130}$"); !result {
				err := fmt.Errorf("invalid PublicKey: %s, should be 64(profile A), 66(profile B, compressed),"+
					"or 130(profile B, uncompressed) hexadecimal digits", publicKey)
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return false, error(errs)
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, err
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

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,required"` // IP that is registered at NRF.
	// IPv6Addr string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"host,required"` // IP used to run the server in the node.
	Port        int    `yaml:"port,omitempty" valid:"port,required"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
}

func (s *Sbi) validate() (bool, error) {
	if tls := s.Tls; tls != nil {
		if result, err := tls.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(s)
	return result, err
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

func (c *Config) GetSbiBindingAddr() string {
	c.RLock()
	defer c.RUnlock()
	return c.GetSbiBindingIP() + ":" + strconv.Itoa(c.GetSbiPort())
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
	return UdmSbiDefaultPort
}

func (c *Config) GetSbiScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Scheme != "" {
		return c.Configuration.Sbi.Scheme
	}
	return UdmSbiDefaultScheme
}

func (c *Config) AreMetricsEnabled() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil {
		return c.Configuration.Metrics.Enable
	}
	return UdmMetricsDefaultEnabled
}

func (c *Config) GetMetricsScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Scheme != "" {
		return c.Configuration.Metrics.Scheme
	}
	return UdmMetricsDefaultScheme
}

func (c *Config) GetMetricsPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Port != 0 {
		return c.Configuration.Metrics.Port
	}
	return UdmMetricsDefaultPort
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
	return UdmMetricsDefaultNamespace
}
