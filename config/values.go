package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/Netflix/go-env"
	"github.com/tomogoma/crdb"
	"gopkg.in/yaml.v2"
)

type MicroService struct {
	RegisterInterval   time.Duration `json:"registerInterval" yaml:"registerInterval" env:"MS_REGISTER_INTERVAL"`
	LoadBalanceVersion string        `json:"loadBalanceVersion" yaml:"loadBalanceVersion" env:"MS_LOAD_BALANCE_VERSION"`
}
type Service struct {
	MicroService
	MasterAPIKey   string   `json:"masterAPIKey" yaml:"masterAPIKey" env:"SRVC_MASTER_API_KEY"`
	AllowedOrigins []string `json:"allowedOrigins" yaml:"allowedOrigins" env:"-"`
	AppName        string   `json:"appName" yaml:"appName" env:"SRVC_APP_NAME"`
	WebAppURL      string   `json:"webAppURL" yaml:"webAppURL" env:"SRVC_WEB_APP_URL"`
	URL            string   `json:"URL" yaml:"URL" env:"SRVC_URL"`
	Port           *int     `json:"port" yaml:"port" env:"PORT"`
}

type Twilio struct {
	ID           string `json:"ID" yaml:"ID" env:"TWILIO_ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone" env:"TWILIO_SENDER_PHONE"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
	TokenKey     string `json:"-" yaml:"-" env:"TWILIO_TOKEN_KEY"`
}

type AfricasTalking struct {
	UserName   string `json:"username" yaml:"username" env:"AFRCAS_TLKN_USERNAME"`
	APIKeyFile string `json:"apiKeyFile" yaml:"apiKeyFile"`
	APIKey     string `json:"-" yaml:"-" env:"AFRCAS_TLKN_API_KEY"`
}

type MessageBird struct {
	AccountName string `json:"accountName" yaml:"accountName" env:"MSG_BRD_ACCOUNT_NAME"`
	APIKeyFile  string `json:"apiKeyFile" yaml:"apiKeyFile"`
	APIKey      string `json:"-" yaml:"-" env:"MSG_BRD_API_KEY"`
}

type SMS struct {
	TestNumber        string         `json:"testNumber" yaml:"testNumber" env:"SMS_TEST_NUMBER"`
	ActiveAPI         string         `json:"activeAPI" yaml:"activeAPI" env:"SMS_ACTIVE_API"`
	Twilio            Twilio         `json:"twilio" yaml:"twilio"`
	AfricasTalking    AfricasTalking `json:"africasTalking" yaml:"africasTalking"`
	MessageBird       MessageBird    `json:"messageBird" yaml:"messageBird"`
	InvitationTplFile string         `json:"invitationTpl" yaml:"invitationTpl"`
	ResetPWDTplFile   string         `json:"resetPwdTpl" yaml:"resetPwdTpl"`
	VerifyTplFile     string         `json:"verifyTpl" yaml:"verifyTpl"`
	InvitationTpl     string         `json:"-" yaml:"-" env:"SMS_INVITATION_TPL"`
	ResetPWDTpl       string         `json:"-" yaml:"-" env:"SMS_RESET_PWD_TPL"`
	VerifyTpl         string         `json:"-" yaml:"-" env:"SMS_VERIFY_TPL"`
}

type Facebook struct {
	SecretFile string `json:"secretFilePath" yaml:"secretFilePath"`
	Secret     string `json:"-" yaml:"-" env:"FBK_SECRET"`
	ID         int64  `json:"ID" yaml:"ID" env:"FBK_ID"`
}

type Auth struct {
	AllowSelfReg       bool          `json:"allowSelfReg" yaml:"allowSelfReg" env:"AUTH_ALLOW_SELFREG"`
	LockDevsToUsers    bool          `json:"lockDevsToUsers" yaml:"lockDevsToUsers" env:"AUTH_LOCK_DEVS_TO_USERS"`
	Facebook           Facebook      `json:"facebook" yaml:"facebook"`
	BlackListFailCount int           `json:"blackListFailCount" yaml:"blackListFailCount" env:"AUTH_BLACKLIST_FAIL_COUNT"`
	BlacklistWindow    time.Duration `json:"blacklistWindow" yaml:"blacklistWindow" env:"AUTH_BLACKLIST_WINDOW"`
	VerifyEmailHosts   bool          `json:"verifyEmailHosts" yaml:"verifyEmailHosts" env:"AUTH_VERIFY_EMAIL_HOSTS"`
}

type JWT struct {
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
	TokenKey     string `json:"-" yaml:"-" env:"AUTH_JWT_TOKEN_KEY"`
}

type SMTP struct {
	ServerAddress     string `json:"serverAddress" yaml:"serverAddress" env:"SMTP_SERVER_ADDRESS"`
	TLSPort           int32  `json:"TLSPort" yaml:"TLSPort" env:"SMTP_TLS_PORT"`
	SSLPort           int32  `json:"SSLPort" yaml:"SSLPort" env:"SMTP_SSL_PORT"`
	Username          string `json:"username" yaml:"username" env:"SMTP_USERNAME"`
	PasswordFile      string `json:"passwordFile" yaml:"passwordFile"`
	Password          string `json:"-" yaml:"-" env:"SMTP_PASSWORD"`
	FromEmail         string `json:"fromEmail" yaml:"fromEmail" env:"SMTP_FROM_EMAIL"`
	TestEmail         string `json:"testEmail" yaml:"testEmail" env:"SMTP_TEST_EMAIL"`
	CreatedAt         string `json:"createdAt" yaml:"createdAt"`
	UpdatedAt         string `json:"updatedAt" yaml:"updatedAt"`
	InvitationTplFile string `json:"invitationTpl" yaml:"invitationTpl"`
	ResetPWDTplFile   string `json:"resetPwdTpl" yaml:"resetPwdTpl"`
	VerifyTplFile     string `json:"verifyTpl" yaml:"verifyTpl"`
	InvitationTpl     string `json:"-" yaml:"-" env:"SMTP_INVITATION_TPL"`
	ResetPWDTpl       string `json:"-" yaml:"-" env:"SMTP_RESETPWD_TPL"`
	VerifyTpl         string `json:"-" yaml:"-" env:"SMTP_VERIFY_TPL"`
}

type General struct {
	Service        Service     `json:"serviceConfig" yaml:"serviceConfig"`
	Database       crdb.Config `json:"database" yaml:"database"`
	Authentication Auth        `json:"authentication" yaml:"authentication"`
	Token          JWT         `json:"token" yaml:"token"`
	SMTP           SMTP        `json:"SMTP" yaml:"SMTP"`
	SMS            SMS         `json:"sms" yaml:"sms"`
	DatabaseURL    string      `json:"databaseURL" yaml:"databaseURL"`
}

func ReadFile(fName string, conf *General) error {
	confD, err := ioutil.ReadFile(fName)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(confD, conf); err != nil {
		return fmt.Errorf("unmarshal yaml file (%s): %v",
			fName, err)
	}
	return nil
}

func ReadEnv(conf *General) error {
	if conf == nil {
		return errors.New("nil config")
	}

	envSet, err := unmarshalServcConf(&conf.Service)
	if err != nil {
		return fmt.Errorf("read service config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.Service.MicroService); err != nil {
		return fmt.Errorf("read Microservice config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.SMS); err != nil {
		return fmt.Errorf("read SMS config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.SMS.Twilio); err != nil {
		return fmt.Errorf("read Twilio SMS config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.SMS.AfricasTalking); err != nil {
		return fmt.Errorf("read Africa's Talking SMS config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.SMS.MessageBird); err != nil {
		return fmt.Errorf("read Message Bird SMS config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.Authentication); err != nil {
		return fmt.Errorf("read auth config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.Authentication.Facebook); err != nil {
		return fmt.Errorf("read facebook auth config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.Token); err != nil {
		return fmt.Errorf("read token config values: %v", err)
	}
	if err := env.Unmarshal(envSet, &conf.SMTP); err != nil {
		return fmt.Errorf("read smtp config values: %v", err)
	}

	if dbURL, exists := envSet[EnvKeyDatabaseURL]; exists {
		conf.DatabaseURL = dbURL
	} else {
		if err := unmarshalDBConf(envSet, &conf.Database); err != nil {
			return fmt.Errorf("read database config values: %v", err)
		}
	}

	return nil
}

const (
	EnvKeyDbUser             = "DB_USER"
	EnvKeyDbPassword         = "DB_PASSWORD"
	EnvKeyDbHost             = "DB_HOST"
	EnvKeyDbPort             = "DB_PORT"
	EnvKeyDbName             = "DB_NAME"
	EnvKeyDbSSLMode          = "DB_SSL_MODE"
	EnvKeySrvcAllowedOrigins = "SRVC_ALLOWED_ORIGINS"
	EnvKeyDatabaseURL        = "DATABASE_URL"
)

func unmarshalServcConf(conf *Service) (env.EnvSet, error) {
	es, err := env.UnmarshalFromEnviron(conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}
	if allowedOrigins, exists := es[EnvKeySrvcAllowedOrigins]; exists {
		conf.AllowedOrigins = strings.Split(allowedOrigins, ",")
	}
	return es, nil
}

func unmarshalDBConf(es env.EnvSet, conf *crdb.Config) error {
	if dbUser, exists := es[EnvKeyDbUser]; exists {
		conf.User = dbUser
	}
	if user, exists := es[EnvKeyDbUser]; exists {
		conf.User = user
	}
	if password, exists := es[EnvKeyDbPassword]; exists {
		conf.Password = password
	}
	if host, exists := es[EnvKeyDbHost]; exists {
		conf.Host = host
	}
	if dBName, exists := es[EnvKeyDbName]; exists {
		conf.DBName = dBName
	}
	if sslMode, exists := es[EnvKeyDbSSLMode]; exists {
		conf.SSLMode = sslMode
	}
	if portStr, exists := es[EnvKeyDbPort]; exists {
		port, err := strconv.ParseInt(portStr, 10, 64)
		if err != nil {
			return fmt.Errorf("db port needs to be an integer, found %v", port)
		}
		conf.Port = int(port)
	}
	return nil
}
