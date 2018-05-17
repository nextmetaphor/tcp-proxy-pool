package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/aws-container-factory/application"
)

type (
	Context struct {
		Log   *logrus.Logger
		Flags *application.CommandLineFlags
	}
)
