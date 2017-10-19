package gcloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/tomogoma/authms/logging"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type logFunc func(context.Context, string, ...interface{})

type Entry struct {
	value    string
	callFunc logFunc
}

type Logger struct {
	Fields  map[string]interface{}
	pending []Entry
	HTTPReq *http.Request
}

func (lg *Logger) WithFields(f map[string]interface{}) logging.Logger {
	newLG := lg.copy()
	for k, v := range f {
		newLG.Fields[k] = v
	}
	return newLG
}

func (lg *Logger) WithField(k string, v interface{}) logging.Logger {
	newLG := lg.copy()
	newLG.Fields[k] = v
	return newLG
}

func (lg *Logger) WithHTTPRequest(r *http.Request) logging.Logger {
	newLg := lg.copy()
	newLg.HTTPReq = r
	return newLg
}

func (lg *Logger) Infof(f string, args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprintf(f, args...), log.Infof)
}

func (lg *Logger) Warnf(f string, args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprintf(f, args...), log.Warningf)
}

func (lg *Logger) Errorf(f string, args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprintf(f, args...), log.Errorf)
}

func (lg *Logger) Fatalf(f string, args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprintf(f, args...), log.Criticalf)
	os.Exit(1)
}

func (lg *Logger) Info(args ...interface{}) {
	lg.log(fmt.Sprint(args...), log.Infof)
}

func (lg *Logger) Warn(args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprint(args...), log.Warningf)
}

func (lg *Logger) Error(args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprint(args...), log.Errorf)
}

func (lg *Logger) Fatal(args ...interface{}) {
	lg.prepare()
	lg.log(fmt.Sprint(args...), log.Criticalf)
	os.Exit(1)
}

func (lg *Logger) copy() *Logger {
	lg.prepare()
	newLG := &Logger{
		Fields:  lg.Fields,
		HTTPReq: lg.HTTPReq,
		pending: lg.pending,
	}
	lg.pending = make([]Entry, 0)
	return newLG
}

func (lg *Logger) prepare() {
	if lg.Fields == nil {
		lg.Fields = make(map[string]interface{})
	}
}

func (lg *Logger) log(payload string, f logFunc) {

	lg.prepare()

	delete(lg.Fields, logging.FieldHost)
	delete(lg.Fields, logging.FieldRequestHandler)
	delete(lg.Fields, logging.FieldMethod)
	delete(lg.Fields, logging.FieldURL)
	fields, _ := json.Marshal(lg.Fields)
	val := fmt.Sprintf("%s %s\n", payload, fields)

	if lg.HTTPReq == nil {
		lg.pending = append(lg.pending, Entry{callFunc: f, value: val})
		return
	}
	ctx := appengine.NewContext(lg.HTTPReq)

	for _, pending := range lg.pending {
		log.Infof(ctx, "***Begin pending logs***")
		pending.callFunc(ctx, pending.value)
		log.Infof(ctx, "***End pending logs***")
	}
	lg.pending = make([]Entry, 0)

	f(ctx, val)
}
