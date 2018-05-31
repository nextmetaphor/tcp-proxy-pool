package main

import (
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/controller"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"os"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
)

const (
	// command-line flags
	logSignalReceived           = "Signal [%s] received, shutting down server"
	settingsFilename            = "tcp-proxy-pool.json"
	logErrorLoadingSettingsFile = "Error loading settings file"
)

func main() {
	//construct the flag map - needed before we construct the logger
	flags := application.CreateFlags()

	// set flags in conjunction with actual command-line arguments
	flags.SetFlagsWithArguments(os.Args[1:])

	// create the main context with the logger and flags
	ctx := controller.Context{
		Logger: log.Get(*flags[application.LogLevelFlag].FlagValue),
	}

	settings, err := application.LoadSettings(settingsFilename)
	if err != nil {
		log.LogError(logErrorLoadingSettingsFile, err, ctx.Logger)
	} else {
		ctx.Settings = *settings
	}

	// TODO overrride settings with flags

	// start the monitoring service
	go ctx.StartMonitor()

	// create the container manager
	cm := cntrmgr.ECS{
		Logger: *ctx.Logger,
		Conf: ctx.Settings.ECS,
	}
	cm.InitialiseECSService()

	//cm := container_manager.DummyContainerManager{}

	// start a listener
	ctx.StartListener(cm)

}
