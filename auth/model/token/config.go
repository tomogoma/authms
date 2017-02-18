package token

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	TknKeyFile string `yaml:"tokenkeyfile"`
}

func (conf Config) TokenKeyFile() string {
	return conf.TknKeyFile
}

var ErrorUnknownKeyType = errors.New("The key type provided is unknown")

func (conf Config) Validate() error {
	if _, err := os.Stat(conf.TknKeyFile); err != nil {
		return fmt.Errorf("token key file inaccessible: %s", err)
	}
	return nil
}
