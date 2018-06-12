package log

import (
	"github.com/sirupsen/logrus"
)

const (
	LevelError = "error"
	LevelWarn  = "warn"
	LevelDebug = "debug"

	logFieldErrorCause = "rootError"
)

var (
	logger = logrus.New()
)

func init() {
	//logger.Formatter = &logrus.JSONFormatter{}
}

func Get(level string) logrus.Logger {
	switch level {
	case LevelError:
		logger.Level = logrus.ErrorLevel
	case LevelWarn:
		logger.Level = logrus.WarnLevel
	case LevelDebug:
		logger.Level = logrus.DebugLevel
	default:
		logger.Level = logrus.InfoLevel
	}
	return *logger
}

func Error(description string, err error, logger logrus.Logger) {
	logger.WithFields(logrus.Fields{
		logFieldErrorCause: err}).Error(description)
}