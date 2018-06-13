package controller

import (
	"net"
	"crypto/tls"
	"errors"
)

type (
	customTLSListener struct {
		net.Listener
		config *tls.Config
	}

	customTCPConn struct {
		*tls.Conn
		InnerConn net.Conn
	}
)

func (l *customTLSListener) Accept() (net.Conn, error) {
	tcpConn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Server(tcpConn, l.config)

	return &customTCPConn{
		Conn:      tlsConn,
		InnerConn: tcpConn,
	}, nil
}

func NewListener(inner net.Listener, config *tls.Config) net.Listener {
	l := new(customTLSListener)
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
