package logging

func LogWarnOnError(lg Logger, err error, action string) {
	if err != nil {
		lg.WithField(FieldAction, action).Warn(err)
	}
}

func LogFatalOnError(lg Logger, err error, action string) {
	if err != nil {
		lg.WithField(FieldAction, action).Fatal(err)
	}
}
