package logging

import "github.com/sirupsen/logrus"

func LogWarnOnError(err error, action string) {
	if err != nil {
		logrus.WithField(FieldAction, action).Warn(err)
	}
}

func LogFatalOnError(err error, action string) {
	if err != nil {
		logrus.WithField(FieldAction, action).Fatal(err)
	}
}
