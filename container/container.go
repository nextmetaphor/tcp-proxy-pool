package container

import (
	"sync"
	"time"
	"net"
)

type (
	Container struct {
		sync.RWMutex
		ExternalID            string
		StartTime             time.Time
		IPAddress             string
		Port                  int
		IsReady               bool
		ConnectionFromClient  net.Conn
		ConnectionToContainer net.Conn
	}
)
