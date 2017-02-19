package config

import (
	"fmt"
	"os"
)

type Token struct {
	TknKeyFile string `yaml:"tokenkeyfile"`
}

func (c Token) TokenKeyFile() string {
	return c.TknKeyFile
}

func (c Token) Validate() error {
	if _, err := os.Stat(c.TknKeyFile); err != nil {
		return fmt.Errorf("token key file inaccessible: %s", err)
	}
	return nil
}
