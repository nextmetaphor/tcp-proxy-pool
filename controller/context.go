package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrpool"
)

type (
	// Context is a struct representing all the components needed by a listener
	Context struct {
		// Logger needs to be a pointer due to MutexWrap
		Logger        *logrus.Logger
		Settings      application.Settings
		MonitorClient monitor.Client
		ContainerPool *cntrpool.ContainerPool
	}
)
