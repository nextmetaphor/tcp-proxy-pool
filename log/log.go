package log

import (
	"github.com/sirupsen/logrus"
)

const (
	// LevelError specifies the log level for error logging
	LevelError = "error"
	// LevelWarning specifies the log level for warning logging
	LevelWarning = "warn"
	// LevelDebug specifies the log level for debug logging
	LevelDebug   = "debug"

	logFieldErrorCause = "rootError"
)

var (
	logger = logrus.New()
)

func init() {
	//logger.Formatter = &logrus.JSONFormatter{}
}

// Get returns a pointer to a logger with the specified level
func Get(level string) *logrus.Logger {
	switch level {
	case LevelError:
		logger.Level = logrus.ErrorLevel
	case LevelWarning:
		logger.Level = logrus.WarnLevel
	case LevelDebug:
		logger.Level = logrus.DebugLevel
	default:
		logger.Level = logrus.InfoLevel
	}
	return logger
}

// Error is a helper function which logs an error message with the specified logger
func Error(description string, err error, logger *logrus.Logger) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			logFieldErrorCause: err}).Error(description)
	}
}