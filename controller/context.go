package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
)

type (
	Context struct {
		Logger         *logrus.Logger
		Settings       application.Settings
		InfluxDBClient *client.Client
		ContainerPool  *ContainerPool
	}
)
