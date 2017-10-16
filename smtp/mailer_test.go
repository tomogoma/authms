package smtp_test

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"encoding/json"
	"io/ioutil"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/smtp"
	testingH "github.com/tomogoma/authms/testing"
	errors "github.com/tomogoma/go-typed-errors"
)

var testConf *model.SMTPConfig

func setup(t *testing.T) model.SMTPConfig {
	if testConf != nil {
		return *testConf
	}
	c := new(model.SMTPConfig)
	confD, err := ioutil.ReadFile("mailer_test_config.json")
	if err != nil {
		t.Fatalf("Error setting up: read test config file: %v", err)
	}
	if err := json.Unmarshal(confD, c); err != nil {
		t.Fatalf("Error setting up: read test config file: %v", err)
	}
	testConf = c
	return *testConf
}

func TestNew(t *testing.T) {
	tt := []struct {
		name   string
		cs     smtp.ConfigStore
		expErr bool
	}{
		{name: "valid", cs: &testingH.DBMock{}, expErr: false},
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
	fmt.Printf("valid conf: %+v\n", validConf)
	invalidConfig := model.SMTPConfig{}
	validEmail := genTestEmail(t, "Authms smtp SendEmail Test")
	tt := []struct {
		name   string
		cs     *testingH.DBMock
		email  model.SendMail
		expErr bool
	}{
		{
			name:   "valid",
			expErr: false,
			cs:     &testingH.DBMock{ExpSMTPConf: validConf},
			email:  validEmail,
		},
		{
			name:   "no config found",
			expErr: true,
			cs:     &testingH.DBMock{ExpSMTPConfErr: errors.NewNotFound("none")},
			email:  validEmail,
		},
		{
			name:   "error getting config",
			expErr: true,
			cs:     &testingH.DBMock{ExpSMTPConfErr: errors.New("database abducted")},
			email:  validEmail,
		},
		{
			name:   "invalid config",
			expErr: true,
			cs:     &testingH.DBMock{ExpSMTPConf: invalidConfig},
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

func TestMailer_SetConfig(t *testing.T) {
	validConf := setup(t)
	invalidConf := model.SMTPConfig{}
	validEmail := genTestEmail(t, "Authms smtp SetConfig Test")
	invalidEmail := model.SendMail{}
	tt := []struct {
		name       string
		conf       model.SMTPConfig
		notifEmail model.SendMail
		cs         *testingH.DBMock
		expErr     bool
	}{
		{
			name:       "valid",
			conf:       validConf,
			notifEmail: validEmail,
			cs:         &testingH.DBMock{},
			expErr:     false,
		},
		{
			name:       "invalid conf",
			conf:       invalidConf,
			notifEmail: validEmail,
			cs:         &testingH.DBMock{},
			expErr:     true,
		},
		{
			name:       "upsert error",
			conf:       validConf,
			notifEmail: validEmail,
			cs:         &testingH.DBMock{ExpUpsSMTPConfErr: errors.New("minions!")},
			expErr:     true,
		},
		{
			name:       "invalid test email",
			conf:       validConf,
			notifEmail: invalidEmail,
			cs:         &testingH.DBMock{},
			expErr:     true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := smtp.New(tc.cs)
			if err != nil {
				t.Fatalf("smtp.New(): %v", err)
			}
			err = s.SetConfig(tc.conf, tc.notifEmail)
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

func genTestEmail(t *testing.T, subject string) model.SendMail {
	tmpltData := struct {
		URL string
	}{URL: "https://google.com"}
	tmplt, err := template.New("content").Parse(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<p>
	Hello tester,
</p>
<div>
	<code>Begin Body</code><br/>
	<a href="{{.URL}}">a link</a><br/>
	<code>End Body</code><p/>
	Best regards,<br/>
	Tester Bot
</div>
<div>
	<small>The footer of the email<small>
</div>
</html>
	`)
	if err != nil {
		t.Fatalf("Error setting up: parse email body template: %v", err)
	}
	bw := new(bytes.Buffer)
	tmplt.Execute(bw, tmpltData)
	return model.SendMail{
		ToEmails: []string{"ogomatom.test3@mailinator.com", "ogomatom.test4@mailinator.com"},
		Subject:  subject,
		Body:     template.HTML(bw.String()),
	}
}
