package main

import (
	"flag"
	"fmt"
	"os"

	"io/ioutil"

	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/mux"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/facebook"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/http"
	"github.com/tomogoma/authms/server/rpc"
	"github.com/tomogoma/authms/sms/africas_talking"
	"github.com/tomogoma/authms/sms/twilio"
	"github.com/tomogoma/authms/store"
	"github.com/tomogoma/go-commons/auth/token"
	configH "github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/authms/sms/messagebird"
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

	pg, err := generator.NewRandom(generator.AllChars)
	logFatalOnError(err, "Initiate password generator")

	db, err := store.NewRoach(conf.Database, pg)
	logFatalOnError(err, "Instantiate DB helper")

	err = db.InitDBConnIfNotInitted()
	logWarnOnError(err, "Initiate DB connection")

	tg, err := token.NewJWTHandler(conf.Token)
	logFatalOnError(err, "Instantiate token handler (generator)")

	authOpts, err := oAuthOptions(conf.Authentication)
	logWarnOnError(err, "Set up OAuth options")

	if conf.SMS.MessageFmt != "" {
		authOpts = append(authOpts, model.WithSMSFormat(conf.SMS.MessageFmt))
	}
	if conf.SMS.SMSCodeValidity > 1*time.Minute {
		authOpts = append(authOpts, model.WithSMSValidity(conf.SMS.SMSCodeValidity))
	}

	s, err := smsAPI(conf.SMS)
	logWarnOnError(err, "Instantiate SMS API")
	if s != nil {
		authOpts = append(authOpts, model.WithSMSer(s))
	}

	a, err := model.New(tg, db, authOpts...)
	logFatalOnError(err, "Instantiate Auth Model")

	serverRPCQuitCh := make(chan error)
	serverHttpQuitCh := make(chan error)
	rpcSrv, err := rpc.NewHandler(config.CanonicalName, a)
	logFatalOnError(err, "Instantate RPC handler")
	go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)

	httpHandler, err := http.NewHandler(a)
	logFatalOnError(err, "Instantiate HTTP handler")
	go serveHttp(conf.Service, httpHandler, serverHttpQuitCh)

	select {
	case err = <-serverHttpQuitCh:
		logFatalOnError(err, "Serve HTTP")
	case err = <-serverRPCQuitCh:
		logFatalOnError(err, "Serve RPC")
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
		authOpts = append(authOpts, model.WithFB(fb))
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

func serveRPC(conf config.ServiceConfig, rpcSrv *rpc.Handler, quitCh chan error) {
	service := micro.NewService(
		micro.Name(config.CanonicalRPCName),
		micro.Version(conf.LoadBalanceVersion),
		micro.RegisterInterval(conf.RegisterInterval),
		micro.WrapHandler(rpcSrv.Wrapper),
	)
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
}

type RouteHandler interface {
	HandleRoute(r *mux.Router) error
}

func serveHttp(conf config.ServiceConfig, rh RouteHandler, quitCh chan error) {
	r := mux.NewRouter()
	if err := rh.HandleRoute(r); err != nil {
		quitCh <- errors.Newf("unable to handle route: %v", err)
	}
	service := web.NewService(
		web.Handler(r),
		web.Name(config.CanonicalWebName),
		web.Version(conf.LoadBalanceVersion),
		web.RegisterInterval(conf.RegisterInterval),
	)
	quitCh <- service.Run()
}

func readFile(path string) (string, error) {
	contentB, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read: %v", err)
	}
	return string(contentB), nil
}
