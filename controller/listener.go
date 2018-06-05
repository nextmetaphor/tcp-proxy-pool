package controller

import (
	"net"
	"io"
	"crypto/tls"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
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
	logErrorAssigningContainer        = "Error: cannot assign container"
	logErrorProxyingConnection        = "Error proxying connection"
)

func (ctx *Context) StartListener(cm cntrmgr.ContainerManager) bool {
	ctx.InitialiseContainerPool(cm)

	monitorClient := ctx.MonitorClient.CreateMonitor()
	if monitorClient != nil {
		defer (*monitorClient).Close()
	}

	ctx.Logger.Infof(logSecureServerStarting,
		ctx.Settings.Listener.Host,
		ctx.Settings.Listener.Port,
		ctx.Settings.Listener.CertFile,
		ctx.Settings.Listener.KeyFile)

	tcpProtocol := ctx.Settings.Listener.Transport
	tcpIP := ctx.Settings.Listener.Host
	tcpPort := ctx.Settings.Listener.Port

	cert, err := tls.LoadX509KeyPair(ctx.Settings.Listener.CertFile, ctx.Settings.Listener.KeyFile)
	if err != nil {
		log.Error(logErrorLoadingCertificates, err, ctx.Logger)
		return false
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, listenErr := Listen(tcpProtocol, tcpIP+":"+tcpPort, tlsConfig)
	if listener != nil {
		defer listener.Close()
	}

	if listenErr != nil {
		log.Error(logErrorCreatingListener, err, ctx.Logger)
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
			log.Error(logErrorAcceptingConnection, err, ctx.Logger)
			return
		}

		go ctx.clientConnect(conn)
	}
}

// clientConnect is called in a separate goroutine for every successful Accept request on the server listener.
func (ctx *Context) clientConnect(serverConn net.Conn) {
	c, err := ctx.AssociateClientWithContainer(serverConn)
	if c != nil {
		defer ctx.DissociateClientWithContainer(c)
	}

	if err != nil {
		ctx.Logger.Warn(logErrorAssigningContainer, err)
		// TODO - check for errors
		serverConn.Close()
		return
	}

	if err := ctx.ConnectClientToContainer(c); err != nil {
		log.Error(logErrorProxyingConnection, err, ctx.Logger)
		return
	}

	ctx.proxy(c)
}

func (ctx *Context) proxy(c *cntr.Container) {
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
				ctx.Logger.Warn(logErrorServerConnNotTCP)
				customTCPConn.InnerConn.Close()
			}
		} else {
			ctx.Logger.Warn(logErrorServerConnNotTCP)
			server.Close()
		}
		waitChannel = serverClosedChannel

	case <-serverClosedChannel:
		if tcpConn, tcpConnErr := client.(*net.TCPConn); tcpConnErr {
			tcpConn.CloseRead()
		} else {
			ctx.Logger.Warn(logErrorClientConnNotTCP)
		}
		waitChannel = clientClosedChannel
	}

	<-waitChannel
}

func (ctx *Context) connectionCopy(srcIsServer bool, dst, src net.Conn, sourceClosedChannel chan struct{}) {
	bytesCopied, err := io.Copy(dst, src);
	if err != nil {
		log.Error(logErrorCopying, err, ctx.Logger)
	}

	ctx.MonitorClient.WriteBytesCopied(srcIsServer, bytesCopied, dst, src)

	if err := src.Close(); err != nil {
		log.Error(logErrorClosing, err, ctx.Logger)
	}

	sourceClosedChannel <- struct{}{}
}
