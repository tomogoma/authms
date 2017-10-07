package main

import (
	"flag"
	"fmt"
	"os"

	"io/ioutil"

	http2 "net/http"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/mux"
	"github.com/limetext/log4go"
	"github.com/micro/go-web"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/db"
	"github.com/tomogoma/authms/facebook"
	"github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/authms/sms/africas_talking"
	"github.com/tomogoma/authms/sms/messagebird"
	"github.com/tomogoma/authms/sms/twilio"
	"github.com/tomogoma/authms/smtp"
	"github.com/tomogoma/go-commons/auth/token"
	configH "github.com/tomogoma/go-commons/config"
)

type defLogWriter struct {
	lg log4go.Logger
}

func (dlw defLogWriter) Write(p []byte) (int, error) {
	dlw.lg.Info("%s", p)
	return len(p), nil
}

func main() {

	confFile := flag.String("conf", config.DefaultConfPath, "location of config file")
	flag.Parse()

	conf := config.General{}
	err := configH.ReadYamlConfig(*confFile, &conf)
	logFatalOnError(err, "Read config file")

	rdb := db.NewRoach(
		db.WithDSN(conf.Database.FormatDSN()),
		db.WithDBName(conf.Database.DBName()),
	)
	err = rdb.InitDBIfNot()
	logWarnOnError(err, "Initiate DB connection")

	JWTKey, err := ioutil.ReadFile(conf.Token.TokenKeyFile)
	logFatalOnError(err, "Read JWT key file")
	tg, err := token.NewJWTHandler(JWTKey)
	logFatalOnError(err, "Instantiate token handler (generator)")

	authOpts, err := oAuthOptions(conf.Authentication)
	logWarnOnError(err, "Set up OAuth options")

	s, err := smsAPI(conf.SMS)
	logWarnOnError(err, "Instantiate SMS API")
	if s != nil {
		authOpts = append(authOpts, model.WithSMSCl(s))
	}

	emailCl, err := smtp.New(rdb)
	logWarnOnError(err, "Instantiate email API")
	if emailCl != nil {
		authOpts = append(authOpts, model.WithEmailCl(emailCl))
	}

	a, err := model.NewAuthentication(rdb, tg, authOpts...)
	logFatalOnError(err, "Instantiate Auth Model")

	//serverRPCQuitCh := make(chan error)
	//rpcSrv, err := rpc.NewHandler(config.CanonicalName, a)
	//logFatalOnError(err, "Instantate RPC handler")
	//go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)

	g, err := api.NewGuard(rdb, api.WithMasterKey(conf.Service.MasterAPIKey))
	logFatalOnError(err, "Instantate API access guard")
	serverHttpQuitCh := make(chan error)
	httpHandler, err := http.NewHandler(a, g)
	logFatalOnError(err, "Instantiate HTTP handler")
	go serveHttp(conf.Service, httpHandler, serverHttpQuitCh)

	select {
	case err = <-serverHttpQuitCh:
		logFatalOnError(err, "Serve HTTP")
		//case err = <-serverRPCQuitCh:
		//	logFatalOnError(err, "Serve RPC")
	}
}

func logWarnOnError(err error, action string) {
	if err != nil {
		logrus.WithField(logging.FieldAction, action).Warn(err)
	}
}

func logFatalOnError(err error, action string) {
	if err != nil {
		logrus.WithField(logging.FieldAction, action).Fatal(err)
	}
}

func oAuthOptions(conf config.Auth) ([]model.Option, error) {
	authOpts := make([]model.Option, 0)
	if conf.Facebook.ID > 0 {
		fbSecret, err := readFile(conf.Facebook.SecretFilePath)
		if err != nil {
			return authOpts, errors.Newf("facebook secret file: %v", err)
		}
		fb, err := facebook.New(conf.Facebook.ID, fbSecret)
		if err != nil {
			return authOpts, errors.Newf("facebook client: %v", err)
		}
		authOpts = append(authOpts, model.WithFacebookCl(fb))
	}
	return authOpts, nil
}

func smsAPI(conf config.SMSConfig) (model.SMSer, error) {
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

//func serveRPC(conf config.ServiceConfig, rpcSrv *rpc.Handler, quitCh chan error) {
//service := micro.NewService(
//	micro.Name(config.CanonicalRPCName),
//	micro.Version(conf.LoadBalanceVersion),
//	micro.RegisterInterval(conf.RegisterInterval),
//	micro.WrapHandler(rpcSrv.Wrapper),
//)
//authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
//err := service.Run()
//quitCh <- err
//}

type RouteHandler interface {
	HandleRoute(r *mux.Router) error
}

func serveHttp(conf config.ServiceConfig, h http2.Handler, quitCh chan error) {
	srvc := web.NewService(
		web.Handler(h),
		web.Name(config.CanonicalWebName),
		web.Version(conf.LoadBalanceVersion),
		web.RegisterInterval(conf.RegisterInterval),
	)
	quitCh <- srvc.Run()
}

func readFile(path string) (string, error) {
	contentB, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read: %v", err)
	}
	return string(contentB), nil
}
