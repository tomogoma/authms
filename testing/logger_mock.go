package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/tomogoma/authms/logging"
)

type Entry struct {
	Level string
	Fmt   *string // nil if logging without formatting e.g. Error() instead of Errorf()
	Args  []interface{}
}

type LoggerMock struct {
	Fields   map[string]interface{}
	Logs     []Entry
	Spinoffs []*LoggerMock
	HTTPReq  *http.Request
}

const (
	LevelInfo  = "Info"
	LevelWarn  = "Warn"
	LevelError = "Error"
	LevelFatal = "Fatal"
)

func (lg *LoggerMock) WithFields(f map[string]interface{}) logging.Logger {
	newLG := lg.copy()
	for k, v := range f {
		newLG.Fields[k] = v
	}
	return newLG
}

func (lg *LoggerMock) WithField(k string, v interface{}) logging.Logger {
	newLG := lg.copy()
	newLG.Fields[k] = v
	return newLG
}

func (lg *LoggerMock) WithHTTPRequest(r *http.Request) logging.Logger {
	newLG := lg.copy()
	newLG.HTTPReq = r
	return newLG
}

func (lg *LoggerMock) Infof(fmt string, args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelInfo, Fmt: &fmt, Args: args})
}

func (lg *LoggerMock) Warnf(fmt string, args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelWarn, Fmt: &fmt, Args: args})
}

func (lg *LoggerMock) Errorf(fmt string, args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelError, Fmt: &fmt, Args: args})
}

func (lg *LoggerMock) Info(args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelInfo, Args: args})
}

func (lg *LoggerMock) Warn(args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelWarn, Args: args})
}

func (lg *LoggerMock) Error(args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelError, Args: args})
}

func (lg *LoggerMock) Fatal(args ...interface{}) {
	lg.prep()
	lg.Logs = append(lg.Logs, Entry{Level: LevelFatal, Args: args})
}

func (lg *LoggerMock) copy() *LoggerMock {
	lg.prep()
	newLG := &LoggerMock{Fields: lg.Fields}
	lg.Spinoffs = append(lg.Spinoffs, newLG)
	return newLG
}

func (lg *LoggerMock) prep() {
	if lg.Logs == nil {
		lg.Logs = make([]Entry, 0)
	}
	if lg.Fields == nil {
		lg.Fields = make(map[string]interface{})
	}
	if lg.Spinoffs == nil {
		lg.Spinoffs = make([]*LoggerMock, 0)
	}
}

func (lg *LoggerMock) PrintLogs(t *testing.T) {
	for _, so := range lg.Spinoffs {
		so.PrintLogs(t)
	}
	for _, e := range lg.Logs {
		var payload string
		if e.Fmt != nil {
			payload = fmt.Sprintf(*e.Fmt, e.Args...)
		} else {
			payload = fmt.Sprint(e.Args...)
		}
		t.Logf("%s: %s\n", e.Level, payload)
	}
}
