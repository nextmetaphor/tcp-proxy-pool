package cntrmgr

import (
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
)

type (
	ContainerManager interface {
		CreateContainer() (*cntr.Container, error)
		DestroyContainer(externalID string) (error)
	}
)