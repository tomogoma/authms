package messagebird_test

import (
	"io/ioutil"
	"testing"

	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/sms/messagebird"
	testingH "github.com/tomogoma/authms/testing"
)

type configMock struct {
	testPhone   string
	accountName string
	apiKey      string
}

var testMessage = "This is a test message, the Message Bird client is being tested on " + config.CanonicalName()

func TestNewClient(t *testing.T) {
	conf := setUp(t)
	tt := []struct {
		name   string
		acName string
		apiKey string
		expErr bool
	}{
		{name: "valid deps", acName: conf.accountName, apiKey: conf.apiKey, expErr: false},
		{name: "empty account name", acName: conf.accountName, apiKey: conf.apiKey, expErr: false},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cl, err := messagebird.NewClient(tc.acName, tc.apiKey)
			if tc.expErr {
				if err != nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("messagebird.NewClient(): %v", err)
			}
			if cl == nil {
				t.Fatal("messagebird.NewClient() yielded nil *messagebird.CLient")
			}
		})
	}
}

func TestClient_SMS(t *testing.T) {
	conf := setUp(t)
	tt := []struct {
		name    string
		conf    configMock
		message string
		expErr  bool
	}{
		{
			name: "Phone no account name",
			conf: configMock{
				accountName: "+254712345678",
				apiKey:      conf.apiKey,
				testPhone:   conf.testPhone,
			},
			message: testMessage,
			expErr:  false,
		},
		{
			name: "empty account name",
			conf: configMock{
				accountName: "",
				apiKey:      conf.apiKey,
				testPhone:   conf.testPhone,
			},
			message: testMessage,
			expErr:  true,
		},
		{
			name: "account name too long",
			conf: configMock{
				accountName: "abcdefghijklmnopqrst", // should not exceed 11 chars
				apiKey:      conf.apiKey,
				testPhone:   conf.testPhone,
			},
			message: testMessage,
			expErr:  true,
		},
		{
			name: "bad api key",
			conf: configMock{
				accountName: conf.accountName,
				apiKey:      "a bad API key",
				testPhone:   conf.testPhone,
			},
			message: testMessage,
			expErr:  true,
		},
		{
			name: "empty to-Phone",
			conf: configMock{
				accountName: conf.accountName,
				apiKey:      conf.apiKey,
				testPhone:   "",
			},
			message: testMessage,
			expErr:  true,
		},
		{
			name: "bad to-phone",
			conf: configMock{
				accountName: conf.accountName,
				apiKey:      conf.apiKey,
				testPhone:   "invalid-phone-number",
			},
			message: testMessage,
			expErr:  true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cl, err := messagebird.NewClient(tc.conf.accountName, tc.conf.apiKey)
			if err != nil {
				t.Fatalf("Error setting up: messagebird.NewClient(): %v", err)
			}
			err = cl.SMS(tc.conf.testPhone, tc.message)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("messageBird.Client#SMS(): %v", err)
			}
		})
	}
}

func setUp(t *testing.T) configMock {
	conf := testingH.ReadConfig(t)
	apiKey, err := ioutil.ReadFile(conf.SMS.MessageBird.APIKeyFile)
	if err != nil {
		t.Fatalf("Error setting up: read API key file: %v", err)
	}
	return configMock{
		testPhone:   conf.SMS.TestNumber,
		accountName: conf.SMS.MessageBird.AccountName,
		apiKey:      string(apiKey),
	}
}
