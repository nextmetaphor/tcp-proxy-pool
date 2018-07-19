package log

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/sirupsen/logrus/hooks/test"
	"errors"
)

func TestGet(t *testing.T) {

	t.Run("Error Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelError).Level, logrus.ErrorLevel, "level incorrect")
	})

	t.Run("Warn Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelWarning).Level, logrus.WarnLevel, "level incorrect")
	})

	t.Run("Debug Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelDebug).Level, logrus.DebugLevel, "level incorrect")
	})

	t.Run("Default Level", func(t *testing.T) {
		assert.Equal(t, Get("default").Level, logrus.InfoLevel, "level incorrect")
	})
}

func TestError(t *testing.T) {
	const (
		errDescription =  "some error"
		errDetails = "my new error"
	)

	logger, hook := test.NewNullLogger()

	e := errors.New(errDetails)
	Error(errDescription, e, logger)
	assert.Equal(t, 1, len(hook.AllEntries()))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, errDescription, hook.LastEntry().Message)
}