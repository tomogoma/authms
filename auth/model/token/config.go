package token

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	TokenKeyFile string
}

var ErrorUnknownKeyType = errors.New("The key type provided is unknown")

func (conf Config) Validate() error {
	if _, err := os.Stat(conf.TokenKeyFile); err != nil {
		return fmt.Errorf("token key file inaccessible: %s", err)
	}
	return nil
}
