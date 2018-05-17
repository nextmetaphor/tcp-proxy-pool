package main

import (
	"github.com/nextmetaphor/aws-container-factory/controller"
	"github.com/nextmetaphor/aws-container-factory/application"
	"github.com/nextmetaphor/aws-container-factory/log"
	"os"
)

const (
	// command-line flags
	logSignalReceived = "Signal [%s] received, shutting down server"
)

func main() {
	//construct the flag map - needed before we construct the logger
	flags := application.CreateFlags()

	// set flags in conjunction with actual command-line arguments
	flags.SetFlagsWithArguments(os.Args[1:])

	// create the main context with the logger and flags
	ctx := controller.Context{
		Log:   log.Get(*flags[application.LogLevelFlag].FlagValue),
		Flags: &flags,
	}

	// start a listener
	ctx.StartListener()

	//for {
	//	// Listen for an incoming connection.
	//	conn, err := l.Accept()
	//	if err != nil {
	//		fmt.Println("Error accepting: ", err.Error())
	//		os.Exit(1)
	//	}
	//	// Handle connections in a new goroutine.
	//	//go handleRequest(conn)
	//}

}
