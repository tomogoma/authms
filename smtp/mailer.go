package smtp

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/tomogoma/authms/model"
	errors "github.com/tomogoma/go-typed-errors"
)

type ConfigStore interface {
	IsNotFoundError(error) bool
	UpsertSMTPConfig(interface{}) error
	GetSMTPConfig(interface{}) error
}

type Mailer struct {
	errors.NotFoundErrCheck
	db      ConfigStore
	tmplt   *template.Template
	appName string
}

func New(cs ConfigStore) (*Mailer, error) {
	if cs == nil {
		return nil, errors.New("ConfigStore was nil")
	}
	tmpl, err := template.New("email").Parse(emailTmplt)
	if err != nil {
		return nil, errors.Newf("parse email template: %v", err)
	}
	return &Mailer{db: cs, tmplt: tmpl}, nil
}

func (m *Mailer) SendEmail(email model.SendMail) error {
	conf, err := m.getConfig()
	if err != nil {
		return err
	}
	msg, err := m.generateMessage(email)
	if err != nil {
		return errors.Newf("generate email message: %v", err)
	}
	return sendMessage(*conf, email.ToEmails, msg)
}

func (m *Mailer) SetConfig(conf model.SMTPConfig, notifEmail model.SendMail) error {
	msg, err := m.generateMessage(notifEmail)
	if err != nil {
		return errors.Newf("generate notification email message: %v", err)
	}
	if err := sendMessage(conf, notifEmail.ToEmails, msg); err != nil {
		return errors.Newf("test configuration: %v", err)
	}
	return m.db.UpsertSMTPConfig(conf)
}

func (m *Mailer) getConfig() (*model.SMTPConfig, error) {
	conf := new(model.SMTPConfig)
	if err := m.db.GetSMTPConfig(conf); err != nil {
		if m.db.IsNotFoundError(err) {
			return nil, errors.NewNotFound("SMTP not configured")
		}
		return nil, errors.Newf("get SMTP configuration: %v", err)
	}
	return conf, nil
}

func (m *Mailer) generateMessage(r model.SendMail) ([]byte, error) {
	email := newEmail(r)
	eb := bytes.NewBuffer(make([]byte, 0, 256))
	if err := m.tmplt.Execute(eb, email); err != nil {
		return nil, errors.Newf("execute template: %v", err)
	}
	return eb.Bytes(), nil
}

func sendMessage(conf model.SMTPConfig, recipients []string, msg []byte) error {
	auth := smtp.PlainAuth("", conf.Username, conf.Password, conf.ServerAddress)
	addr := fmt.Sprintf("%s:%d", conf.ServerAddress, conf.TLSPort)
	err := smtp.SendMail(addr, auth, conf.FromEmail, recipients, msg)
	if err != nil {
		return err
	}
	return nil
}
