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

func InstantiateJWTHandler(lg logging.Logger, conf config.JWT) *token.Handler {
	JWTKey := []byte(conf.TokenKey)
	if len(JWTKey) == 0 {
		var err error
		JWTKey, err = ioutil.ReadFile(conf.TokenKeyFile)
		logging.LogFatalOnError(lg, err, "Read JWT key file")
	}
	jwter, err := token.NewHandler(JWTKey)
	logging.LogFatalOnError(lg, err, "Instantiate JWT handler")
	return jwter
}

func InstantiateFacebook(conf config.Facebook) (*facebook.FacebookOAuth, error) {
	if conf.ID < 1 {
		return nil, nil
	}
	fbSecret := conf.Secret
	if len(fbSecret) == 0 {
		var err error
		fbSecret, err = readFile(conf.SecretFile)
		if err != nil {
			return nil, fmt.Errorf("read facebook secret from file: %v", err)
		}
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
	var err error
	switch conf.ActiveAPI {
	case config.SMSAPIAfricasTalking:
		apiKey := conf.AfricasTalking.APIKey
		if len(apiKey) == 0 {
			apiKey, err = readFile(conf.AfricasTalking.APIKeyFile)
			if err != nil {
				return nil, fmt.Errorf("read africa's talking API key: %v", err)
			}
		}
		s, err = africas_talking.NewSMSCl(conf.AfricasTalking.UserName, apiKey)
		if err != nil {
			return nil, fmt.Errorf("new africasTalking client: %v", err)
		}
	case config.SMSAPITwilio:
		tkn := conf.Twilio.TokenKey
		if len(tkn) == 0 {
			tkn, err = readFile(conf.Twilio.TokenKeyFile)
			if err != nil {
				return nil, fmt.Errorf("read twilio token: %v", err)
			}
		}
		s, err = twilio.NewSMSCl(conf.Twilio.ID, tkn, conf.Twilio.SenderPhone)
		if err != nil {
			return nil, fmt.Errorf("new twilio client: %v", err)
		}
	case config.SMSAPIMessageBird:
		apiKey := conf.MessageBird.APIKey
		if len(apiKey) == 0 {
			apiKey, err = readFile(conf.MessageBird.APIKeyFile)
			if err != nil {
				return nil, fmt.Errorf("read messageBird API key: %v", err)
			}
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

	pass := conf.Password
	if len(pass) == 0 {
		pass, err = readFile(conf.PasswordFile)
		logging.LogWarnOnError(lg, err, "Read SMTP password file")
	}

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

	conf := readConfig(confFile, lg)

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
		authOpts = append(authOpts, model.WithEmailInviteTplt(template.New("SMTP.InvitationTpl").Parse(conf.SMTP.InvitationTpl)))
	} else if conf.SMTP.InvitationTplFile != "" {
		authOpts = append(authOpts, model.WithEmailInviteTplt(template.ParseFiles(conf.SMTP.InvitationTplFile)))
	}

	if conf.SMS.InvitationTpl != "" {
		authOpts = append(authOpts, model.WithPhoneInviteTplt(template.New("SMS.InvitationTpl").Parse(conf.SMS.InvitationTpl)))
	} else if conf.SMS.InvitationTplFile != "" {
		authOpts = append(authOpts, model.WithPhoneInviteTplt(template.ParseFiles(conf.SMS.InvitationTplFile)))
	}

	if conf.SMTP.ResetPWDTpl != "" {
		authOpts = append(authOpts, model.WithEmailResetPassTplt(template.New("SMTP.ResetPWDTpl").Parse(conf.SMTP.ResetPWDTpl)))
	} else if conf.SMTP.ResetPWDTplFile != "" {
		authOpts = append(authOpts, model.WithEmailResetPassTplt(template.ParseFiles(conf.SMTP.ResetPWDTplFile)))
	}

	if conf.SMS.ResetPWDTpl != "" {
		authOpts = append(authOpts, model.WithPhoneResetPassTplt(template.New("SMS.ResetPWDTpl").Parse(conf.SMS.ResetPWDTpl)))
	} else if conf.SMS.ResetPWDTplFile != "" {
		authOpts = append(authOpts, model.WithPhoneResetPassTplt(template.ParseFiles(conf.SMS.ResetPWDTplFile)))
	}

	if conf.SMTP.VerifyTpl != "" {
		authOpts = append(authOpts, model.WithEmailVerifyTplt(template.New("SMTP.VerifyTpl").Parse(conf.SMTP.VerifyTpl)))
	} else if conf.SMTP.VerifyTplFile != "" {
		authOpts = append(authOpts, model.WithEmailVerifyTplt(template.ParseFiles(conf.SMTP.VerifyTplFile)))
	}

	if conf.SMS.VerifyTpl != "" {
		authOpts = append(authOpts, model.WithPhoneVerifyTplt(template.New("SMS.VerifyTpl").Parse(conf.SMS.VerifyTpl)))
	} else if conf.SMS.VerifyTplFile != "" {
		authOpts = append(authOpts, model.WithPhoneVerifyTplt(template.ParseFiles(conf.SMS.VerifyTplFile)))
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

	tg := InstantiateJWTHandler(lg, conf.Token)

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

func readConfig(confFile string, lg logging.Logger) *config.General {

	conf := &config.General{}

	if len(confFile) > 0 {
		lg.WithField(logging.FieldAction, "Read config file").Info("started")
		err := config.ReadFile(confFile, conf)
		logging.LogWarnOnError(lg, err, "Read config file")
		lg.WithField(logging.FieldAction, "Read config file").Info("complete")
	}

	lg.WithField(logging.FieldAction, "Read environment config values").Info("started")
	err := config.ReadEnv(conf)
	logging.LogWarnOnError(lg, err, "Read environment config values")
	lg.WithField(logging.FieldAction, "Read environment config values").Info("complete")

	if conf.Service.Port == nil {
		port := 8080
		lg.WithField(logging.FieldAction, "Set default Port").Infof("No port config found fallback to %d", port)
		conf.Service.Port = &port
	}

	return conf
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
