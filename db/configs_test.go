package db_test

import (
	"testing"

	"reflect"

	"github.com/tomogoma/authms/model"
)

func TestRoach_UpsertSMTPConfig(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	smtpConf := newSMTPConfig()
	tt := []struct {
		name string
		conf model.SMTPConfig
	}{
		{
			name: "insert",
			conf: *smtpConf,
		},
		{
			name: "update",
			conf: *smtpConf,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := r.UpsertSMTPConfig(tc.conf); err != nil {
				t.Fatalf("Got error: %v", err)
			}
		})
	}
}

func TestRoach_GetSMTPConfig(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	expSMTPConf := newSMTPConfig()
	if err := r.UpsertSMTPConfig(expSMTPConf); err != nil {
		t.Fatalf("Error setting up: insert SMTP conf: %v", err)
	}
	actSMTPConf := model.SMTPConfig{}
	if err := r.GetSMTPConfig(&actSMTPConf); err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if !reflect.DeepEqual(*expSMTPConf, actSMTPConf) {
		t.Fatalf("SMTPCOnf mismatch\nExpect:\t%+v\nGot:\t%+v",
			*expSMTPConf, actSMTPConf)
	}
}

func TestRoach_GetSMTPConfig_notFound(t *testing.T) {
	conf := setup(t)
	defer tearDown(t, conf)
	r := newRoach(t, conf)
	actSMTPConf := model.SMTPConfig{}
	err := r.GetSMTPConfig(&actSMTPConf)
	if !r.IsNotFoundError(err) {
		t.Fatalf("Expected IsNotFound, got: %v", err)
	}
}

func newSMTPConfig() *model.SMTPConfig {
	return &model.SMTPConfig{
		Username:      "username",
		SSLPort:       445,
		TLSPort:       530,
		ServerAddress: "test.t.co",
		FromEmail:     "test@mail.addr",
		Password:      "a password",
	}
}
