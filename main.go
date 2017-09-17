package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"io/ioutil"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/mux"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/facebook"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/http"
	"github.com/tomogoma/authms/server/rpc"
	"github.com/tomogoma/authms/sms"
	"github.com/tomogoma/authms/sms/africas_talking"
	"github.com/tomogoma/authms/sms/twilio"
	"github.com/tomogoma/authms/store"
	"github.com/tomogoma/authms/verification"
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

var confFile = flag.String(
	"conf",
	config.DefaultConfPath,
	"location of config file",
)

func main() {
	flag.Parse()
	conf := config.General{}
	err := configH.ReadYamlConfig(*confFile, &conf)
	if err != nil {
		log.Fatalf("Error Reading config file: %s", err)
	}
	lg := log4go.NewDefaultLogger(log4go.FINEST)
	log.SetOutput(defLogWriter{lg: lg})
	defer func() {
		runtime.Gosched()
		time.Sleep(100 * time.Millisecond)
	}()
	tg, err := token.NewJWTHandler(conf.Token)
	if err != nil {
		lg.Critical("Error instantiating token generator: %s", err)
		return
	}
	authQuitCh := make(chan error)
	pg, err := generator.NewRandom(generator.AllChars)
	if err != nil {
		lg.Critical("Error instantiating password generator: %s", err)
		return
	}
	db, err := store.NewRoach(conf.Database, pg)
	if err != nil {
		lg.Critical("Error instantiating db helper: %v", err)
		return
	}
	if err := db.InitDBConnIfNotInitted(); err != nil {
		lg.Warn("Error initiating connection to db: %v", err)
	}
	s, err := smsAPIOrStub(conf.SMS)
	if err != nil {
		lg.Warn("Error instantiating SMS API: %v", conf.SMS.ActiveAPI, err)
	}
	ng, err := generator.NewRandom(generator.NumberChars)
	if err != nil {
		lg.Critical("Error instantiating number generator: %s", err)
		return
	}
	pv, err := verification.New(conf.SMS.Verification.MessageFmt,
		conf.SMS.Verification.SMSCodeValidity, s, ng, tg)
	if err != nil {
		lg.Critical("Error instantiating SMS code verifier: %v", err)
		return
	}
	if err != nil {
		lg.Critical("Error instantiating auth module: %s", err)
		return
	}
	oAuthOpts, err := oAuthOptions(conf.Authentication)
	if err != nil {
		lg.Warn("Authentication options: %v", conf.SMS.ActiveAPI, err)
	}
	a, err := model.New(tg, lg, db, pv, oAuthOpts...)
	serverRPCQuitCh := make(chan error)
	serverHttpQuitCh := make(chan error)
	rpcSrv, err := rpc.New(config.CanonicalName, a, lg)
	if err != nil {
		lg.Critical("Error instantiating rpc server module: %s", err)
		return
	}
	go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)
	httpHandler, err := http.NewHandler(a)
	if err != nil {
		lg.Critical("Error instantiating rpc server module: %s", err)
		return
	}
	go serveHttp(conf.Service, httpHandler, serverHttpQuitCh)
	select {
	case err = <-authQuitCh:
		lg.Critical("auth quit with error: %v", err)
	case err = <-serverHttpQuitCh:
		lg.Critical("http server quit with error: %v", err)
	case err = <-serverRPCQuitCh:
		lg.Critical("rpc server quit with error: %v", err)
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

func smsAPIOrStub(conf config.SMSConfig) (verification.SMSer, error) {
	if conf.ActiveAPI == "" {
		return &sms.Stub{}, nil
	}
	var s verification.SMSer
	switch conf.ActiveAPI {
	case config.SMSAPIAfricasTalking:
		apiKey, err := readFile(conf.AfricasTalking.APIKeyFile)
		if err != nil {
			return &sms.Stub{}, fmt.Errorf("africa's talking API key: %v", err)
		}
		s, err = africas_talking.NewSMSCl(conf.AfricasTalking.UserName, apiKey)
		if err != nil {
			return &sms.Stub{}, fmt.Errorf("africasTalking: %v", err)
		}
	case config.SMSAPITwilio:
		tkn, err := readFile(conf.Twilio.TokenKeyFile)
		if err != nil {
			return &sms.Stub{}, fmt.Errorf("twilio token: %v", err)
		}
		s, err = twilio.NewSMSCl(conf.Twilio.ID, tkn, conf.Twilio.SenderPhone)
		if err != nil {
			return &sms.Stub{}, fmt.Errorf("twilio: %v", err)
		}
	default:
		return &sms.Stub{}, fmt.Errorf("invalid API selected can be %s or %s",
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

func serveRPC(conf config.ServiceConfig, rpcSrv *rpc.Server, quitCh chan error) {
	service := micro.NewService(
		micro.Name(config.CanonicalRPCName),
		micro.Version(conf.LoadBalanceVersion),
		micro.RegisterInterval(conf.RegisterInterval),
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
