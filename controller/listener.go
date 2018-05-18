package controller

import (
	"net"
	"github.com/nextmetaphor/aws-container-factory/application"
	"io"
)

const (
	logSecureServerStarting     = "Server starting on address [%s] and port [%s] with a secure configuration: cert[%s] key[%s]"
	logErrorCreatingListener    = "Error creating listener"
	logErrorAcceptingConnection = "Error accepting connection"
	logErrorProxyingConnection  = "Error proxying connection"
	logErrorCopying             = "Error copying"
	logErrorClosing             = "Error closing"
)

func (ctx Context) StartListener() bool {
	ctx.Log.Infof(logSecureServerStarting,
		*(*ctx.Flags)[application.HostNameFlag].FlagValue,
		*(*ctx.Flags)[application.PortNameFlag].FlagValue,
		*(*ctx.Flags)[application.CertFileFlag].FlagValue,
		*(*ctx.Flags)[application.KeyFileFlag].FlagValue)

	listener, err := net.Listen(
		*(*ctx.Flags)[application.TransportProtocolFlag].FlagValue,
		*(*ctx.Flags)[application.HostNameFlag].FlagValue + ":" + *(*ctx.Flags)[application.PortNameFlag].FlagValue)

	// make sure we close the listener, even if we have an error in the next step
	if listener != nil {
		defer listener.Close()
	}

	if err != nil {
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

		go ctx.handleConnection(conn)
	}
}

func (ctx Context) handleConnection(listener net.Conn) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		ctx.Log.Error(logErrorProxyingConnection, err)
		return
	}

	ctx.proxy(listener.(*net.TCPConn), conn.(*net.TCPConn))
}

func (ctx Context) proxy(server, client *net.TCPConn) {
	serverClosedFlag := make(chan struct{}, 1)
	clientClosedFlag := make(chan struct{}, 1)

	go ctx.connectionCopy(server, client, clientClosedFlag)
	go ctx.connectionCopy(client, server, serverClosedFlag)

	var waitFlag chan struct{}
	select {
	case <-clientClosedFlag:
		server.SetLinger(0)
		server.CloseRead()
		waitFlag = serverClosedFlag
	case <-serverClosedFlag:
		client.CloseRead()
		waitFlag = clientClosedFlag
	}

	<-waitFlag
}

func (ctx Context) connectionCopy(dst, src net.Conn, sourceClosedFlag chan struct{}) {
	if _, err := io.Copy(dst, src); err != nil {
		ctx.Log.Error(logErrorCopying, err)
	}
	if err := src.Close(); err != nil {
		ctx.Log.Error(logErrorClosing, err)
	}
	sourceClosedFlag <- struct{}{}
}
