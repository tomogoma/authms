package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/dbhelper/helper"
	"github.com/tomogoma/authms/auth/oauth"
)

const (
	RunTypeHttp = "http"
	RunTypeRPC = "rpc"
)

var ErrorInvalidRunType = fmt.Errorf("Invalid runtype; expected one of %s, %s", RunTypeRPC, RunTypeHttp)
var ErrorInvalidRegInterval = errors.New("Register interval was invalid cannot be < 1ms")

type ServiceConfig struct {
	RunType          string        `json:"runType,omitempty"`
	RegisterInterval time.Duration `json:"registerInterval,omitempty"`
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
	Service        ServiceConfig `json:"serviceConfig,omitempty"`
	Database       helper.DSN    `json:"database,omitempty"`
	Authentication auth.Config   `json:"authentication,omitempty"`
	Token          Token  `json:"token,omitempty"`
	OAuth          oauth.Config  `json:"OAuth,omitempty"`
}

func (c Config) Validate() error {
	if err := c.Service.Validate(); err != nil {
		return err
	}
	if err := c.Token.Validate(); err != nil {
		return err
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
