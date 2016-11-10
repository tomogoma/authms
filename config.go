package main

import (
	"time"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/token"
)

const (
	runTypeHttp = "http"
	runTypeRPC  = "rpc"
)

type Config struct {
	RunType          string        `json:"type,omitempty"`
	HttpAddress      string        `json:"httpAddress"`
	RegisterInterval time.Duration `json:"registerInterval,omitempty"`
	Database         helper.DSN    `json:"database,omitempty"`
	Authentication   auth.Config   `json:"authentication,omitempty"`
	Token            token.Config  `json:"token,omitempty"`
}

func (c Config) Validate() error {
	return nil
}
