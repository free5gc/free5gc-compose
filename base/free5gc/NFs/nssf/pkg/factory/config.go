/*
 * NSSF Configuration Factory
 */

package factory

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asaskevich/govalidator"

	"github.com/free5gc/nssf/internal/logger"
	"github.com/free5gc/openapi/models"
)

const (
	NssfDefaultTLSKeyLogPath    = "./log/nssfsslkey.log"
	NssfDefaultCertPemPath      = "./cert/nssf.pem"
	NssfDefaultPrivateKeyPath   = "./cert/nssf.key"
	NssfDefaultConfigPath       = "./config/nssfcfg.yaml"
	NssfSbiDefaultIPv4          = "127.0.0.31"
	NssfSbiDefaultPort          = 8000
	NssfSbiDefaultScheme        = "https"
	NssfDefaultNrfUri           = "https://127.0.0.10:8000"
	NssfMetricsDefaultEnabled   = false
	NssfMetricsDefaultPort      = 9091
	NssfMetricsDefaultScheme    = "https"
	NssfMetricsDefaultNamespace = "free5gc"
	NssfNssaiavailResUriPrefix  = "/nnssf-nssaiavailability/v1"
	NssfNsselectResUriPrefix    = "/nnssf-nsselection/v2"
)

type Config struct {
	Info          *Info          `yaml:"info" valid:"required"`
	Configuration *Configuration `yaml:"configuration" valid:"required"`
	Subscriptions []Subscription `yaml:"subscriptions,omitempty"`
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
	Version     string `yaml:"version" valid:"required,in(1.0.2)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	NssfName                 string                  `yaml:"nssfName,omitempty"`
	Sbi                      *Sbi                    `yaml:"sbi"`
	Metrics                  *Metrics                `yaml:"metrics,omitempty" valid:"optional"`
	ServiceNameList          []models.ServiceName    `yaml:"serviceNameList"`
	NrfUri                   string                  `yaml:"nrfUri"`
	NrfCertPem               string                  `yaml:"nrfCertPem,omitempty" valid:"optional"`
	SupportedPlmnList        []models.PlmnId         `yaml:"supportedPlmnList,omitempty"`
	SupportedNssaiInPlmnList []SupportedNssaiInPlmn  `yaml:"supportedNssaiInPlmnList"`
	NsiList                  []NsiConfig             `yaml:"nsiList,omitempty"`
	AmfSetList               []AmfSetConfig          `yaml:"amfSetList"`
	AmfList                  []AmfConfig             `yaml:"amfList"`
	TaList                   []TaConfig              `yaml:"taList"`
	MappingListFromPlmn      []MappingFromPlmnConfig `yaml:"mappingListFromPlmn"`
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

func (c *Configuration) validate() (bool, error) {
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

	for index, serviceName := range c.ServiceNameList {
		switch serviceName {
		case "nnssf-nsselection":
		case "nnssf-nssaiavailability":
		default:
			err := errors.New("Invalid serviceNameList[" + strconv.Itoa(index) + "]: " +
				string(serviceName) + ", should be nausf-auth.")
			return false, err
		}
	}

	for index, plmnId := range c.SupportedPlmnList {
		if result := govalidator.StringMatches(plmnId.Mcc, "^[0-9]{3}$"); !result {
			err := errors.New("Invalid plmnSupportList[" + strconv.Itoa(index) + "].Mcc: " +
				plmnId.Mcc + ", should be 3 digits interger.")
			return false, err
		}

		if result := govalidator.StringMatches(plmnId.Mnc, "^[0-9]{2,3}$"); !result {
			err := errors.New("Invalid plmnSupportList[" + strconv.Itoa(index) + "].Mnc: " +
				plmnId.Mnc + ", should be 2 or 3 digits interger.")
			return false, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Sbi struct {
	Scheme models.UriScheme `yaml:"scheme"`
	// Currently only support IPv4 and thus `Ipv4Addr` field shall not be empty
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,required"` // IP that is registered at NRF.
	// IPv6Addr string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"host,required"` // IP used to run the server in the node.
	Port        int    `yaml:"port"`
	Tls         *Tls   `yaml:"tls,omitempty" valid:"optional"`
}

func (s *Sbi) validate() (bool, error) {
	if tls := s.Tls; tls != nil {
		if result, err := tls.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(s)
	return result, appendInvalid(err)
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

type AmfConfig struct {
	NfId                           string                                  `yaml:"nfId"`
	SupportedNssaiAvailabilityData []models.SupportedNssaiAvailabilityData `yaml:"supportedNssaiAvailabilityData"`
}

type TaConfig struct {
	Tai                  *models.Tai               `yaml:"tai"`
	AccessType           *models.AccessType        `yaml:"accessType"`
	SupportedSnssaiList  []models.ExtSnssai        `yaml:"supportedSnssaiList"`
	RestrictedSnssaiList []models.RestrictedSnssai `yaml:"restrictedSnssaiList,omitempty"`
}

type SupportedNssaiInPlmn struct {
	PlmnId              *models.PlmnId  `yaml:"plmnId"`
	SupportedSnssaiList []models.Snssai `yaml:"supportedSnssaiList"`
}

type NsiConfig struct {
	Snssai             *models.Snssai          `yaml:"snssai"`
	NsiInformationList []models.NsiInformation `yaml:"nsiInformationList"`
}

type AmfSetConfig struct {
	AmfSetId                       string                                  `yaml:"amfSetId"`
	AmfList                        []string                                `yaml:"amfList,omitempty"`
	NrfAmfSet                      string                                  `yaml:"nrfAmfSet,omitempty"`
	SupportedNssaiAvailabilityData []models.SupportedNssaiAvailabilityData `yaml:"supportedNssaiAvailabilityData"`
}

type MappingFromPlmnConfig struct {
	OperatorName    string                   `yaml:"operatorName,omitempty"`
	HomePlmnId      *models.PlmnId           `yaml:"homePlmnId"`
	MappingOfSnssai []models.MappingOfSnssai `yaml:"mappingOfSnssai"`
}

type Subscription struct {
	SubscriptionId   string                                  `yaml:"subscriptionId"`
	SubscriptionData *models.NssfEventSubscriptionCreateData `yaml:"subscriptionData"`
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
	return NssfMetricsDefaultEnabled
}

func (c *Config) GetMetricsScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Scheme != "" {
		return c.Configuration.Metrics.Scheme
	}
	return NssfMetricsDefaultScheme
}

func (c *Config) GetMetricsPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Port != 0 {
		return c.Configuration.Metrics.Port
	}
	return NssfMetricsDefaultPort
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
	return NssfMetricsDefaultNamespace
}
