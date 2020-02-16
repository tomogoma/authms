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

	log := &logrus.Wrapper{}
	conf, authentication, APIGuard, _, _, _, _ := bootstrap.Instantiate(*confPath, log)

	httpHandler, err := httpInternal.NewHandler(authentication, APIGuard, log,
		conf.Service.WebAppURL, conf.Service.AllowedOrigins)
	logging.LogFatalOnError(log, err, "Instantiate http Handler")

	srvURL, err := url.Parse(conf.Service.URL)
	logging.LogFatalOnError(log, err, "parse service URL")

	port := ":" + srvURL.Port()
	if len(port) == 1 {
		port = ":8080"
	}
	logging.LogFatalOnError(
		log,
		http.ListenAndServe(port, httpHandler),
		"Run server",
	)

}
