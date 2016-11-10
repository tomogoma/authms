package main

import (
	"log"
	"time"

	"github.com/limetext/log4go"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/authms/server/rpc"
)

const (
	serviceName    = "AuthMS"
	serviceVersion = "0.0.1"
)

type defLogWriter struct {
	lg log4go.Logger
}

func (dlw defLogWriter) Write(p []byte) (int, error) {
	return len(p), dlw.lg.Error("%s", p)
}

func main() {
	conf := Config{
		RunType:          runTypeRPC,
		RegisterInterval: 5 * time.Second,
		HttpAddress:      ":3345",
		Database: helper.DSN{
			UName:       "root",
			Host:        "localhost:26257",
			DB:          "authms",
			SslCert:     "/etc/cockroachdb/certs/node.cert",
			SslKey:      "/etc/cockroachdb/certs/node.key",
			SslRootCert: "/etc/cockroachdb/certs/ca.cert",
		},
		Authentication: auth.Config{
			BlackListFailCount: auth.MinBlackListFailCount,
			BlacklistWindow:    auth.MinBlackListWindow,
		},
		Token: token.Config{
			TokenKeyFile: "ssh-keys/sha256.key",
		},
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
	switch conf.RunType {
	case runTypeRPC:
		serveRPC(conf, rpcSrv, serverRPCQuitCh)
	case runTypeHttp:
		serveHttp(rpcSrv, serverHttpQuitCh)
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

func serveRPC(c Config, rpcSrv *rpc.Server, quitCh chan error) {
	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(serviceVersion),
		micro.RegisterInterval(c.RegisterInterval),
	)
	service.Init()
	authms.RegisterAuthMSHandler(service.Server(), rpcSrv)
	go func() {
		if err := service.Run(); err != nil {
			quitCh <- err
		}
	}()
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
	go func() {
		if err := server.Run(); err != nil {
			quitCh <- err
		}
	}()
}
