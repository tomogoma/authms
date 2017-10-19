package logging

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type Logger interface {
	WithHTTPRequest(r *http.Request) Logger
	WithFields(map[string]interface{}) Logger
	WithField(string, interface{}) Logger
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

type Entry struct {
	Fields  map[string]interface{}
	Level   string
	Payload string
	Time    time.Time
}

type EntryLoggerFunc func(entry Entry)

type EntryLogger interface {
	Log(Entry)
}

// EntryLogWrapper implements the Logger interface by using an EntryLoggerFunc
// to print out the final log.
// To use an entry logger, add a blank identifier import e.g. for the standard
// logger use:
//     _ "github.com/tomogoma/authms/logging/standard"
// for the purpose of its side effects.
type EntryLogWrapper struct {
	Fields  map[string]interface{}
	HTTPReq *http.Request
}

const (
	LevelInfo  = "Info"
	LevelWarn  = "Warn"
	LevelError = "Error"
	LevelFatal = "Fatal"
)

var entryLoggerFunc EntryLoggerFunc

func SetEntryLoggerFunc(loggerFunc EntryLoggerFunc) {
	entryLoggerFunc = loggerFunc
}

func (lg *EntryLogWrapper) WithHTTPRequest(r *http.Request) Logger {
	newLG := lg.copy()
	newLG.HTTPReq = r
	return newLG
}

func (lg *EntryLogWrapper) WithFields(f map[string]interface{}) Logger {
	newLG := lg.copy()
	for k, v := range f {
		newLG.Fields[k] = v
	}
	return newLG
}

func (lg *EntryLogWrapper) WithField(k string, v interface{}) Logger {
	newLG := lg.copy()
	newLG.Fields[k] = v
	return newLG
}

func (lg *EntryLogWrapper) Infof(f string, args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelInfo,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Warnf(f string, args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelWarn,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Errorf(f string, args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelError,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Fatalf(f string, args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelFatal,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
	os.Exit(1)
}

func (lg *EntryLogWrapper) Info(args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelInfo,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Warn(args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelWarn,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Error(args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelError,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Fatal(args ...interface{}) {
	lg.prepare()
	entryLoggerFunc(Entry{
		Level:   LevelFatal,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
	os.Exit(1)
}

func (lg *EntryLogWrapper) copy() *EntryLogWrapper {
	lg.prepare()
	return &EntryLogWrapper{
		Fields:  lg.Fields,
		HTTPReq: lg.HTTPReq,
	}
}

func (lg *EntryLogWrapper) prepare() {
	if lg.Fields == nil {
		lg.Fields = make(map[string]interface{})
	}
}
