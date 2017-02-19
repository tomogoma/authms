package config

import (
	"fmt"
	conf_helper "github.com/tomogoma/go-commons/config"
)

func ReadFile(filePath string) (Config, error) {
	config := Config{}
	err := conf_helper.ReadYamlConfig(filePath, &config)
	if err = config.Validate(); err != nil {
		return config, fmt.Errorf("Config file had invalid values: %s",
			err)
	}
	return config, nil
}
