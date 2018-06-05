package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
)

type (
	Context struct {
		Logger        *logrus.Logger
		Settings      application.Settings
		MonitorClient monitor.MonitorClient
		ContainerPool *ContainerPool
	}
)
