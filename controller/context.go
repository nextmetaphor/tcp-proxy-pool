package controller

import (
	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type (
	Context struct {
		Log            *logrus.Logger
		Router         *mux.Router
	}
)
