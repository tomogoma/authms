package sms_test

import (
	"testing"
	"github.com/tomogoma/authms/auth/phone/sms"
	"flag"
	"github.com/tomogoma/go-commons/config"
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

func init() {
	flag.Parse()
}

func Test(t *testing.T) {
	conf := AuthMsConfigMock{}
	err := config.ReadYamlConfig(*confFile, &conf)
	if err != nil {
		t.Fatalf("Error setting up (reading config file): %v", err)
	}
	_, err = sms.NewTwilio(conf.Twilio)
	if err != nil {
		t.Fatalf("sms.NewTwilio(): %v", err)
	}
}
