package controller

import (
	"net"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"io"
	"crypto/tls"
)

const (
	logSecureServerStarting           = "Server starting on address [%s] and port [%s] with a secure configuration: cert[%s] key[%s]"
	logErrorCreatingListener          = "Error creating customTLSListener"
	logErrorAcceptingConnection       = "Error accepting connection"
	logErrorCopying                   = "Error copying"
	logErrorClosing                   = "Error closing"
	logErrorLoadingCertificates       = "Error loading certificates"
	logErrorServerConnNotTCP          = "Error: server connection not TCP"
	logErrorClientConnNotTCP          = "Error: client connection not TCP"
)

var totalConnections int

func (ctx Context) StartListener() bool {
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

	ctx.handleConnections(listener)

	return true
}

func (ctx Context) handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			ctx.Log.Error(logErrorAcceptingConnection, err)
		}
		totalConnections++

		go ctx.handleConnection(conn)
	}
}

func (ctx Context) handleConnection(serverConn net.Conn) {
	upstreamConn := ctx.GetUpstreamConnection()
	if upstreamConn != nil {
		ctx.proxy(serverConn, upstreamConn)
	}
}

func (ctx Context) proxy(server, client net.Conn) {
	clientClosedChannel := make(chan struct{}, 1)
	serverClosedChannel := make(chan struct{}, 1)

	go ctx.connectionCopy(false, server, client, clientClosedChannel)
	go ctx.connectionCopy(true, client, server, serverClosedChannel)

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
}

func (ctx Context) connectionCopy(srcIsServer bool, dst, src net.Conn, sourceClosedChannel chan struct{}) {
	_, err := io.Copy(dst, src);
	if err != nil {
		ctx.Log.Error(logErrorCopying, err)
	}

	if err := src.Close(); err != nil {
		ctx.Log.Error(logErrorClosing, err)
	}

	sourceClosedChannel <- struct{}{}
}
