package cntrmgr

import (
	"time"
	"math/rand"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
)

type (
	// DummyContainerManager simply forwards all container requests to a standard IP address and port.
	// This container needs to be already running; it does not create or destroy any containers as part of the
	// lifecycle. It is primarily used for testing.
	DummyContainerManager struct{}
)

// CreateContainer simply returns the configuration of an (assumed) already-running container. It is required
// to satisfy the ContainerManager interface.
func (cm DummyContainerManager) CreateContainer() (*cntr.Container, error) {
	// TODO
	return &cntr.Container{
		ExternalID: strconv.Itoa(rand.Int()),
		StartTime:  time.Now(),
		IPAddress:  "192.168.64.30",
		Port:       32523,
	}, nil
}

// DestroyContainer does nothing! It is required to satisfy the ContainerManager interface.
func (cm DummyContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}
