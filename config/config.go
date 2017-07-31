package config

import (
	"errors"
	"fmt"
	"time"

	"io/ioutil"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/database/cockroach"
)

const (
	RunTypeHttp = "http"
	RunTypeRPC  = "rpc"

	SMSAPITwilio         = "twilio"
	SMSAPIAfricasTalking = "africasTalking"
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
	if sc.RegisterInterval <= 1*time.Millisecond {
		return ErrorInvalidRegInterval
	}
	return nil
}

type VerificationConfig struct {
	MessageFmt      string        `json:"messageFormat" yaml:"messageFormat"`
	SMSCodeValidity time.Duration `json:"smsCodeValidity" yaml:"smsCodeValidity"`
}

func (c VerificationConfig) MessageFormat() string {
	return c.MessageFmt
}

func (c VerificationConfig) ValidityPeriod() time.Duration {
	return c.SMSCodeValidity
}

type TwilioConfig struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}

func (c TwilioConfig) TwilioID() string {
	return c.ID
}
func (c TwilioConfig) TwilioTokenKeyFile() string {
	return c.TokenKeyFile
}
func (c TwilioConfig) TwilioSenderPhone() string {
	return c.SenderPhone
}

type AfricasTalkingConfig struct {
	UserName   string `json:"username" yaml:"username"`
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

func (atc AfricasTalkingConfig) Username() string {
	return atc.UserName
}

func (atc AfricasTalkingConfig) APIKey() string {
	keyB, err := ioutil.ReadFile(atc.APIKeyFile)
	if err != nil {
		return ""
	}
	return string(keyB)
}

type SMSConfig struct {
	TestNumber     string               `json:"testNumber" yaml:"testNumber"`
	Twilio         TwilioConfig         `json:"twilio" yaml:"twilio"`
	AfricasTalking AfricasTalkingConfig `json:"africasTalking" yaml:"africasTalking"`
	Verification   VerificationConfig   `json:"verification" yaml:"verification"`
	ActiveAPI      string               `json:"activeAPI" yaml:"activeAPI"`
}

type Config struct {
	Service        ServiceConfig    `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       cockroach.DSN    `json:"database,omitempty" yaml:"database"`
	Authentication auth.Config      `json:"authentication,omitempty" yaml:"authentication"`
	Token          token.ConfigStub `json:"token,omitempty" yaml:"token"`
	OAuth          oauth.Config     `json:"OAuth,omitempty" yaml:"OAuth"`
	SMS            SMSConfig        `json:"sms" yaml:"sms"`
}
