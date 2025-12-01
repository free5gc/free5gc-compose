/*
 * NEF Configuration Factory
 */

package factory

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/davecgh/go-spew/spew"
	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/openapi/models"
)

const (
	ServiceTraffInflu  string = "3gpp-traffic-influence"
	ServicePfdMng      string = "3gpp-pfd-management"
	ServiceNefPfd      string = string(models.ServiceName_NNEF_PFDMANAGEMENT)
	ServiceNefOam      string = "nnef-oam"
	ServiceNefCallback string = "nnef-callback"
)

const (
	NefDefaultTLSKeyLogPath    = "./log/nefsslkey.log"
	NefDefaultCertPemPath      = "./cert/nef.pem"
	NefDefaultPrivateKeyPath   = "./cert/nef.key"
	NefDefaultConfigPath       = "./config/nefcfg.yaml"
	NefExpectedConfigVersion   = "1.0.1"
	NefSbiDefaultIPv4          = "127.0.0.5"
	NefSbiDefaultPort          = 8000
	NefSbiDefaultScheme        = "https"
	NefMetricsDefaultEnabled   = false
	NefMetricsDefaultPort      = 9091
	NefMetricsDefaultScheme    = "https"
	NefMetricsDefaultNamespace = "free5gc"
	NefDefaultNrfUri           = "https://127.0.0.10:8000"
	TraffInfluResUriPrefix     = "/" + ServiceTraffInflu + "/v1"
	PfdMngResUriPrefix         = "/" + ServicePfdMng + "/v1"
	NefPfdMngResUriPrefix      = "/" + ServiceNefPfd + "/v1"
	NefOamResUriPrefix         = "/" + ServiceNefOam + "/v1"
	NefCallbackResUriPrefix    = "/" + ServiceNefCallback + "/v1"
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

	if info := c.Info; info != nil {
		if !govalidator.IsIn(info.Version, NefExpectedConfigVersion) {
			err := errors.New("config version should be " + NefExpectedConfigVersion)
			return false, appendInvalid(err)
		}
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
	Version     string `yaml:"version,omitempty" valid:"type(string)"`
	Description string `yaml:"description,omitempty" valid:"type(string)"`
}

type Configuration struct {
	Sbi         *Sbi `yaml:"sbi,omitempty" valid:"required"`
	Metrics     *Metrics
	NrfUri      string    `yaml:"nrfUri,omitempty" valid:"required"`
	NrfCertPem  string    `yaml:"nrfCertPem,omitempty" valid:"optional"`
	ServiceList []Service `yaml:"serviceList,omitempty" valid:"required"`
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

	for i, s := range c.ServiceList {
		switch s.ServiceName {
		case ServiceNefPfd:
		case ServiceNefOam:
		default:
			err := errors.New("invalid serviceList[" + strconv.Itoa(i) + "]: " +
				s.ServiceName + ", should be " + ServiceNefPfd + " or " + ServiceNefOam)
			return false, appendInvalid(err)
		}
	}
	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"scheme,required"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,required"` // IP that is registered at NRF.
	// IPv6Addr  string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty" valid:"host,required"` // IP used to run the server in the node.
	Port        int    `yaml:"port,omitempty" valid:"port,optional"`
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

type Service struct {
	ServiceName string `yaml:"serviceName"`
	SuppFeat    string `yaml:"suppFeat,omitempty"`
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

func (c *Config) AreMetricsEnabled() bool {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil {
		return c.Configuration.Metrics.Enable
	}
	return NefMetricsDefaultEnabled
}

func (c *Config) GetMetricsScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Scheme != "" {
		return c.Configuration.Metrics.Scheme
	}
	return NefMetricsDefaultScheme
}

func (c *Config) GetMetricsPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Metrics != nil && c.Configuration.Metrics.Port != 0 {
		return c.Configuration.Metrics.Port
	}
	return NefMetricsDefaultPort
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
	return NefMetricsDefaultNamespace
}

func appendInvalid(err error) error {
	var errs govalidator.Errors
	if err == nil {
		return nil
	}
	es, ok := err.(govalidator.Errors)
	if ok {
		for _, e := range es.Errors() {
			errs = append(errs, fmt.Errorf("invalid %w", e))
		}
	} else {
		errs = append(errs, err)
	}
	return error(errs)
}

func (c *Config) Print() {
	c.RLock()
	defer c.RUnlock()

	spew.Config.Indent = "\t"
	str := spew.Sdump(c.Configuration)
	logger.CfgLog.Infof("==================================================")
	logger.CfgLog.Infof("%s", str)
	logger.CfgLog.Infof("==================================================")
}

func (c *Config) Version() string {
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

func (c *Config) SbiScheme() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Sbi.Scheme != "" {
		return c.Configuration.Sbi.Scheme
	}
	return NefSbiDefaultScheme
}

func (c *Config) SbiPort() int {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Sbi.Port != 0 {
		return c.Configuration.Sbi.Port
	}
	return NefSbiDefaultPort
}

func (c *Config) SbiBindingIP() string {
	c.RLock()
	defer c.RUnlock()

	bindIP := "0.0.0.0"
	if c.Configuration.Sbi.BindingIPv4 != "" {
		if bindIP = os.Getenv(c.Configuration.Sbi.BindingIPv4); bindIP != "" {
			logger.CfgLog.Infof("Parsing ServerIPv4 [%s] from ENV Variable", bindIP)
		} else {
			bindIP = c.Configuration.Sbi.BindingIPv4
		}
	}
	return bindIP
}

func (c *Config) SbiBindingAddr() string {
	return c.SbiBindingIP() + ":" + strconv.Itoa(c.SbiPort())
}

func (c *Config) SbiRegisterIP() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Sbi.RegisterIPv4 != "" {
		return c.Configuration.Sbi.RegisterIPv4
	}
	return NefSbiDefaultIPv4
}

