package container_manager

import "github.com/nextmetaphor/tcp-proxy-pool/container"

type (
	ContainerManager interface {
		CreateContainer() (*container.Container, error)
		DestroyContainer(externalID string) (error)
	}
)