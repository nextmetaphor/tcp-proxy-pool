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
	logErrorCreatingListener          = "Error creating customTLSListener"
	logErrorAcceptingConnection       = "Error accepting connection"
	logErrorCopying                   = "Error copying"
	logErrorClosing                   = "Error closing"
	logErrorCreatingMonitorConnection = "Error creating monitoring connection"
	logErrorCreatingMonitorBatch      = "Error creating monitoring batch"
	logErrorLoadingCertificates       = "Error loading certificates"
	logErrorServerConnNotTCP          = "Error: server connection not TCP"
	logErrorClientConnNotTCP          = "Error: client connection not TCP"
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

	tcpProtocol := *(*ctx.Flags)[application.TransportProtocolFlag].FlagValue
	tcpIP := *(*ctx.Flags)[application.HostNameFlag].FlagValue
	//tcpPort := *(*ctx.Flags)[application.PortNameFlag].FlagValue

	cert, err := tls.LoadX509KeyPair(*(*ctx.Flags)[application.CertFileFlag].FlagValue, *(*ctx.Flags)[application.KeyFileFlag].FlagValue)
	if err != nil {
		ctx.Log.Error(logErrorLoadingCertificates, err)
		return false
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, listenErr := Listen(tcpProtocol, tcpIP+":28443", tlsConfig)
	if listener != nil {
		defer listener.Close()
	}

	if listenErr != nil {
		ctx.Log.Error(logErrorCreatingListener, err)
		return false
	}

	ctx.handleConnections(listener, monitorClient)

	return true
}

func (ctx Context) handleConnections(listener net.Listener, monitorClient client.Client) {
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
	upstreamConn := ctx.GetUpstreamConnection()
	if upstreamConn != nil {
		//ctx.proxy(serverConn.(*customTCPConn).InnerConn.(*net.TCPConn), upstreamConn.(*net.TCPConn), monitorClient)
		ctx.proxy(serverConn, upstreamConn, monitorClient)
	}
}

func (ctx Context) proxy(server, client net.Conn, monitorClient client.Client) {
	clientClosedChannel := make(chan struct{}, 1)
	serverClosedChannel := make(chan struct{}, 1)

	go ctx.connectionCopy(false, server, client, clientClosedChannel, monitorClient)
	go ctx.connectionCopy(true, client, server, serverClosedChannel, monitorClient)

	var waitChannel chan struct{}
	select {
	case <-clientClosedChannel:
		if customTCPConn, customTCPConnErr := server.(*customTCPConn); customTCPConnErr {
			if tcpConn, tcpConnErr := customTCPConn.InnerConn.(*net.TCPConn); tcpConnErr {
				tcpConn.SetLinger(0)
				tcpConn.Close()
			} else {
				ctx.Log.Warn(logErrorServerConnNotTCP)
			}
		} else {
			ctx.Log.Warn(logErrorServerConnNotTCP)
		}
		waitChannel = serverClosedChannel
	case <-serverClosedChannel:
		if tcpConn, tcpConnErr := client.(*net.TCPConn); tcpConnErr {
			tcpConn.CloseRead()
		}  else {
			ctx.Log.Warn(logErrorClientConnNotTCP)
		}
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
