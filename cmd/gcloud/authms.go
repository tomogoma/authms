package gcloud

import (
	"net/http"

	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	httpInternal "github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/logging"
)

func init() {
	config.DefaultConfDir("conf")
	_, authentication, APIGuard, _, _, _, _ := bootstrap.Instantiate(config.DefaultConfPath())

	httpHandler, err := httpInternal.NewHandler(authentication, APIGuard)
	logging.LogFatalOnError(err, "Instantiating http Handler")

	http.Handle("/", httpHandler)
}
