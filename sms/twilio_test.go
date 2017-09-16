package sms_test

import (
	"flag"
	"testing"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/sms"
	configH "github.com/tomogoma/go-commons/config"
)

type ConfigMock struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
}

func (c ConfigMock) TwilioID() string {
	return c.ID
}
func (c ConfigMock) TwilioTokenKeyFile() string {
	return c.TokenKeyFile
}
func (c ConfigMock) TwilioSenderPhone() string {
	return c.SenderPhone
}

type SMSMock struct {
	TestNumber string     `json:"testNumber" yaml:"testNumber"`
	Twilio     ConfigMock `json:"twilio" yaml:"twilio"`
}

type AuthMsConfigMock struct {
	SMS SMSMock `json:"sms" yaml:"sms"`
}

var confFile = flag.String(
	"conf",
	config.DefaultConfPath,
	"/path/to/config/file.conf.yml",
)
var testMessage = "Twilio tests are running, this is a confirmation of success"

func init() {
	flag.Parse()
}

func setupTwilio(t *testing.T) SMSMock {
	conf := AuthMsConfigMock{}
	err := configH.ReadYamlConfig(*confFile, &conf)
	if err != nil {
		t.Fatalf("Error setting up (reading config file): %v", err)
	}
	return conf.SMS
}

func TestNewTwilio(t *testing.T) {
	validConf := setupTwilio(t)
	testCases := []struct {
		desc   string
		conf   sms.TwConfig
		expErr bool
	}{
		{desc: "valid config", conf: validConf.Twilio},
		{desc: "nil config", conf: nil, expErr: true},
		{desc: "missing token key file", conf: ConfigMock{}, expErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tw, err := sms.NewTwilio(tc.conf)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("sms.NewTwilio(): %v", err)
			}
			if tw == nil {
				t.Fatalf("got nil Twilio API")
			}
		})
	}
}

func TestTwilio_SMS(t *testing.T) {
	conf := setupTwilio(t)
	testCases := []struct {
		desc     string
		toNumber string
		message  string
		expErr   bool
	}{
		{desc: "successful", toNumber: conf.TestNumber, message: testMessage},
		{desc: "empty number", toNumber: "", message: testMessage, expErr: true},
		{desc: "empty message", toNumber: conf.TestNumber, message: "", expErr: true},
	}
	tw, err := sms.NewTwilio(conf.Twilio)
	if err != nil {
		t.Fatalf("sms.NewTwilio(): %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tw.SMS(tc.toNumber, tc.message)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unable to send test sms: %v", err)
			}
		})
	}
}
