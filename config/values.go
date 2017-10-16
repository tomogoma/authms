package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
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

type Database struct {
	User           string `json:"user,omitempty" yaml:"user,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	Host           string `json:"host,omitempty" yaml:"host,omitempty"`
	Port           string `json:"port,omitempty" yaml:"port,omitempty"`
	DBName         string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
	ConnectTimeout int    `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"`
	SSLMode        string `json:"sslMode,omitempty" yaml:"sslMode,omitempty"`
	SSLCert        string `json:"sslCert,omitempty" yaml:"sslCert,omitempty"`
	SSLKey         string `json:"sslKey,omitempty" yaml:"sslKey,omitempty"`
	SSLRootCert    string `json:"sslRootCert,omitempty" yaml:"sslRootCert,omitempty"`
}

func (d Database) FormatDSN() string {
	return fmt.Sprintf("user='%s' password='%s'"+
		" host='%s' port='%s' dbname='%s'"+
		" connect_timeout='%d'"+
		" sslmode='%s' sslcert='%s' sslkey='%s' sslrootcert='%s'",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		d.ConnectTimeout,
		d.SSLMode,
		d.SSLCert,
		d.SSLKey,
		d.SSLRootCert,
	)
}

type General struct {
	Service        Service  `json:"serviceConfig,omitempty" yaml:"serviceConfig"`
	Database       Database `json:"database,omitempty" yaml:"database"`
	Authentication Auth     `json:"authentication,omitempty" yaml:"authentication"`
	Token          JWT      `json:"token,omitempty" yaml:"token"`
	SMS            SMS      `json:"sms" yaml:"sms"`
}

func ReadFile(fName string) (*General, error) {
	confD, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}
	conf := &General{}
	if err := yaml.Unmarshal(confD, conf); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file (%s): %v",
			fName, err)
	}
	return conf, nil
}
