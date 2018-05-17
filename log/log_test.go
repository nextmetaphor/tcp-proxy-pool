package log

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {

	t.Run("Error Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelError).Level, logrus.ErrorLevel, "level incorrect")
	})

	t.Run("Warn Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelWarn).Level, logrus.WarnLevel, "level incorrect")
	})

	t.Run("Debug Level", func(t *testing.T) {
		assert.Equal(t, Get(LevelDebug).Level, logrus.DebugLevel, "level incorrect")
	})

	t.Run("Default Level", func(t *testing.T) {
		assert.Equal(t, Get("default").Level, logrus.InfoLevel, "level incorrect")
	})
}
