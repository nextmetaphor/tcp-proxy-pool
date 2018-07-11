package cntr

import (
	"time"
	"net"
)

type (
	// Container contains the details of a running container within the pool
	Container struct {
		// ExternalID is the ID of the running container and must be unique within the pool
		ExternalID            string

		// StartTime holds the time that the container was initially started
		StartTime             time.Time

		// IPAddress holds the IP address on which the container is running
		IPAddress             string

		// Port holds the port on which the container is running
		Port                  int

		// ConnectionFromClient represents the client connection; if this is nil then this container is available
		ConnectionFromClient  net.Conn

		// ConnectionToContainer represents the container connection; this should not be set to nil once set
		ConnectionToContainer net.Conn
	}
)
