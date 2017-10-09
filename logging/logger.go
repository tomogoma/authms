package logging

import (
	"fmt"
	"os"
	"time"
)

type Logger interface {
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
	Fields map[string]interface{}
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

func (lg *EntryLogWrapper) WithFields(f map[string]interface{}) Logger {
	lg.prepare()
	newLG := &EntryLogWrapper{}
	newLG.Fields = lg.Fields
	for k, v := range f {
		newLG.Fields[k] = v
	}
	return newLG
}

func (lg *EntryLogWrapper) WithField(k string, v interface{}) Logger {
	lg.prepare()
	newLG := &EntryLogWrapper{}
	newLG.Fields = lg.Fields
	newLG.Fields[k] = v
	return newLG
}

func (lg *EntryLogWrapper) Infof(f string, args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelInfo,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Warnf(f string, args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelWarn,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Errorf(f string, args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelError,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
}

func (lg *EntryLogWrapper) Fatalf(f string, args ...interface{}) {
	entryLoggerFunc(Entry{
		Level:   LevelFatal,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintf(f, args...),
	})
	os.Exit(1)
}

func (lg *EntryLogWrapper) Info(args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelInfo,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Warn(args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelWarn,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Error(args ...interface{}) {
	go entryLoggerFunc(Entry{
		Level:   LevelError,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
}

func (lg *EntryLogWrapper) Fatal(args ...interface{}) {
	entryLoggerFunc(Entry{
		Level:   LevelFatal,
		Time:    time.Now(),
		Fields:  lg.Fields,
		Payload: fmt.Sprintln(args...),
	})
	os.Exit(1)
}

func (lg *EntryLogWrapper) prepare() {
	if lg.Fields == nil {
		lg.Fields = make(map[string]interface{})
	}
}
