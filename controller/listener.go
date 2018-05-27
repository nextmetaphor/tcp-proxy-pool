package controller

import (
	"net"
	"github.com/nextmetaphor/tcp-proxy-pool/application"
	"io"
	"crypto/tls"
)

const (
	logSecureServerStarting     = "Server starting on address [%s] and port [%s] with a secure configuration: cert[%s] key[%s]"
	logErrorCreatingListener    = "Error creating customTLSListener"
	logErrorAcceptingConnection = "Error accepting connection"
	logErrorCopying             = "Error copying"
	logErrorClosing             = "Error closing"
	logErrorLoadingCertificates = "Error loading certificates"
	logErrorServerConnNotTCP    = "Error: server connection not TCP"
	logErrorClientConnNotTCP    = "Error: client connection not TCP"
	logErrorAssigningContainer  = "Error: cannot assign container"
	logErrorProxyingConnection  = "Error proxying connection"
)

func (ctx *Context) StartListener() bool {
	ctx.InitialiseContainerPool(ECSContainerManager{})

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

// handleConnections is called when the container pool has been initialised and the listener has been started.
// A separate goroutine is created to handle each Accept request on the listener.
func (ctx *Context) handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			ctx.Log.Error(logErrorAcceptingConnection, err)
			return
		}

		go ctx.clientConnect(conn)
	}
}

// clientConnect is called in a separate goroutine for every successful Accept request on the server listener.
func (ctx *Context) clientConnect(serverConn net.Conn) {
	c, err := ctx.AssociateClientWithContainer(serverConn)
	if err != nil {
		ctx.Log.Warn(logErrorAssigningContainer, err)
		// TODO - check for errors
		serverConn.Close()
		return
	}

	if err := ctx.ConnectClientToContainer(c); err != nil {
		ctx.Log.Error(logErrorProxyingConnection, err)
		return
	}

	ctx.proxy(c)
}

func (ctx *Context) proxy(c *Container) {
	server := c.ConnectionFromClient
	client := c.ConnectionToContainer

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
				tcpConn.CloseRead()
			} else {
				ctx.Log.Warn(logErrorServerConnNotTCP)
				customTCPConn.InnerConn.Close()
			}
		} else {
			ctx.Log.Warn(logErrorServerConnNotTCP)
			server.Close()
		}
		waitChannel = serverClosedChannel

	case <-serverClosedChannel:
		if tcpConn, tcpConnErr := client.(*net.TCPConn); tcpConnErr {
			tcpConn.CloseRead()
		} else {
			ctx.Log.Warn(logErrorClientConnNotTCP)
		}
		waitChannel = clientClosedChannel
	}

	<-waitChannel

	ctx.DissociateClientWithContainer(c)
}

func (ctx *Context) connectionCopy(srcIsServer bool, dst, src net.Conn, sourceClosedChannel chan struct{}) {
	_, err := io.Copy(dst, src);
	if err != nil {
		ctx.Log.Error(logErrorCopying, err)
	}

	if err := src.Close(); err != nil {
		ctx.Log.Error(logErrorClosing, err)
	}

	sourceClosedChannel <- struct{}{}
}
