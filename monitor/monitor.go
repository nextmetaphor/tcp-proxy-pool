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
	Settings struct {
		Address  string
		Database string
	}

	Client struct {
		Logger   *logrus.Logger
		Settings Settings
		Client   *client.Client
	}
)

func (mon *Client) CreateMonitor() *client.Client {
	if (mon == nil) || strings.TrimSpace(mon.Settings.Address) == "" {
		return nil
	}

	monitorClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: mon.Settings.Address,
	})
	if err != nil {
		log.Error(logErrorCreatingMonitorConnection, err, mon.Logger)
	}

	mon.Client = &monitorClient
	return &monitorClient
}

func (mon *Client) writePoint(measurementName string, tags map[string]string, fields map[string]interface{}) {
	if strings.TrimSpace(mon.Settings.Address) == "" {
		return
	}

	// TODO - new batch every time???
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  mon.Settings.Database,
		Precision: "ns",
	})
	if err != nil {
		log.Error(logErrorCreatingMonitorBatch, err, mon.Logger)
		return
	}

	pt, err := client.NewPoint(measurementName, tags, fields, time.Now())
	if err != nil {
		log.Error(logErrorCreatingPoint, err, mon.Logger)
		return
	}
	bp.AddPoint(pt)

	if mon.Client != nil {
		if err := (*mon.Client).Write(bp); err != nil {
			log.Error(logErrorWritingPoint, err, mon.Logger)
		}
	}
}

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

func (mon *Client) WriteConnectionAccepted(src net.Conn) {
	go mon.writePoint(
		measurementConnectionPool,
		map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.RemoteAddr().String(),
		},
		map[string]interface{}{fieldConnectionsAccepted: 1})
}

func (mon *Client) WriteConnectionRejected(src net.Conn) {
	go mon.writePoint(
		measurementConnectionPool,
		map[string]string{
			tagTCPProxyPoolClientConn: src.LocalAddr().String(),
			tagTCPProxyPoolServerConn: src.RemoteAddr().String(),
		},
		map[string]interface{}{fieldConnectionsRejected: 1})
}

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
