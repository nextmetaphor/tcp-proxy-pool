package controller

import (
	"net"
	"github.com/nextmetaphor/aws-container-factory/application"
)

const (
	logSecureServerStarting  = "Server starting on address [%s] and port [%s] with a secure configuration: cert[%s] key[%s]"
	logErrorCreatingListener = "Error creating listener"
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

	return true
}
