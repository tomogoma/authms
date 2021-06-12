package main

import (
	"net/http"

	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	httpInternal "github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/logging"
	"github.com/tomogoma/authms/logging/logrus"
	"google.golang.org/appengine"
)

func main() {

	config.DefaultConfDir("conf")
	log := &logrus.Wrapper{}
	boot := bootstrap.Instantiate(config.DefaultConfPath(), log)

	httpHandler, err := httpInternal.NewHandler(
		httpInternal.RequiredParams{Auth: boot.Authentication, Guard: boot.Guard, Logger: log},
		httpInternal.OptionalParams{WebappUrl: boot.Conf.Service.WebAppURL,
			AllowedOrigins: boot.Conf.Service.AllowedOrigins},
	)
	logging.LogFatalOnError(log, err, "Instantiate http Handler")

	http.Handle("/", httpHandler)
	appengine.Main()
}
