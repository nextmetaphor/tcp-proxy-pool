package monitor

import (
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"strings"
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
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
)

// CreateMonitor simply creates a pointer to a Client
// TODO return error
func CreateMonitor(ms Settings, l *logrus.Logger) *Client {
	if strings.TrimSpace(ms.Address) == "" {
		return nil
	}

	monitorClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: ms.Address,
	})
	if err != nil {
		log.Error(logErrorCreatingMonitorConnection, err, l)
	}

	return &Client{
		Client:   monitorClient,
		settings: ms,
		logger:   l,
	}
}

func (mon *Client) writePoint(measurementName string, tags map[string]string, fields map[string]interface{}) {
	if strings.TrimSpace(mon.settings.Address) == "" {
		return
	}

	// TODO - new batch every time???
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  mon.settings.Database,
		Precision: "ns",
	})
	if err != nil {
		log.Error(logErrorCreatingMonitorBatch, err, mon.logger)
		return
	}

	pt, err := client.NewPoint(measurementName, tags, fields, time.Now())
	if err != nil {
		log.Error(logErrorCreatingPoint, err, mon.logger)
		return
	}
	bp.AddPoint(pt)

	if mon.Client != nil {
		if err := (mon.Client).Write(bp); err != nil {
			log.Error(logErrorWritingPoint, err, mon.logger)
		}
	}
}

// WriteBytesCopied writes the number of bytes copied to the monitor connection
func (mon *Client) WriteBytesCopied(srcIsServer bool, totalBytesCopied int64, dst, src net.Conn) {
	var fields map[string]interface{}
	var tags map[string]string
	if srcIsServer {
		fields = map[string]interface{}{fieldCopiedFromServer: totalBytesCopied}
		tags = map[string]string{
			tagTCPProxyPoolClientConn: dst.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.LocalAddr().String(),
		}
	} else {
		fields = map[string]interface{}{fieldCopiedToServer: totalBytesCopied}
		tags = map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: dst.LocalAddr().String(),
		}
	}

	go mon.writePoint(
		measurementDataTransfer,
		tags,
		fields)
}

// WriteConnectionAccepted writes a point to the monitor to indicate that a connection was accepted
func (mon *Client) WriteConnectionAccepted(src net.Conn) {
	go mon.writePoint(
		measurementConnectionPool,
		map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.RemoteAddr().String(),
		},
		map[string]interface{}{fieldConnectionsAccepted: 1})
}

// WriteConnectionRejected writes a point to indicate that a connection was rejected
func (mon *Client) WriteConnectionRejected(src net.Conn) {
	go mon.writePoint(
		measurementConnectionPool,
		map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.RemoteAddr().String(),
		},
		map[string]interface{}{fieldConnectionsRejected: 1})
}

// WriteConnectionPoolStats writes a the number of connections in use and the pool size to the monitor
func (mon *Client) WriteConnectionPoolStats(src net.Conn, connectionsInUse, connectionPoolSize int) {
	go mon.writePoint(
		measurementConnectionPool,
		map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.RemoteAddr().String()},
		map[string]interface{}{
			fieldConnectionsInUse:   connectionsInUse,
			fieldConnectionPoolSize: connectionPoolSize})
}
