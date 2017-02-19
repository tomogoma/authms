package main

import (
	"flag"
	"log"
	"time"

	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/dbhelper/helper"
	"github.com/tomogoma/authms/auth/dbhelper/history"
	"github.com/tomogoma/authms/auth/dbhelper/token"
	"github.com/tomogoma/authms/auth/oauth"
	"github.com/tomogoma/authms/auth/password"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/rpc"
)

const (
	serviceName    = "authms"
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
	defer time.Sleep(500 * time.Millisecond)
	db, err := helper.SQLDB(conf.Database)
	if err != nil {
		lg.Critical("Error connecting to db: %s", err)
		return
	}
	hm, err := history.NewModel(db)
	if err != nil {
		lg.Critical("Error instantiating history model: %s", err)
		return
	}
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
	a, err := auth.New(db, hm, tg, pg, conf.Authentication, lg, oa, authQuitCh)
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
		micro.Metadata(map[string]string{
			"type": "helloworld",
		}),
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
