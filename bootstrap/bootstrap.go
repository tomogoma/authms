package bootstrap

import (
	"io/ioutil"

	"fmt"
	"os"

	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/facebook"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/sms/africas_talking"
	"github.com/tomogoma/authms/sms/messagebird"
	"github.com/tomogoma/authms/sms/twilio"
	"github.com/tomogoma/authms/smtp"
	"github.com/tomogoma/go-commons/auth/token"
	configH "github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/go-commons/errors"
)

func InstantiateRoach(conf cockroach.DSN) *db.Roach {
	var opts []db.Option
	if dsn := conf.FormatDSN(); dsn != "" {
		opts = append(opts, db.WithDSN(dsn))
	}
	if dbn := conf.DBName(); dbn != "" {
		opts = append(opts, db.WithDBName(dbn))
	}
	rdb := db.NewRoach(opts...)
	err := rdb.InitDBIfNot()
	logging.LogWarnOnError(err, "Initiate Cockroach DB connection")
	return rdb
}

func InstantiateJWTHandler(tknKyF string) *token.JWTHandler {
	JWTKey, err := ioutil.ReadFile(tknKyF)
	logging.LogFatalOnError(err, "Read JWT key file")
	jwter, err := token.NewJWTHandler(JWTKey)
	logging.LogFatalOnError(err, "Instantiate JWT handler")
	return jwter
}

func InstantiateFacebook(conf config.Facebook) (*facebook.FacebookOAuth, error) {
	if conf.ID < 1 {
		return nil, nil
	}
	fbSecret, err := readFile(conf.SecretFilePath)
	if err != nil {
		return nil, errors.Newf("read facebook secret from file: %v", err)
	}
	fb, err := facebook.New(conf.ID, fbSecret)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func InstantiateSMSer(conf config.SMS) (model.SMSer, error) {
	if conf.ActiveAPI == "" {
		return nil, nil
	}
	var s model.SMSer
	switch conf.ActiveAPI {
	case config.SMSAPIAfricasTalking:
		apiKey, err := readFile(conf.AfricasTalking.APIKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read africa's talking API key: %v", err)
		}
		s, err = africas_talking.NewSMSCl(conf.AfricasTalking.UserName, apiKey)
		if err != nil {
			return nil, fmt.Errorf("new africasTalking client: %v", err)
		}
	case config.SMSAPITwilio:
		tkn, err := readFile(conf.Twilio.TokenKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read twilio token: %v", err)
		}
		s, err = twilio.NewSMSCl(conf.Twilio.ID, tkn, conf.Twilio.SenderPhone)
		if err != nil {
			return nil, fmt.Errorf("new twilio client: %v", err)
		}
	case config.SMSAPIMessageBird:
		apiKey, err := readFile(conf.MessageBird.APIKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read messageBird API key: %v", err)
		}
		s, err = messagebird.NewClient(conf.MessageBird.AccountName, apiKey)
		if err != nil {
			return nil, fmt.Errorf("new messageBird client: %v", err)
		}
	default:
		return nil, fmt.Errorf("invalid API selected can be %s or %s",
			config.SMSAPIAfricasTalking, config.SMSAPITwilio)
	}
	var testMessage string
	if hostName, err := os.Hostname(); err == nil {
		testMessage = fmt.Sprintf("The SMS API is being used on %s", hostName)
	} else {
		testMessage = "The SMS API is being used on an unknown host"
	}
	if err := s.SMS(conf.TestNumber, testMessage); err != nil {
		return s, fmt.Errorf("test SMS: %v", err)
	}
	return s, nil
}

func Instantiate(confFile string) (config.General, *model.Authentication, *api.Guard, *db.Roach, *token.JWTHandler, model.SMSer, *smtp.Mailer) {

	conf := config.General{}
	err := configH.ReadYamlConfig(confFile, &conf)
	logging.LogFatalOnError(err, "Read config file")

	rdb := InstantiateRoach(conf.Database)
	tg := InstantiateJWTHandler(conf.Token.TokenKeyFile)

	var authOpts []model.Option

	fb, err := InstantiateFacebook(conf.Authentication.Facebook)
	logging.LogWarnOnError(err, "Set up OAuth options")
	if fb != nil {
		authOpts = append(authOpts, model.WithFacebookCl(fb))
	}

	sms, err := InstantiateSMSer(conf.SMS)
	logging.LogWarnOnError(err, "Instantiate SMS API")
	if sms != nil {
		authOpts = append(authOpts, model.WithSMSCl(sms))
	}

	emailCl, err := smtp.New(rdb)
	logging.LogFatalOnError(err, "Instantiate email API")
	authOpts = append(authOpts, model.WithEmailCl(emailCl))

	a, err := model.NewAuthentication(rdb, tg, authOpts...)
	logging.LogFatalOnError(err, "Instantiate Auth Model")

	g, err := api.NewGuard(rdb, api.WithMasterKey(conf.Service.MasterAPIKey))
	logging.LogFatalOnError(err, "Instantate API access guard")

	return conf, a, g, rdb, tg, sms, emailCl
}

func readFile(path string) (string, error) {
	contentB, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read: %v", err)
	}
	return string(contentB), nil
}
