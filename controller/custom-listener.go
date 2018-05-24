package controller

import (
	"net"
	"crypto/tls"
	"errors"
)

type (
	listener struct {
		net.Listener
		config *tls.Config
	}

	tcpConnection struct {
		tls.Conn
		InnerConn net.Conn
	}
)

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	daConn := tls.Server(c, l.config)

	return &tcpConnection{
		Conn:      *daConn,
		InnerConn: c,
	}, nil
}

func NewListener(inner net.Listener, config *tls.Config) net.Listener {
	l := new(listener)
	l.Listener = inner
	l.config = config
	return l
}

func Listen(network, laddr string, config *tls.Config) (net.Listener, error) {
	if config == nil || (len(config.Certificates) == 0 && config.GetCertificate == nil) {
		return nil, errors.New("tls: neither Certificates nor GetCertificate set in Config")
	}
	l, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return NewListener(l, config), nil
}
