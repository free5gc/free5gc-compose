/*
 * TNGF Configuration Factory
 */

package factory

import (
	"fmt"
	"os"

	"github.com/asaskevich/govalidator"
	yaml "gopkg.in/yaml.v2"

	"github.com/free5gc/tngf/internal/logger"
)

var TngfConfig *Config

// TODO: Support configuration update from REST api
func InitConfigFactory(f string, cfg *Config) error {
	if f == "" {
		// Use default config path
		f = TngfDefaultConfigPath
	}

	if content, err := os.ReadFile(f); err != nil {
		return fmt.Errorf("[Factory] %+v", err)
	} else {
		logger.CfgLog.Infof("Read config from [%s]", f)
		if yamlErr := yaml.Unmarshal(content, cfg); yamlErr != nil {
			return fmt.Errorf("[Factory] %+v", yamlErr)
		}
	}

	return nil
}

func CheckConfigVersion() error {
	currentVersion := TngfConfig.GetVersion()

	if currentVersion != TngfExpectedConfigVersion {
		return fmt.Errorf("config version is [%s], but expected is [%s]",
			currentVersion, TngfExpectedConfigVersion)
	}

	logger.CfgLog.Infof("config version [%s]", currentVersion)

	return nil
}

func ReadConfig(cfgPath string) (*Config, error) {
	cfg := &Config{}
	if err := InitConfigFactory(cfgPath, cfg); err != nil {
		return nil, fmt.Errorf("ReadConfig [%s] Error: %+v", cfgPath, err)
	}
	if _, err := cfg.Validate(); err != nil {
		validErrs := err.(govalidator.Errors).Errors()
		for _, validErr := range validErrs {
			logger.CfgLog.Errorf("%+v", validErr)
		}
		logger.CfgLog.Errorf("[-- PLEASE REFER TO SAMPLE CONFIG FILE COMMENTS --]")
		return nil, fmt.Errorf("Config validate Error")
	}
	return cfg, nil
}
