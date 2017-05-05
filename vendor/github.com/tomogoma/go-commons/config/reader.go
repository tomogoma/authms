package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/tomogoma/go-commons/errors"
)

func ReadYamlConfig(confFile string, conf interface{}) error {
	configB, err := ioutil.ReadFile(confFile)
	if err != nil {
		return errors.Newf("unable to read conf file at '%s': %s",
			confFile, err)
	}
	if err := yaml.Unmarshal(configB, conf); err != nil {
		return errors.Newf("unable to unmarshal conf file values for" +
			" file at '%s': %s", confFile, err)
	}
	return nil
}