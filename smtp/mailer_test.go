package smtp_test

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"testing"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/smtp"
	"github.com/tomogoma/go-commons/errors"
)

type ConfigStoreMock struct {
	errors.NotFoundErrCheck
	expConf []byte
	expErr  error
}

func (c *ConfigStoreMock) GetSMTPConfig() ([]byte, error) {
	return c.expConf, c.expErr
}

var testConf []byte

func setup(t *testing.T) []byte {
	if len(testConf) > 0 {
		return testConf
	}
	var err error
	testConf, err = ioutil.ReadFile("mailer_test_config.json")
	if err != nil {
		t.Fatalf("Error setting up: read test config: %v", err)
	}
	return testConf
}

func TestNew(t *testing.T) {
	tt := []struct {
		name   string
		cs     smtp.ConfigStore
		expErr bool
	}{
		{name: "valid", cs: &ConfigStoreMock{}, expErr: false},
		{name: "nil ConfigStore", cs: nil, expErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := smtp.New(tc.cs)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if s == nil {
				t.Fatalf("Got nil *smtp.Mailer")
			}
		})
	}
}

func TestMailer_SendEmail(t *testing.T) {
	validConf := setup(t)
	invalidJSON := append(validConf, []byte("}{]")...)
	invalidConfigStr := smtp.Config{}
	invalidConfig, err := json.Marshal(invalidConfigStr)
	if err != nil {
		t.Fatalf("Error setting up: marshal test config: %v", err)
	}
	tmpltData := struct {
		URL string
	}{URL: "https://google.com"}
	tmplt, err := template.New("content").Parse(`
<code>Begin Body</code><br/>
<a href="{{.URL}}">a link</a><br/>
<code>End Body</code>
	`)
	if err != nil {
		t.Fatalf("Error setting up: parse email body template: %v", err)
	}
	bw := new(bytes.Buffer)
	tmplt.Execute(bw, tmpltData)
	validEmail := model.SendMail{
		ToEmails:      []string{"ogomatom.test1@mailinator.com", "ogomatom.test2@mailinator.com"},
		RecipientName: "<b>Tester</b>",
		Subject:       "Authms SMTP Test",
		Body:          template.HTML(bw.String()),
		Signature:     "Best regards,<br/>Tester Bot",
		Footer:        "<small>The footer of the email<small>",
	}
	tt := []struct {
		name   string
		cs     *ConfigStoreMock
		email  model.SendMail
		expErr bool
	}{
		{
			name:   "valid",
			expErr: false,
			cs:     &ConfigStoreMock{expConf: validConf},
			email:  validEmail,
		},
		{
			name:   "no config found",
			expErr: true,
			cs:     &ConfigStoreMock{expErr: errors.NewNotFound("none")},
			email:  validEmail,
		},
		{
			name:   "error getting config",
			expErr: true,
			cs:     &ConfigStoreMock{expErr: errors.NewNotFound("none")},
			email:  validEmail,
		},
		{
			name:   "config not json",
			expErr: true,
			cs:     &ConfigStoreMock{expConf: invalidJSON},
			email:  validEmail,
		},
		{
			name:   "invalid config values",
			expErr: true,
			cs:     &ConfigStoreMock{expConf: invalidConfig},
			email:  validEmail,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := smtp.New(tc.cs)
			if err != nil {
				t.Fatalf("smtp.New(): %v", err)
			}
			err = s.SendEmail(tc.email)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
		})
	}
}
