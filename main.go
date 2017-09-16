package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/mux"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/sms"
	"github.com/tomogoma/authms/auth/phone/verification"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/generator"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/http"
	"github.com/tomogoma/authms/server/rpc"
	"github.com/tomogoma/authms/store"
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
	conf := config.Config{}
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
	db, err := store.NewRoach(conf.Database, pg, hash.Hasher{})
	if err != nil {
		lg.Critical("Error instantiating db helper: %v", err)
		return
	}
	if err := db.InitDBConnIfNotInitted(); err != nil {
		lg.Warn("Error initiating connection to db: %v", err)
	}
	var s verification.SMSer
	if conf.SMS.ActiveAPI != "" {
		switch conf.SMS.ActiveAPI {
		case config.SMSAPIAfricasTalking:
			s, err = sms.NewAfricasTalking(conf.SMS.AfricasTalking)
		case config.SMSAPITwilio:
			s, err = sms.NewTwilio(conf.SMS.Twilio)
		default:
			lg.Critical("Invalid SMS API selected can be africasTalking or twilio")
			return
		}
		var testMessage string
		if hostName, err := os.Hostname(); err == nil {
			testMessage = fmt.Sprintf("The SMS API is being used on %s", hostName)
		} else {
			testMessage = "The SMS API is being used on an unknown host"
		}
		if err := s.SMS(conf.SMS.TestNumber, testMessage); err != nil {
			lg.Warn("Error sending test SMS during start up: %v", err)
		}
	} else {
		s = &sms.Stub{}
	}
	if err != nil {
		lg.Critical("Error instantiating %s API client: %v",
			conf.SMS.ActiveAPI, err)
		return
	}
	ng, err := generator.NewRandom(generator.NumberChars)
	if err != nil {
		lg.Critical("Error instantiating number generator: %s", err)
		return
	}
	pv, err := verification.New(conf.SMS.Verification, s, ng, tg)
	if err != nil {
		lg.Critical("Error instantiating SMS code verifier: %v", err)
		return
	}
	oa, err := oauth.New(conf.OAuth)
	if err != nil {
		lg.Critical("Error instantiating OAuth module: %s", err)
		return
	}
	a, err := auth.New(tg, lg, db, oa, pv)
	if err != nil {
		lg.Critical("Error instantiating auth module: %s", err)
		return
	}
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
