package main

import (
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"github.com/nextmetaphor/tcp-proxy-pool/controller"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"os"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
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

	// load the settings from file
	settings, err := application.LoadSettings(settingsFilename)
	if err != nil {
		log.Error(logErrorLoadingSettingsFile, err, ctx.Logger)
	} else {
		ctx.Settings = *settings
	}

	// TODO overrride settings with flags

	//// start the monitor service
	monitorClient := monitor.CreateMonitor(ctx.Settings.Monitor, ctx.Logger)
	if (monitorClient != nil) && (monitorClient.Client != nil) {
		defer monitorClient.Client.Close()
	}
	ctx.MonitorClient = *monitorClient

	// start the statistics service
	go ctx.StartStatistics()

	// create the appropriate container manager
	cm := cntrmgr.DummyContainerManager{}
	//cm := cntrmgr.ECS{
	//	logger: *ctx.logger,
	//	Conf:   ctx.settings.ECS,
	//}
	//cm.InitialiseECSService()

	// start a listener
	ctx.StartListener(cm)

}
