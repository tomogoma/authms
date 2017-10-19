package logrus

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tomogoma/authms/logging"
)

type Wrapper struct {
	Entry   *logrus.Entry
	HTTPReq *http.Request
}

func (lg *Wrapper) WithFields(f map[string]interface{}) logging.Logger {
	newLg := lg.copy()
	newLg.Entry = newLg.Entry.WithFields(f)
	return newLg
}

func (lg *Wrapper) WithField(k string, v interface{}) logging.Logger {
	newLg := lg.copy()
	newLg.Entry = newLg.Entry.WithField(k, v)
	return newLg
}

func (lg *Wrapper) WithHTTPRequest(r *http.Request) logging.Logger {
	newLg := lg.copy()
	newLg.HTTPReq = r
	return newLg
}

func (lg *Wrapper) Infof(f string, args ...interface{}) {
	lg.prepare()
	lg.Entry.Infof(f, args...)
}

func (lg *Wrapper) Warnf(f string, args ...interface{}) {
	lg.prepare()
	lg.Entry.Warnf(f, args...)
}

func (lg *Wrapper) Errorf(f string, args ...interface{}) {
	lg.prepare()
	lg.Entry.Errorf(f, args...)
}

func (lg *Wrapper) Fatalf(f string, args ...interface{}) {
	lg.prepare()
	lg.Entry.Fatalf(f, args...)
}

func (lg *Wrapper) Info(args ...interface{}) {
	lg.prepare()
	lg.Entry.Info(args...)
}

func (lg *Wrapper) Warn(args ...interface{}) {
	lg.prepare()
	lg.Entry.Warn(args...)
}

func (lg *Wrapper) Error(args ...interface{}) {
	lg.prepare()
	lg.Entry.Error(args...)
}

func (lg *Wrapper) Fatal(args ...interface{}) {
	lg.prepare()
	lg.Entry.Fatal(args...)
}

func (lg *Wrapper) copy() *Wrapper {
	lg.prepare()
	return &Wrapper{
		Entry:   lg.Entry,
		HTTPReq: lg.HTTPReq,
	}
}

func (lg *Wrapper) prepare() {
	if lg.Entry == nil {
		lg.Entry = logrus.NewEntry(logrus.New())
	}
}
