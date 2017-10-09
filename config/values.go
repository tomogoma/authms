package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/tomogoma/go-commons/database/cockroach"
)

type Service struct {
	RegisterInterval   time.Duration `json:"registerInterval,omitempty" yaml:"registerInterval"`
	LoadBalanceVersion string        `json:"loadBalanceVersion,omitempty" yaml:"loadBalanceVersion"`
	MasterAPIKey       string        `json:"masterAPIKey,omitempty" yaml:"masterAPIKey"`
}

type Twilio struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}

type AfricasTalking struct {
	UserName   string `json:"username" yaml:"username"`
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

type MessageBird struct {
	AccountName string `json:"accountName" yaml:"accountName"`
	APIKeyFile  string `json:"apiKeyFile" yaml:"apiKeyFile"`
}

type SMS struct {
	TestNumber     string         `json:"testNumber" yaml:"testNumber"`
	ActiveAPI      string         `json:"activeAPI" yaml:"activeAPI"`
	Twilio         Twilio         `json:"twilio" yaml:"twilio"`
	AfricasTalking AfricasTalking `json:"africasTalking" yaml:"africasTalking"`
	MessageBird    MessageBird    `json:"messageBird" yaml:"messageBird"`
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

type JWT struct {
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}
type DSN struct {
	UName       string `yaml:"userName,omitempty"`
	Password    string `yaml:"password,omitempty"`
	Host        string `yaml:"host,omitempty"`
	DB          string `yaml:"dbName,omitempty"`
	SslCert     string `yaml:"sslCert,omitempty"`
	SslKey      string `yaml:"sslKey,omitempty"`
	SslRootCert string `yaml:"sslRootCert,omitempty"`
}

func (d DSN) DBName() string {
	return d.DB
}

func (d DSN) Validate() error {
	return nil
}

func (d DSN) FormatDSN() string {

	dsnPrefix := "postgres://"
	var dsnSuffix string

	if d.SslCert != "" {
		dsnSuffix = fmt.Sprintf("?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
			d.SslCert, d.SslKey, d.SslRootCert)
	}

	host := d.Host
	if d.Host == "" {
		host = "127.0.0.1:26257"
	}

	if d.UName != "" {
		if d.Password != "" {
			password := url.QueryEscape(d.Password)
			return fmt.Sprintf("%s%s:%s@%s/%s%s",
				dsnPrefix, d.UName, password, host, d.DB, dsnSuffix,
			)
		}
		return fmt.Sprintf("%s%s@%s/%s%s",
			dsnPrefix, d.UName, host, d.DB, dsnSuffix,
		)
	}

	return fmt.Sprintf("%s%s/%s%s", dsnPrefix, host, d.DB, dsnSuffix)

}

type General struct {
	Service        Service       `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       cockroach.DSN `json:"database,omitempty" yaml:"database"`
	Authentication Auth          `json:"authentication,omitempty" yaml:"authentication"`
	Token          JWT           `json:"token,omitempty" yaml:"token"`
	SMS            SMS           `json:"sms" yaml:"sms"`
}
