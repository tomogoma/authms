package main

import (
	"log"

	"time"

	"bitbucket.org/tomogoma/auth-ms/auth"
	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/history"
	"bitbucket.org/tomogoma/auth-ms/server/http"
	"github.com/limetext/log4go"
)

type defLogWriter struct {
	lg log4go.Logger
}

func (dlw defLogWriter) Write(p []byte) (n int, err error) {
	return len(p), dlw.lg.Error("%s", p)
}

func main() {

	lg := log4go.NewDefaultLogger(log4go.FINEST)
	log.SetOutput(defLogWriter{lg: lg})
	defer time.Sleep(5 * time.Second)

	dsn := helper.DSN{
		UName:       "root",
		Host:        "z500:26257",
		DB:          "authms",
		SslCert:     "/etc/cockroachdb/certs/node.cert",
		SslKey:      "/etc/cockroachdb/certs/node.key",
		SslRootCert: "/etc/cockroachdb/certs/ca.cert",
	}

	db, err := helper.SQLDB(dsn)
	if err != nil {
		lg.Critical("Error connecting to db: %s", err)
		return
	}

	hm, err := history.NewModel(db)
	if err != nil {
		lg.Critical("Error instantiating history model: %s", err)
		return
	}

	authQuitCh := make(chan error)
	aConf := auth.Config{
		BlackListFailCount: auth.MinBlackListFailCount,
		BlacklistWindow:    auth.MinBlackListWindow,
	}

	a, err := auth.New(dsn, hm, aConf, lg, authQuitCh)
	if err != nil {
		lg.Critical("Error instantiating auth module: %s", err)
		return
	}

	httpQuitCh := make(chan error)
	httpSrv, err := http.New(a, lg, httpQuitCh)
	if err != nil {
		lg.Critical("Error instantiating http server module: %s", err)
		return
	}

	go httpSrv.Start(":3345")

	select {
	case err = <-authQuitCh:
		lg.Critical("auth quit with error: %s", err)
	case err = <-httpQuitCh:
		lg.Critical("http server quit with error: %s", err)
	}
}
