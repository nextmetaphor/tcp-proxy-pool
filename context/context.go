package context

import (
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/sirupsen/logrus"
)

type Base struct {
	Settings application.Settings
	Logger   logrus.Logger

}