package twilio_test

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/sms/twilio"
	"gopkg.in/yaml.v2"
)

type ConfigMock struct {
	ID           string `json:"ID" yaml:"ID"`
	SenderPhone  string `json:"senderPhone" yaml:"senderPhone"`
	TokenKeyFile string `json:"tokenKeyFile" yaml:"tokenKeyFile"`
	readToken    string
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
	config.DefaultConfPath(),
	"/path/to/config/file.conf.yml",
)
var testMessage = "Twilio tests are running, this is a confirmation of success"

func init() {
	flag.Parse()
}

func setupTwilio(t *testing.T) SMSMock {
	conf := AuthMsConfigMock{}
	confB, err := ioutil.ReadFile(*confFile)
	if err != nil {
		t.Fatalf("Error setting up: read conf file: %v", err)
	}
	err = yaml.Unmarshal(confB, &conf)
	if err != nil {
		t.Fatalf("Error setting up (reading config file): %v", err)
	}
	tokenB, err := ioutil.ReadFile(conf.SMS.Twilio.TokenKeyFile)
	if err != nil {
		t.Fatalf("Error setting up (reading token file): %v", err)
	}
	conf.SMS.Twilio.readToken = string(tokenB)
	return conf.SMS
}

func TestNewTwilio(t *testing.T) {
	validConf := setupTwilio(t)
	testCases := []struct {
		desc        string
		id          string
		token       string
		senderPhone string
		expErr      bool
	}{
		{
			desc:        "valid config",
			id:          validConf.Twilio.ID,
			token:       validConf.Twilio.readToken,
			senderPhone: validConf.Twilio.SenderPhone,
		},
		{
			desc:        "missing id",
			expErr:      true,
			token:       validConf.Twilio.readToken,
			senderPhone: validConf.Twilio.SenderPhone,
		},
		{
			desc:        "missing token",
			id:          validConf.Twilio.ID,
			expErr:      true,
			senderPhone: validConf.Twilio.SenderPhone,
		},
		{
			desc:   "missing sender phone",
			id:     validConf.Twilio.ID,
			token:  validConf.Twilio.readToken,
			expErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tw, err := twilio.NewSMSCl(tc.id, tc.token, tc.senderPhone)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("twilio.NewSMSCl(): %v", err)
			}
			if tw == nil {
				t.Fatalf("got nil SMSCl API")
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
	tw, err := twilio.NewSMSCl(conf.Twilio.ID, conf.Twilio.readToken, conf.Twilio.SenderPhone)
	if err != nil {
		t.Fatalf("twilio.NewSMSCl(): %v", err)
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
