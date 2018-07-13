package cntrmgr

import (
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
)

type (
	// ContainerManager contains simple methods that need to be implemented for every container manager, specifically
	// to create a container, and to destroy a specified container
	ContainerManager interface {
		CreateContainer() (*cntr.Container, error)
		DestroyContainer(externalID string) (error)
	}
)