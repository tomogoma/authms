package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/auth/token"
)

const (
	RunTypeHttp = "http"
	RunTypeRPC = "rpc"
)

var ErrorInvalidRunType = fmt.Errorf("Invalid runtype; expected one of %s, %s", RunTypeRPC, RunTypeHttp)
var ErrorInvalidRegInterval = errors.New("Register interval was invalid cannot be < 1ms")

type ServiceConfig struct {
	RunType          string        `json:"runType,omitempty" yaml:"runType"`
	RegisterInterval time.Duration `json:"registerInterval,omitempty" yaml:"registerInterval"`
}

func (sc ServiceConfig) Validate() error {
	if sc.RunType != RunTypeHttp && sc.RunType != RunTypeRPC {
		return ErrorInvalidRunType
	}
	if sc.RegisterInterval <= 1 * time.Millisecond {
		return ErrorInvalidRegInterval
	}
	return nil
}

type Config struct {
	Service        ServiceConfig `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       cockroach.DSN    `json:"database,omitempty" yaml:"database"`
	Authentication auth.Config   `json:"authentication,omitempty" yaml:"authentication"`
	Token          token.DefaultConfig  `json:"token,omitempty" yaml:"token"`
	OAuth          oauth.Config  `json:"OAuth,omitempty" yaml:"OAuth"`
}

func (c Config) Validate() error {
	if err := c.Service.Validate(); err != nil {
		return err
	}
	if c.Token.TokenKeyFile() == "" {
		return errors.New("Token key file not provided")
	}
	if err := c.Authentication.Validate(); err != nil {
		return err
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	if err := c.OAuth.Validate(); err != nil {
		return err
	}
	return nil
}
