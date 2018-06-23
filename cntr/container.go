package cntr

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

		// IsBeingRemoved is set to true when a container connection is in the process of being removed
		IsBeingRemoved bool

		// ConnectionFromClient represents the client connection; if this is nil then this container is available
		ConnectionFromClient  net.Conn

		// ConnectionToContainer represents the container connection; this should not be set to nil once set
		ConnectionToContainer net.Conn
	}
)
