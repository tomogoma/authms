package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/authms/auth/phone/sms"
	"github.com/tomogoma/authms/auth/phone/verification"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/http"
	"github.com/tomogoma/authms/server/rpc"
	"github.com/tomogoma/go-commons/auth/token"
	configH "github.com/tomogoma/go-commons/config"
)

const (
	serviceName    = "authms"
	rpcNamePrefix  = ""
	webnamePrefix  = "go.micro.web."
	serviceVersion = "0.0.1"
)

type defLogWriter struct {
	lg log4go.Logger
}

func (dlw defLogWriter) Write(p []byte) (int, error) {
	dlw.lg.Info("%s", p)
	return len(p), nil
}

var confFile = flag.String("conf", "/etc/authms/authms.conf.yml", "location of config file")

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
	pg, err := password.NewGenerator(password.AllChars)
	if err != nil {
		lg.Critical("Error instantiating password generator: %s", err)
		return
	}
	db, err := dbhelper.New(conf.Database, pg, hash.Hasher{})
	if err != nil {
		lg.Critical("Error instantiating db helper: %v", err)
		return
	}
	var s verification.SMSer
	switch conf.SMS.ActiveAPI {
	case config.SMSAPIAfricasTalking:
		s, err = sms.NewAfricasTalking(conf.SMS.AfricasTalking)
	case config.SMSAPITwilio:
		s, err = sms.NewTwilio(conf.SMS.Twilio)
	default:
		lg.Critical("Invalid SMS API selected can be africasTalking or twilio")
		return
	}
	if err != nil {
		lg.Critical("Error instantiating %s API client: %v",
			conf.SMS.ActiveAPI, err)
		return
	}
	var testMessage string
	if hostName, err := os.Hostname(); err == nil {
		testMessage = fmt.Sprintf("The SMS API is being used on %s", hostName)
	} else {
		testMessage = "The SMS API is being used on an unknown host"
	}
	if err := s.SMS(conf.SMS.TestNumber, testMessage); err != nil {
		lg.Critical("Error sending test SMS during start up: %v", err)
		return
	}
	ng, err := password.NewGenerator(password.NumberChars)
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
	rpcSrv, err := rpc.New(serviceName, a, lg)
	if err != nil {
		lg.Critical("Error instantiating rpc server module: %s", err)
		return
	}
	go serveRPC(conf.Service.RegisterInterval, rpcSrv, serverRPCQuitCh)
	httpHandler, err := http.New(a)
	if err != nil {
		lg.Critical("Error instantiating rpc server module: %s", err)
		return
	}
	go serveHttp(conf.Service.RegisterInterval, httpHandler, serverHttpQuitCh)
	select {
	case err = <-authQuitCh:
		lg.Critical("auth quit with error: %v", err)
	case err = <-serverHttpQuitCh:
		lg.Critical("http server quit with error: %v", err)
	case err = <-serverRPCQuitCh:
		lg.Critical("rpc server quit with error: %v", err)
	}
}

func serveRPC(regInterval time.Duration, rpcSrv *rpc.Server, quitCh chan error) {
	service := micro.NewService(
		micro.Name(rpcNamePrefix+serviceName),
		micro.Version(serviceVersion),
		micro.RegisterInterval(regInterval),
	)
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
}

type RouteHandler interface {
	HandleRoute(r *mux.Router)
}

func serveHttp(regInterval time.Duration, rh RouteHandler, quitCh chan error) {
	r := mux.NewRouter()
	rh.HandleRoute(r)
	service := web.NewService(
		web.Handler(r),
		web.Name(webnamePrefix+serviceName),
		web.Version(serviceVersion),
		web.RegisterInterval(regInterval),
	)
	quitCh <- service.Run()
}
