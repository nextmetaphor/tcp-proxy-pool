package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/nextmetaphor/tcp-proxy-pool/configuration"
)

type (
	Context struct {
		Logger         *logrus.Logger
		Settings       configuration.ApplicationSettings
		//Flags          *application.CommandLineFlags
		InfluxDBClient *client.Client
		ContainerPool  *ContainerPool
	}
)
