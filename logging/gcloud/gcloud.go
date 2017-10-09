package gcloud

import (
	"context"
	"encoding/json"
	"fmt"

	gcLogging "cloud.google.com/go/logging"
	"github.com/tomogoma/authms/logging"
)

type Logger struct {
	projectID  string
	loggerName string
	errors     []error
}

const LabelLogError = "logError"

var logger *Logger

func init() {
	logger = &Logger{errors: make([]error, 0)}
	logging.SetEntryLoggerFunc(logger.Log)
}

func SetProject(projectID, loggerName string) *Logger {
	logger.projectID = projectID
	logger.loggerName = loggerName
	return logger
}

func (lg *Logger) Log(e logging.Entry) {
	cl, err := gcLogging.NewClient(context.Background(), lg.projectID)
	if err != nil {
		err = fmt.Errorf("instantiate gcloud logger: %v", err)
		lg.errors = append(lg.errors, err)
		return
	}
	defer func() {
		if err := cl.Close(); err != nil {
			err = fmt.Errorf("closing log client and flushing buffer: %v", err)
			lg.errors = append(lg.errors, err)
		}
		lg.errors = make([]error, 0)
	}()

	log := cl.Logger(lg.loggerName)
	lg.dumpErrors(log)

	labels := make(map[string]string)
	for k, v := range e.Fields {
		label, err := json.Marshal(v)
		if err != nil {
			err = fmt.Errorf("marshal log fields: %v", err)
			lg.errors = append(lg.errors, err)
			return
		}
		labels[k] = string(label)
	}

	log.Log(gcLogging.Entry{
		Payload: e.Payload,
		Labels:  labels,
	})
}

func (lg *Logger) dumpErrors(log *gcLogging.Logger) {
	for _, err := range lg.errors {
		log.Log(gcLogging.Entry{
			Payload: err.Error(),
			Labels:  map[string]string{LabelLogError: ""},
		})
	}
}
