package smtp_test

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"testing"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/smtp"
	"github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/go-commons/errors"
)

type ConfigStoreMock struct {
	errors.NotFoundErrCheck
	expConf model.SMTPConfig
	expErr  error
}

func (c *ConfigStoreMock) GetSMTPConfig(conf interface{}) error {
	if c.expErr != nil {
		return c.expErr
	}
	rve := reflect.ValueOf(conf).Elem()
	rve.FieldByName("Username").SetString(c.expConf.Username)
	rve.FieldByName("Password").SetString(c.expConf.Password)
	rve.FieldByName("FromEmail").SetString(c.expConf.FromEmail)
	rve.FieldByName("ServerAddress").SetString(c.expConf.ServerAddress)
	rve.FieldByName("TLSPort").SetInt(int64(c.expConf.TLSPort))
	rve.FieldByName("SSLPort").SetInt(int64(c.expConf.SSLPort))
	return nil
}

func (c *ConfigStoreMock) UpsertSMTPConfig(conf interface{}) error {
	return c.expErr
}

var testConf *model.SMTPConfig

func setup(t *testing.T) model.SMTPConfig {
	if testConf != nil {
		return *testConf
	}
	c := new(model.SMTPConfig)
	if err := config.ReadJSONConfig("mailer_test_config.json", c); err != nil {
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
	fmt.Printf("valid conf: %+v\n", validConf)
	invalidConfig := model.SMTPConfig{}
	validEmail := genTestEmail(t, "Authms smtp SendEmail Test")
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
			cs:     &ConfigStoreMock{expErr: errors.New("database abducted")},
			email:  validEmail,
		},
		{
			name:   "invalid config",
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

func TestMailer_SetConfig(t *testing.T) {
	validConf := setup(t)
	invalidConf := model.SMTPConfig{}
	validEmail := genTestEmail(t, "Authms smtp SetConfig Test")
	invalidEmail := model.SendMail{}
	tt := []struct {
		name       string
		conf       model.SMTPConfig
		notifEmail model.SendMail
		cs         *ConfigStoreMock
		expErr     bool
	}{
		{
			name:       "valid",
			conf:       validConf,
			notifEmail: validEmail,
			cs:         &ConfigStoreMock{},
			expErr:     false,
		},
		{
			name:       "invalid conf",
			conf:       invalidConf,
			notifEmail: validEmail,
			cs:         &ConfigStoreMock{},
			expErr:     true,
		},
		{
			name:       "upsert error",
			conf:       validConf,
			notifEmail: validEmail,
			cs:         &ConfigStoreMock{expErr: errors.New("minions!")},
			expErr:     true,
		},
		{
			name:       "invalid test email",
			conf:       validConf,
			notifEmail: invalidEmail,
			cs:         &ConfigStoreMock{},
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
