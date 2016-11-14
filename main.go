package main

import (
	"flag"
	"log"
	"time"

	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/token"
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
	return len(p), dlw.lg.Error("%s", p)
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
	defer time.Sleep(5 * time.Second)
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
	authQuitCh := make(chan error)
	a, err := auth.New(db, hm, tg, conf.Authentication, lg, authQuitCh)
	if err != nil {
		lg.Critical("Error instantiating auth module: %s", err)
		return
	}
	serverRPCQuitCh := make(chan error)
	serverHttpQuitCh := make(chan error)
	rpcSrv, err := rpc.New(a, lg)
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
		lg.Critical("auth quit with error: %s", err)
	case err = <-serverHttpQuitCh:
		lg.Critical("http server quit with error: %s", err)
	case err = <-serverRPCQuitCh:
		lg.Critical("rpc server quit with error: %s", err)
	}
}

func serveRPC(c config.ServiceConfig, rpcSrv *rpc.Server, quitCh chan error) {
	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(serviceVersion),
		micro.RegisterInterval(c.RegisterInterval),
	)
	service.Init()
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
}

func serveHttp(rpcServ *rpc.Server, quitCh chan error) {
	//httpSrv, err := http.New(a, lg, quitCh)
	//if err != nil {
	//	lg.Critical("Error instantiating http server module: %s", err)
	//	return
	//}
	//go httpSrv.Start(c.HttpAddress)
	server.Init(
		server.Name(serviceName),
		server.Version(serviceVersion),
		//server.RegisterInterval(c.RegisterInterval),
	)
	server.Handle(server.NewHandler(rpcServ))
	err := server.Run()
	quitCh <- err
}
