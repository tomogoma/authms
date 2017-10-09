package gcloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gcLogging "cloud.google.com/go/logging"
	"github.com/tomogoma/authms/logging"
	"google.golang.org/appengine"
)

type Logger struct {
	projectID  string
	loggerName string
	pending    []logging.Entry
}

var logger *Logger

func init() {
	logger = &Logger{pending: make([]logging.Entry, 0)}
	logging.SetEntryLoggerFunc(logger.Log)
}

func SetProject(projectID, loggerName string) *Logger {
	logger.projectID = projectID
	logger.loggerName = loggerName
	return logger
}

func (lg *Logger) Log(e logging.Entry) {
	reqObjI, exists := e.Fields[logging.FieldHttpReqObj]
	if !exists {
		lg.pending = append(lg.pending, e)
		return
	}
	cl, err := gcLogging.NewClient(appengine.NewContext(reqObjI.(*http.Request)), lg.projectID)
	if err != nil {
		fmt.Printf("instantiate gcloud logging client: %v\n", err)
		return
	}
	defer func() {
		if err := cl.Close(); err != nil {
			lg.pending = append(lg.pending, logging.Entry{
				Time:    time.Now(),
				Level:   logging.LevelError,
				Payload: err.Error(),
				Fields: map[string]interface{}{
					logging.FieldAction: "close gcloud logger & flush buffer",
				},
			})
		}
		lg.pending = make([]logging.Entry, 0)
	}()

	logger := cl.Logger(lg.loggerName)
	lg.dumpPending(logger)

	logger.Log(gcLogging.Entry{
		Payload: e.Payload,
		Labels:  lg.fieldsToLabels(e.Fields),
	})
}

func (lg *Logger) dumpPending(log *gcLogging.Logger) {
	for _, p := range lg.pending {
		log.Log(gcLogging.Entry{
			Payload: p.Payload,
			Labels:  lg.fieldsToLabels(p.Fields),
		})
	}
}

func (lg *Logger) fieldsToLabels(fs map[string]interface{}) map[string]string {
	labels := make(map[string]string)
	for k, v := range fs {
		label, err := json.Marshal(v)
		if err != nil {
			labels[k] = fmt.Sprintf("%s", v)
		} else {
			labels[k] = string(label)
		}
	}
	return labels
}
