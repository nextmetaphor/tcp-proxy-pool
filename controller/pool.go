package controller

import (
	"net"
)

const (
	logErrorProxyingConnection = "Error proxying connection"
)

func (ctx Context) GetUpstreamConnection() net.Conn {
	conn, err := net.Dial("tcp", "192.168.64.26:32583")
	if err != nil {
		ctx.Log.Error(logErrorProxyingConnection, err)
		return nil
	}

	return conn
}
