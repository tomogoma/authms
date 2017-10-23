package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"

	"html/template"

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
	"github.com/tomogoma/crdb"
	token "github.com/tomogoma/jwt"
)

func InstantiateRoach(lg logging.Logger, conf crdb.Config) *db.Roach {
	var opts []db.Option
	if dsn := conf.FormatDSN(); dsn != "" {
		opts = append(opts, db.WithDSN(dsn))
	}
	if dbn := conf.DBName; dbn != "" {
		opts = append(opts, db.WithDBName(dbn))
	}
	rdb := db.NewRoach(opts...)
	err := rdb.InitDBIfNot()
	logging.LogWarnOnError(lg, err, "Initiate Cockroach DB connection")
	return rdb
}

func InstantiateJWTHandler(lg logging.Logger, tknKyF string) *token.Handler {
	JWTKey, err := ioutil.ReadFile(tknKyF)
	logging.LogFatalOnError(lg, err, "Read JWT key file")
	jwter, err := token.NewHandler(JWTKey)
	logging.LogFatalOnError(lg, err, "Instantiate JWT handler")
	return jwter
}

func InstantiateFacebook(conf config.Facebook) (*facebook.FacebookOAuth, error) {
	if conf.ID < 1 {
		return nil, nil
	}
	fbSecret, err := readFile(conf.SecretFilePath)
	if err != nil {
		return nil, fmt.Errorf("read facebook secret from file: %v", err)
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
	host := hostname()
	testMessage = fmt.Sprintf("The SMS API is being used on %s", host)
	if err := s.SMS(conf.TestNumber, testMessage); err != nil {
		return s, fmt.Errorf("test SMS: %v", err)
	}
	return s, nil
}

func InstantiateSMTP(rdb *db.Roach, lg logging.Logger, conf config.SMTP) *smtp.Mailer {

	emailCl, err := smtp.New(rdb)
	logging.LogFatalOnError(lg, err, "Instantiate email API")

	err = emailCl.Configured()
	if err == nil {
		return emailCl
	}

	if !emailCl.IsNotFoundError(err) {
		logging.LogFatalOnError(lg, err, "Check email API configured")
	}

	pass, err := readFile(conf.PasswordFile)
	logging.LogWarnOnError(lg, err, "Read SMTP password file")

	host := hostname()
	err = emailCl.SetConfig(
		smtp.Config{
			ServerAddress: conf.ServerAddress,
			TLSPort:       conf.TLSPort,
			SSLPort:       conf.SSLPort,
			Username:      conf.Username,
			FromEmail:     conf.FromEmail,
			Password:      pass,
		},
		model.SendMail{
			ToEmails: []string{conf.TestEmail},
			Subject:  "Authentication Micro-Service Started on " + host,
			Body:     template.HTML("The authentication micro-service is being used on " + host),
		},
	)
	logging.LogWarnOnError(lg, err, "Set default SMTP config")

	return emailCl
}

func Instantiate(confFile string, lg logging.Logger) (config.General, *model.Authentication, *api.Guard, *db.Roach, model.JWTEr, model.SMSer, *smtp.Mailer) {

	conf, err := config.ReadFile(confFile)
	logging.LogFatalOnError(lg, err, "Read config file")

	rdb := InstantiateRoach(lg, conf.Database)
	tg := InstantiateJWTHandler(lg, conf.Token.TokenKeyFile)

	var authOpts []model.Option
	fb, err := InstantiateFacebook(conf.Authentication.Facebook)
	logging.LogWarnOnError(lg, err, "Set up OAuth options")
	if fb != nil {
		authOpts = append(authOpts, model.WithFacebookCl(fb))
	}

	sms, err := InstantiateSMSer(conf.SMS)
	logging.LogWarnOnError(lg, err, "Instantiate SMS API")
	if sms != nil {
		authOpts = append(authOpts, model.WithSMSCl(sms))
	}

	emailCl := InstantiateSMTP(rdb, lg, conf.SMTP)
	authOpts = append(authOpts, model.WithEmailCl(emailCl))

	authOpts = append(
		authOpts,
		model.WithAppName(conf.Service.AppName),
		model.WithDevLockedToUser(conf.Authentication.LockDevsToUsers),
		model.WithSelfRegAllowed(conf.Authentication.AllowSelfReg),
	)

	a, err := model.NewAuthentication(rdb, tg, authOpts...)
	logging.LogFatalOnError(lg, err, "Instantiate Auth Model")

	g, err := api.NewGuard(rdb, api.WithMasterKey(conf.Service.MasterAPIKey))
	logging.LogFatalOnError(lg, err, "Instantate API access guard")

	return *conf, a, g, rdb, tg, sms, emailCl
}

func hostname() string {
	hostName, err := os.Hostname()
	if err != nil {
		return "an unknown host"
	}
	return hostName
}

func readFile(path string) (string, error) {
	contentB, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read: %v", err)
	}
	return string(contentB), nil
}
