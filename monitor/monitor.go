package monitor

import (
	"github.com/sirupsen/logrus"
	"net"
)

type (
	// Settings represents the various configuration parameters for a monitor and are typically read
	// from an external configuration file
	Settings struct {
		Address  string
		Database string
	}

	// Client represents the internal representation of a monitor, specifically containing
	// references to the logging components, monitor client etc needed
	Client struct {
		logger   *logrus.Logger
		settings Settings
	}

	// Monitor should be implemented to write to a time-series database for the various methods required
	Monitor interface {
		WriteBytesCopied(srcIsServer bool, totalBytesCopied int64, dst, src net.Conn)
		WriteConnectionAccepted(src net.Conn)
		WriteConnectionPoolStats(src net.Conn, connectionsInUse, connectionPoolSize int)
		WriteContainerCreated(numContainersCreated int)
		WriteContainerDestroyed(numContainersDestroyed int)
		CloseMonitorConnection()
	}
)
