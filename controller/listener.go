package controller

import (
	"net"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"io"
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"log"
	"crypto/tls"
)

const (
	logSecureServerStarting           = "Server starting on address [%s] and port [%s] with a secure configuration: cert[%s] key[%s]"
	logErrorCreatingListener          = "Error creating listener"
	logErrorAcceptingConnection       = "Error accepting connection"
	logErrorCopying                   = "Error copying"
	logErrorClosing                   = "Error closing"
	logErrorCreatingMonitorConnection = "Error creating monitoring connection"
	logErrorCreatingMonitorBatch      = "Error creating monitoring batch"
	logErrorLoadingCertificates       = "Error loading certificates"
)

var totalConnections int

func (ctx Context) writePoint(monitorClient client.Client, measurementName string, tags map[string]string, fields map[string]interface{}) {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "tcp-proxy-pool",
		Precision: "ns",
	})
	if err != nil {
		ctx.Log.Error(logErrorCreatingMonitorBatch, err)
	}

	pt, err := client.NewPoint(measurementName, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	if err := monitorClient.Write(bp); err != nil {
		ctx.Log.Error(err)
	}
}

func (ctx Context) StartListener() bool {
	//First start the monitor client, we'll need to pass this around
	monitorClient, err := client.NewUDPClient(client.UDPConfig{
		Addr: "192.168.64.26:30102",
	})

	//monitorClient, err := client.NewHTTPClient(client.HTTPConfig{
	//	Addr:     "http://influxdb:8086",
	//	Username: "admin",
	//	Password: "admin",
	//})

	if err != nil {
		ctx.Log.Error(logErrorCreatingMonitorConnection, err)
	}
	defer monitorClient.Close()

	ctx.Log.Infof(logSecureServerStarting,
		*(*ctx.Flags)[application.HostNameFlag].FlagValue,
		*(*ctx.Flags)[application.PortNameFlag].FlagValue,
		*(*ctx.Flags)[application.CertFileFlag].FlagValue,
		*(*ctx.Flags)[application.KeyFileFlag].FlagValue)

	var tcpListener net.Listener
	var listenErr error

	tcpProtocol := *(*ctx.Flags)[application.TransportProtocolFlag].FlagValue
	tcpIP := *(*ctx.Flags)[application.HostNameFlag].FlagValue
	//tcpPort := *(*ctx.Flags)[application.PortNameFlag].FlagValue

	ctx.Log.Debug("About to listen...")
	tcpListener, listenErr = net.ListenTCP(tcpProtocol, &net.TCPAddr{IP: []byte(tcpIP), Port: 8443})
	// make sure we close the tcpListener, even if we have an error in the next step
	if tcpListener != nil {
		defer tcpListener.Close()
	}

	if listenErr != nil {
		ctx.Log.Error(logErrorCreatingListener, err)
		return false
	}

	// add TLS configuration if required
	if *(*ctx.Flags)[application.CertFileFlag].FlagValue != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(*(*ctx.Flags)[application.CertFileFlag].FlagValue, *(*ctx.Flags)[application.KeyFileFlag].FlagValue)
		if err != nil {
			ctx.Log.Error(logErrorLoadingCertificates, err)
			return false
		}

		ctx.Log.Debug("Definitely listening on a TLS connection...")
		tcpListener = NewListener(tcpListener, &tls.Config{Certificates: []tls.Certificate{cert}})
	} else {
		ctx.Log.Debug("Definitely NOT listening on a TLS connection...")
	}

	ctx.handleConnections(tcpListener, monitorClient)

	return true
}

func (ctx Context) handleConnections(tcpListener net.Listener, monitorClient client.Client) {
	listener := tcpListener

	for {
		conn, err := listener.Accept()
		if err != nil {
			ctx.Log.Error(logErrorAcceptingConnection, err)
		}
		totalConnections++
		ctx.writePoint(monitorClient,
			"connections",
			map[string]string{"listen-requests": "request-count"},
			map[string]interface{}{"request-count": totalConnections})

		go ctx.handleConnection(conn, monitorClient)
	}
}

func (ctx Context) handleConnection(serverConn net.Conn, monitorClient client.Client) {
	conn := ctx.GetUpstreamConnection()
	if conn != nil {
		ctx.proxy(serverConn.(*tcpConnection).InnerConn.(*net.TCPConn), conn.(*net.TCPConn), monitorClient)
	}
}

func (ctx Context) proxy(server, client *net.TCPConn, monitorClient client.Client) {
	clientClosedChannel := make(chan struct{}, 1)
	serverClosedChannel := make(chan struct{}, 1)

	go ctx.connectionCopy(false, server, client, clientClosedChannel, monitorClient)
	go ctx.connectionCopy(true, client, server, serverClosedChannel, monitorClient)

	var waitChannel chan struct{}
	select {
	case <-clientClosedChannel:
		server.SetLinger(0)
		server.CloseRead()
		waitChannel = serverClosedChannel
	case <-serverClosedChannel:
		client.CloseRead()
		waitChannel = clientClosedChannel
	}

	<-waitChannel

	totalConnections--
	ctx.writePoint(monitorClient,
		"connections",
		map[string]string{"listen-requests": "request-count"},
		map[string]interface{}{"request-count": totalConnections})
}

func (ctx Context) connectionCopy(srcIsServer bool, dst, src net.Conn, sourceClosedChannel chan struct{}, monitorClient client.Client) {
	bytesCopied, err := io.Copy(dst, src);
	if err != nil {
		ctx.Log.Error(logErrorCopying, err)
	}

	if srcIsServer {
		ctx.writePoint(monitorClient,
			"connections",
			map[string]string{"bytes-copied": "total"},
			map[string]interface{}{"bytesCopiedFromClient": bytesCopied})
	} else {
		ctx.writePoint(monitorClient,
			"connections",
			map[string]string{"bytes-copied": "total"},
			map[string]interface{}{"bytesCopiedFromServer": bytesCopied})
	}

	if err := src.Close(); err != nil {
		ctx.Log.Error(logErrorClosing, err)
	}

	sourceClosedChannel <- struct{}{}
}
