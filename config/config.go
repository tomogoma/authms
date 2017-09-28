package config

import (
	"time"

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
	SMSAPIMessageBird    = "messageBird"

	TimeFormat = time.RFC3339
)

type ServiceConfig struct {
	RegisterInterval   time.Duration `json:"registerInterval,omitempty" yaml:"registerInterval"`
	LoadBalanceVersion string        `json:"loadBalanceVersion,omitempty" yaml:"loadBalanceVersion"`
}

type TwilioConfig struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}

type AfricasTalkingConfig struct {
	UserName   string `json:"username" yaml:"username"`
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

type MessageBirdConfig struct {
	AccountName string `json:"accountName" yaml:"accountName"`
	APIKeyFile  string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

type SMSConfig struct {
	TestNumber      string               `json:"testNumber" yaml:"testNumber"`
	Twilio          TwilioConfig         `json:"twilio" yaml:"twilio"`
	AfricasTalking  AfricasTalkingConfig `json:"africasTalking" yaml:"africasTalking"`
	ActiveAPI       string               `json:"activeAPI" yaml:"activeAPI"`
	MessageFmt      string               `json:"messageFormat" yaml:"messageFormat"`
	SMSCodeValidity time.Duration        `json:"smsCodeValidity" yaml:"smsCodeValidity"`
	MessageBird     MessageBirdConfig    `json:"messageBird" yaml:"messageBird"`
}

type Facebook struct {
	SecretFilePath string `json:"secretFilePath,omitempty" yaml:"secretFilePath"`
	ID             int64  `json:"ID,omitempty" yaml:"ID"`
}

type Auth struct {
	Facebook           Facebook      `json:"facebook,omitempty" yaml:"facebook"`
	BlackListFailCount int           `json:"blackListFailCount" yaml:"blackListFailCount"`
	BlacklistWindow    time.Duration `json:"blacklistWindow" yaml:"blacklistWindow"`
}

type General struct {
	Service        ServiceConfig    `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       cockroach.DSN    `json:"database,omitempty" yaml:"database"`
	Authentication Auth             `json:"authentication,omitempty" yaml:"authentication"`
	Token          token.ConfigStub `json:"token,omitempty" yaml:"token"`
	SMS            SMSConfig        `json:"sms" yaml:"sms"`
}
