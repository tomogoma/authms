package main

import (
	"flag"
	"log"
	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/rpc"
	"runtime"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/authms/auth/dbhelper"
	"github.com/tomogoma/authms/auth/hash"
	"github.com/tomogoma/authms/auth/phone/verification"
	"github.com/tomogoma/authms/auth/phone/sms"
)

const (
	serviceName = "authms"
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
	conf, err := config.ReadFile(*confFile)
	if err != nil {
		log.Fatalf("Error Reading config file: %s", err)
	}
	lg := log4go.NewDefaultLogger(log4go.FINEST)
	log.SetOutput(defLogWriter{lg: lg})
	defer runtime.Gosched()
	tg, err := token.NewGenerator(conf.Token)
	if err != nil {
		lg.Critical("Error instantiating token generator: %s", err)
		return
	}
	pg, err := password.NewGenerator(password.AllChars)
	if err != nil {
		lg.Critical("Error instantiating token generator: %s", err)
		return
	}
	authQuitCh := make(chan error)
	oa, err := oauth.New(conf.OAuth)
	if err != nil {
		lg.Critical("Error instantiating OAuth module: %s", err)
		return
	}
	db, err := dbhelper.New(conf.Database, pg, hash.Hasher{})
	if err != nil {
		lg.Critical("Error instantiating db helper: %v", err)
		return
	}
	s, err := sms.New(conf.Twilio)
	if err != nil {
		lg.Critical("Error instantiating SMS API client: %v", err)
		return
	}
	pv, err := verification.New(conf.Twilio, s, pg, tg)
	if err != nil {
		lg.Critical("Error instantiating SMS code verifier: %v", err)
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
	switch conf.Service.RunType {
	case config.RunTypeRPC:
		go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)
	case config.RunTypeHttp:
		go serveHttp(rpcSrv, serverHttpQuitCh)
	default:
		lg.Critical("Invalid runt type chosen")
		return
	}
	select {
	case err = <-authQuitCh:
		lg.Critical("auth quit with error: %v", err)
	case err = <-serverHttpQuitCh:
		lg.Critical("http server quit with error: %v", err)
	case err = <-serverRPCQuitCh:
		lg.Critical("rpc server quit with error: %v", err)
	}
}

func serveRPC(c config.ServiceConfig, rpcSrv *rpc.Server, quitCh chan error) {
	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(serviceVersion),
		micro.RegisterInterval(c.RegisterInterval),
	)
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
}

func serveHttp(rpcServ *rpc.Server, quitCh chan error) {
	server.Init(
		server.Name(serviceName),
		server.Version(serviceVersion),
		//server.RegisterInterval(c.RegisterInterval),
	)
	server.Handle(server.NewHandler(rpcServ))
	err := server.Run()
	quitCh <- err
}
