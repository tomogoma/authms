package logging

type Logger interface {
	WithFields(map[string]interface{}) Logger
	WithField(string, interface{}) Logger
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
}
