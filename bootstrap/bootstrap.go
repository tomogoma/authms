package bootstrap

import (
	"fmt"
	"io/ioutil"
	"net/url"
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
	"path"
)

func InstantiateRoach(lg logging.Logger, conf crdb.Config) *db.Roach {
	lg.WithField(logging.FieldAction, "Initiate Cockroach DB connection").Info("started")
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
	lg.WithField(logging.FieldAction, "Initiate Cockroach DB connection").Info("completed")
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

func InstantiateSMSer(lg logging.Logger, conf config.SMS) (model.SMSer, error) {
	if conf.ActiveAPI == "" {
		lg.WithField(logging.FieldAction, "Instantiate SMS API").Infof("no active SMS API found")
		return nil, nil
	}

	lg.WithField(logging.FieldAction, "Instantiate SMS API").Infof("active SMS API is %s", conf.ActiveAPI)
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
	if conf.TestNumber != "" {
		var testMessage string
		host := hostname()
		testMessage = fmt.Sprintf("The SMS API is being used on %s", host)
		if err := s.SMS(conf.TestNumber, testMessage); err != nil {
			return s, fmt.Errorf("test SMS: %v", err)
		}
	}
	return s, nil
}

func InstantiateSMTP(rdb *db.Roach, lg logging.Logger, conf config.SMTP) *smtp.Mailer {

	lg.WithField(logging.FieldAction, "Instantiate email API").Info("started")
	emailCl, err := smtp.New(rdb)
	logging.LogFatalOnError(lg, err, "Instantiate email API")

	defer func() {
		lg.WithField(logging.FieldAction, "Instantiate email API").Info("completed")
	}()

	err = emailCl.Configured()
	if err == nil {
		return emailCl
	}

	if !emailCl.IsNotFoundError(err) {
		logging.LogWarnOnError(lg, err, "Check email API configured")
		return emailCl
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

	lg.WithField(logging.FieldAction, "Read config file").Info("started")
	conf, err := config.ReadFile(confFile)
	logging.LogFatalOnError(lg, err, "Read config file")
	lg.WithField(logging.FieldAction, "Read config file").Info("complete")

	rdb := InstantiateRoach(lg, conf.Database)

	lg.WithField(logging.FieldAction, "Set up OAuth options").Info("started")
	var authOpts []model.Option
	fb, err := InstantiateFacebook(conf.Authentication.Facebook)
	logging.LogWarnOnError(lg, err, "Set up OAuth options")
	if fb != nil {
		lg.WithField(logging.FieldAction, "Set up OAuth options").Info("using facebook for OAuth")
		authOpts = append(authOpts, model.WithFacebookCl(fb))
	}
	if len(authOpts) == 0 {
		lg.WithField(logging.FieldAction, "Set up OAuth options").Info("no OAuth options configured")
	}
	lg.WithField(logging.FieldAction, "Set up OAuth options").Info("completed")

	lg.WithField(logging.FieldAction, "Instantiate SMS API").Info("started")
	sms, err := InstantiateSMSer(lg, conf.SMS)
	logging.LogWarnOnError(lg, err, "Instantiate SMS API")
	if sms != nil {
		authOpts = append(authOpts, model.WithSMSCl(sms))
	}
	lg.WithField(logging.FieldAction, "Instantiate SMS API").Info("completed")

	emailCl := InstantiateSMTP(rdb, lg, conf.SMTP)
	authOpts = append(authOpts, model.WithEmailCl(emailCl))

	if conf.SMTP.InvitationTpl != "" {
		authOpts = append(authOpts, model.WithEmailInviteTplt(template.ParseFiles(conf.SMTP.InvitationTpl)))
	}
	if conf.SMS.InvitationTpl != "" {
		authOpts = append(authOpts, model.WithPhoneInviteTplt(template.ParseFiles(conf.SMS.InvitationTpl)))
	}
	if conf.SMTP.ResetPWDTpl != "" {
		authOpts = append(authOpts, model.WithEmailResetPassTplt(template.ParseFiles(conf.SMTP.ResetPWDTpl)))
	}
	if conf.SMS.ResetPWDTpl != "" {
		authOpts = append(authOpts, model.WithPhoneResetPassTplt(template.ParseFiles(conf.SMS.ResetPWDTpl)))
	}
	if conf.SMTP.VerifyTpl != "" {
		authOpts = append(authOpts, model.WithEmailVerifyTplt(template.ParseFiles(conf.SMTP.VerifyTpl)))
	}
	if conf.SMS.VerifyTpl != "" {
		authOpts = append(authOpts, model.WithPhoneVerifyTplt(template.ParseFiles(conf.SMS.VerifyTpl)))
	}

	srvcURL, err := url.Parse(conf.Service.URL)
	logging.LogFatalOnError(lg, err, "Parse service URL")
	srvcURL.Path = path.Join(srvcURL.Path, config.WebRootURL())

	srvcConfLg := lg.WithField(logging.FieldAction, "Configure service")
	srvcConfLg.Info("Started")

	authOpts = append(
		authOpts,
		model.WithAppName(conf.Service.AppName),
		model.WithWebAppURL(conf.Service.WebAppURL),
		model.WithServiceURL(srvcURL.String()),
		model.WithDevLockedToUser(conf.Authentication.LockDevsToUsers),
		model.WithSelfRegAllowed(conf.Authentication.AllowSelfReg),
		model.WithVerifyEmailHost(conf.Authentication.VerifyEmailHosts),
	)

	tg := InstantiateJWTHandler(lg, conf.Token.TokenKeyFile)

	a, err := model.NewAuthentication(rdb, tg, authOpts...)
	logging.LogFatalOnError(lg, err, "Instantiate Auth Model")

	g, err := api.NewGuard(rdb, api.WithMasterKey(conf.Service.MasterAPIKey))
	logging.LogFatalOnError(lg, err, "Instantate API access guard")

	srvcConfLg.Infof("Name: '%s'", conf.Service.AppName)
	srvcConfLg.Infof("WebApp: '%s'", conf.Service.WebAppURL)
	srvcConfLg.Infof("URL: '%s'", srvcURL.String())
	srvcConfLg.Infof("Locks devices to users: '%t'", conf.Authentication.LockDevsToUsers)
	srvcConfLg.Infof("Allows self registration: '%t'", conf.Authentication.AllowSelfReg)
	srvcConfLg.Infof("Verifies Email Hosts: '%t'", conf.Authentication.VerifyEmailHosts)
	srvcConfLg.Info("completed")

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
