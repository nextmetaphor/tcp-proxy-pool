package monitor

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	logErrorCreatingMonitorBatch      = "Error creating monitoring batch"
	logErrorCreatingMonitorConnection = "Error creating monitoring connection"
	logErrorCreatingPoint             = "Error creating point"
	logErrorWritingPoint              = "Error writing point"

	measurementDataTransfer = "data-transfer"
	fieldCopiedToServer     = "copied-to-server"
	fieldCopiedFromServer   = "copied-from-server"

	measurementConnectionPool = "connection-pool"
	fieldConnectionsAccepted  = "connections-accepted"
	fieldConnectionsRejected  = "connections-rejected"
	fieldConnectionsInUse     = "connections-in-use"
	fieldConnectionPoolSize   = "connection-pool-size"

	tagTCPProxyPoolClientConn = "client-conn"
	tagTCPProxyPoolServerConn = "server-conn"
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
		Client   client.Client
	}

	Monitor interface {
		WriteBytesCopied(srcIsServer bool, totalBytesCopied int64, dst, src net.Conn)
		WriteConnectionAccepted(src net.Conn)
		WriteConnectionPoolStats(src net.Conn, connectionsInUse, connectionPoolSize int)
	}
)