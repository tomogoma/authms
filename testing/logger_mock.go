package testing

import "github.com/tomogoma/authms/logging"

type Entry struct {
	Level string
	Fmt   string
	Args  []interface{}
}

type LoggerMock struct {
	Fields map[string]interface{}
	Logs   []Entry
}

const (
	LevelInfo  = "Info"
	LevelWarn  = "Warn"
	LevelError = "Error"
)

func (lg *LoggerMock) WithFields(f map[string]interface{}) logging.Logger {
	newLG := &LoggerMock{Fields: lg.Fields}
	if newLG.Fields == nil {
		newLG.Fields = make(map[string]interface{})
	}
	for k, v := range f {
		newLG.Fields[k] = v
	}
	return newLG
}

func (lg *LoggerMock) WithField(k string, v interface{}) logging.Logger {
	newLG := &LoggerMock{Fields: lg.Fields}
	if newLG.Fields == nil {
		newLG.Fields = make(map[string]interface{})
	}
	newLG.Fields[k] = v
	return newLG
}

func (lg *LoggerMock) Infof(fmt string, args ...interface{}) {
	if lg.Logs == nil {
		lg.Logs = make([]Entry, 0)
	}
	lg.Logs = append(lg.Logs, Entry{Level: LevelInfo, Fmt: fmt, Args: args})
}

func (lg *LoggerMock) Warnf(fmt string, args ...interface{}) {
	if lg.Logs == nil {
		lg.Logs = make([]Entry, 0)
	}
	lg.Logs = append(lg.Logs, Entry{Level: LevelWarn, Fmt: fmt, Args: args})
}

func (lg *LoggerMock) Errorf(fmt string, args ...interface{}) {
	if lg.Logs == nil {
		lg.Logs = make([]Entry, 0)
	}
	lg.Logs = append(lg.Logs, Entry{Level: LevelError, Fmt: fmt, Args: args})
}