func (c *Config) SbiRegisterAddr() string {
	return c.SbiRegisterIP() + ":" + strconv.Itoa(c.SbiPort())
}

func (c *Config) SbiUri() string {
	return c.SbiScheme() + "://" + c.SbiRegisterAddr()
}

func (c *Config) NrfUri() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.NrfUri != "" {
		return c.Configuration.NrfUri
	}
	return NefDefaultNrfUri
}

func (c *Config) NrfCertPem() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.NrfCertPem != "" {
		return c.Configuration.NrfCertPem
	}
	return "" // havn't setup in config
}

func (c *Config) ServiceList() []Service {
	c.RLock()
	defer c.RUnlock()

	if len(c.Configuration.ServiceList) > 0 {
		return c.Configuration.ServiceList
	}
	return nil
}

func (c *Config) GetCertPemPath() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Sbi.Tls != nil {
		return c.Configuration.Sbi.Tls.Pem
	}
	return NefDefaultCertPemPath
}

func (c *Config) GetCertKeyPath() string {
	c.RLock()
	defer c.RUnlock()

	if c.Configuration.Sbi.Tls != nil {
		return c.Configuration.Sbi.Tls.Key
	}
	return NefDefaultPrivateKeyPath
}

func (c *Config) NFServices() []models.NrfNfManagementNfService {
	versions := strings.Split(c.Version(), ".")
	majorVersionUri := "v" + versions[0]
	var nfServices []models.NrfNfManagementNfService
	for i, s := range c.ServiceList() {
		nfService := models.NrfNfManagementNfService{
			ServiceInstanceId: strconv.Itoa(i),
			ServiceName:       models.ServiceName(s.ServiceName),
			Versions: []models.NfServiceVersion{
				{
					ApiFullVersion:  c.Version(),
					ApiVersionInUri: majorVersionUri,
				},
			},
			Scheme:          models.UriScheme(c.SbiScheme()),
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       c.SbiUri(),
			IpEndPoints: []models.IpEndPoint{
				{
					Ipv4Address: c.SbiRegisterIP(),
					Transport:   models.NrfNfManagementTransportProtocol_TCP,
					Port:        int32(c.SbiPort()),
				},
			},
			SupportedFeatures: s.SuppFeat,
		}
		nfServices = append(nfServices, nfService)
	}
	return nfServices
}

func (c *Config) ServiceUri(name string) string {
	switch name {
	case ServiceTraffInflu:
		return c.SbiUri() + TraffInfluResUriPrefix
	case ServicePfdMng:
		return c.SbiUri() + PfdMngResUriPrefix
	case ServiceNefPfd:
		return c.SbiUri() + NefPfdMngResUriPrefix
	case ServiceNefOam:
		return c.SbiUri() + NefOamResUriPrefix
	case ServiceNefCallback:
		return c.SbiUri() + NefCallbackResUriPrefix
	default:
		return ""
	}
}
