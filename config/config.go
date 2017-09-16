package config

import (
	"errors"
	"time"

	"io/ioutil"

	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/database/cockroach"
)

// Compile time constants that should not be configurable
// during runtime.
const (
	Name                = "authms"
	Version             = "v1"
	Description         = "Authentication Micro-Service"
	CanonicalName       = Name + Version
	RPCNamePrefix       = ""
	CanonicalRPCName    = RPCNamePrefix + CanonicalName
	WebNamePrefix       = "go.micro.web."
	CanonicalWebName    = WebNamePrefix + CanonicalName
	DefaultSysDUnitName = CanonicalName + ".service"

	DefaultInstallDir       = "/usr/local/bin"
	DefaultInstallPath      = DefaultInstallDir + "/" + CanonicalName
	DefaultSysDUnitFilePath = "/etc/systemd/system/" + DefaultSysDUnitName
	DefaultConfDir          = "/etc/" + Name
	DefaultConfPath         = DefaultConfDir + "/" + CanonicalName + ".conf.yml"

	SMSAPITwilio         = "twilio"
	SMSAPIAfricasTalking = "africasTalking"
)

type ServiceConfig struct {
	RegisterInterval   time.Duration `json:"registerInterval,omitempty" yaml:"registerInterval"`
	LoadBalanceVersion string        `json:"loadBalanceVersion,omitempty" yaml:"loadBalanceVersion"`
}

func (sc ServiceConfig) Validate() error {
	if sc.RegisterInterval <= 1*time.Millisecond {
		return errors.New("register interval was invalid cannot be < 1ms")
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

type OAuth struct {
	FacebookSecretFileLoc string `json:"facebookSecretFileLoc,omitempty" yaml:"facebookSecretFileLoc"`
	FacebookID            int64  `json:"facebookID,omitempty" yaml:"facebookID"`
}

func (c OAuth) Validate() error {
	if c.FacebookSecretFileLoc == "" {
		return errors.New("facebook secret file location was empty")
	}
	if c.FacebookID < 1 {
		return errors.New("facebook id was invalid")
	}
	return nil
}

type Config struct {
	Service        ServiceConfig    `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       cockroach.DSN    `json:"database,omitempty" yaml:"database"`
	Authentication auth.Config      `json:"authentication,omitempty" yaml:"authentication"`
	Token          token.ConfigStub `json:"token,omitempty" yaml:"token"`
	OAuth          OAuth            `json:"OAuth,omitempty" yaml:"OAuth"`
	SMS            SMSConfig        `json:"sms" yaml:"sms"`
}
