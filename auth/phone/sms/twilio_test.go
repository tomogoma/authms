package sms_test

import (
	"flag"
	"github.com/tomogoma/authms/auth/phone/sms"
	"github.com/tomogoma/go-commons/config"
	"testing"
)

type ConfigMock struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
	TestNumber   string `json:"testNumber" yaml:"testNumber"`
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

func (c ConfigMock) TwilioTestNumber() string {
	return c.TestNumber
}

type AuthMsConfigMock struct {
	Twilio ConfigMock `json:"twilio" yaml:"twilio"`
}

var confFile = flag.String("conf", "/etc/authms/authms.conf.yml", "/path/to/config.conf.yml")
var testMessage = "Twilio tests are running, this is a confirmation of success"

func init() {
	flag.Parse()
}

func setupTwilio(t *testing.T) ConfigMock {
	conf := AuthMsConfigMock{}
	err := config.ReadYamlConfig(*confFile, &conf)
	if err != nil {
		t.Fatalf("Error setting up (reading config file): %v", err)
	}
	return conf.Twilio
}

func TestNewTwilio(t *testing.T) {
	validConf := setupTwilio(t)
	testCases := []struct {
		desc   string
		conf   sms.TwConfig
		expErr bool
	}{
		{desc: "valid config", conf: validConf},
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
		{desc: "successful", toNumber: conf.TwilioTestNumber(), message: testMessage},
		{desc: "empty number", toNumber: "", message: testMessage, expErr: true},
		{desc: "empty message", toNumber: conf.TwilioTestNumber(), message: "", expErr: true},
	}
	tw, err := sms.NewTwilio(conf)
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
