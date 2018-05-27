package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/influxdata/influxdb/client/v2"
)

type (
	Context struct {
		Log            *logrus.Logger
		Flags          *application.CommandLineFlags
		InfluxDBClient *client.Client
		ContainerPool  *ContainerPool
	}
)
