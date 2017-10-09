package gcloud

import (
	"net/http"

	"github.com/tomogoma/authms/bootstrap"
	"github.com/tomogoma/authms/config"
	httpInternal "github.com/tomogoma/authms/handler/http"
	"github.com/tomogoma/authms/logging"
	_ "github.com/tomogoma/authms/logging/standard"
)

func init() {

	config.DefaultConfDir("conf")
	log := &logging.EntryLogWrapper{}
	_, authentication, APIGuard, _, _, _, _ := bootstrap.Instantiate(config.DefaultConfPath(), log)

	httpHandler, err := httpInternal.NewHandler(authentication, APIGuard, log)
	logging.LogFatalOnError(log, err, "Instantiating http Handler")

	http.Handle("/", httpHandler)
}
