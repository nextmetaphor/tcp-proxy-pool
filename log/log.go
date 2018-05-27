package log

import (
	"github.com/sirupsen/logrus"
)

const (
	LevelError = "error"
	LevelWarn  = "warn"
	LevelDebug = "debug"
)

var (
	logger = logrus.New()
)

func init() {
	//logger.Formatter = &logrus.JSONFormatter{}
}

func Get(level string) *logrus.Logger {
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
	return logger
}

func LogError(err error) {
	//ctx.Log.WithFields(logrus.Fields{
	//	"errCode": errCode}).Warn("handleSelectNodeRequest: Error when encoding returned Node")

}