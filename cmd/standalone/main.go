package main

import (
	"flag"
	"net/http"
	"net/url"

	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	httpInternal "github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/logging/logrus"
)

var confPath = flag.String("conf", config.DefaultConfPath(), "/path/to/config_file.yml")

func main() {

	flag.Parse()

	logWrapper := &logrus.Wrapper{}
	conf, authentication, APIGuard, _, _, _, _ := bootstrap.Instantiate(*confPath, logWrapper)

	listenNSrvLg := logWrapper.WithField(logging.FieldAction, "Listen and serve")

	srvURL, err := url.Parse(conf.Service.URL)
	logging.LogFatalOnError(listenNSrvLg, err, "parse service URL")
	port := ":" + srvURL.Port()
	if len(port) == 1 {
		port = ":8080"
	}

	listenNSrvLg.Infof("Will listen on '%s'", port)

	httpHandler, err := httpInternal.NewHandler(authentication, APIGuard, listenNSrvLg,
		conf.Service.WebAppURL, conf.Service.AllowedOrigins)
	logging.LogFatalOnError(listenNSrvLg, err, "Instantiate http Handler")

	logging.LogFatalOnError(
		listenNSrvLg,
		http.ListenAndServe(port, httpHandler),
		"Run server",
	)

}
