package main

import (
	"flag"
	http2 "net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/authms/api"
	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/handler/rpc"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/logging/logrus"
	_ "github.com/tomogoma/authms/logging/standard"
)

func main() {

	confFile := flag.String("conf", config.DefaultConfPath(), "location of config file")
	flag.Parse()
	log := &logrus.Wrapper{}
	boot := bootstrap.Instantiate(*confFile, log)

	serverRPCQuitCh := make(chan error)
	rpcSrv, err := rpc.NewHandler(boot.Guard, boot.Authentication)
	logging.LogFatalOnError(log, err, "Instantate RPC handler")
	go serveRPC(boot.Conf.Service, rpcSrv, serverRPCQuitCh)

	serverHttpQuitCh := make(chan error)
	httpHandler, err := http.NewHandler(
		http.RequiredParams{Auth: boot.Authentication, Guard: boot.Guard, Logger: log},
		http.OptionalParams{WebappUrl: boot.Conf.Service.WebAppURL,
			AllowedOrigins: boot.Conf.Service.AllowedOrigins},
	)
	logging.LogFatalOnError(log, err, "Instantiate HTTP handler")
	go serveHttp(boot.Conf.Service, httpHandler, serverHttpQuitCh)

	select {
	case err = <-serverHttpQuitCh:
		logging.LogFatalOnError(log, err, "Serve HTTP")
	case err = <-serverRPCQuitCh:
		logging.LogFatalOnError(log, err, "Serve RPC")
	}
}

func serveRPC(conf config.Service, usrsH *rpc.UsersHandler, quitCh chan error) {
	service := micro.NewService(
		micro.Name(config.CanonicalRPCName()),
		micro.Version(conf.LoadBalanceVersion),
		micro.RegisterInterval(conf.RegisterInterval),
		micro.WrapHandler(usrsH.Wrapper),
	)
	api.RegisterUsersHandler(service.Server(), usrsH)
	quitCh <- service.Run()
}

func serveHttp(conf config.Service, h http2.Handler, quitCh chan error) {
	srvc := web.NewService(
		web.Handler(h),
		web.Name(config.CanonicalWebName()),
		web.Version(conf.LoadBalanceVersion),
		web.RegisterInterval(conf.RegisterInterval),
	)
	quitCh <- srvc.Run()
}
