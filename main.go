package main

import (
	"log"

	"time"

	"github.com/limetext/log4go"
	"github.com/tomogoma/authms/auth"
	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/history"
	"github.com/tomogoma/authms/auth/model/token"
	"github.com/tomogoma/authms/server/http"
)

type defLogWriter struct {
	lg log4go.Logger
}

func (dlw defLogWriter) Write(p []byte) (int, error) {
	return len(p), dlw.lg.Error("%s", p)
}

func main() {

	lg := log4go.NewDefaultLogger(log4go.FINEST)
	log.SetOutput(defLogWriter{lg: lg})
	defer time.Sleep(5 * time.Second)

	dsn := helper.DSN{
		UName:       "root",
		Host:        "localhost:26257",
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
		DBName:             dsn.DB,
	}

	tgConf := token.Config{
		TokenKeyFile: "ssh-keys/sha256.key",
	}
	tg, err := token.NewGenerator(tgConf)

	a, err := auth.New(db, hm, tg, aConf, lg, authQuitCh)
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
