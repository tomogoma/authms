package main

import (
	"flag"
	http2 "net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-web"
	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	"github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/handler/rpc"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/logging/logrus"
	_ "github.com/tomogoma/authms/logging/standard"
	"github.com/tomogoma/authms/proto/authms"
)

func main() {

	confFile := flag.String("conf", config.DefaultConfPath(), "location of config file")
	flag.Parse()
	log := &logrus.Wrapper{}
	conf, authentication, APIGuard, _, _, _, _ := bootstrap.Instantiate(*confFile, log)

	serverRPCQuitCh := make(chan error)
	rpcSrv, err := rpc.NewHandler(config.CanonicalName(), authentication)
	logging.LogFatalOnError(log, err, "Instantate RPC handler")
	go serveRPC(conf.Service, rpcSrv, serverRPCQuitCh)

	serverHttpQuitCh := make(chan error)
	httpHandler, err := http.NewHandler(authentication, APIGuard, log,
		conf.Service.AllowedOrigins)
	logging.LogFatalOnError(log, err, "Instantiate HTTP handler")
	go serveHttp(conf.Service, httpHandler, serverHttpQuitCh)

	select {
	case err = <-serverHttpQuitCh:
		logging.LogFatalOnError(log, err, "Serve HTTP")
	case err = <-serverRPCQuitCh:
		logging.LogFatalOnError(log, err, "Serve RPC")
	}
}

func serveRPC(conf config.Service, rpcSrv *rpc.Handler, quitCh chan error) {
	service := micro.NewService(
		micro.Name(config.CanonicalRPCName()),
		micro.Version(conf.LoadBalanceVersion),
		micro.RegisterInterval(conf.RegisterInterval),
		micro.WrapHandler(rpcSrv.Wrapper),
	)
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	err := service.Run()
	quitCh <- err
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
