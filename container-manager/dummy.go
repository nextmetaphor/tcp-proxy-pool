package container_manager

import (
	"time"
	"math/rand"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/container"
)

type (
	DummyContainerManager struct{}
)

func (cm DummyContainerManager) CreateContainer() (*container.Container, error) {
	// TODO
	return &container.Container{
		ExternalID: strconv.Itoa(rand.Int()),
		StartTime:  time.Now(),
		IPAddress:  "192.168.64.26",
		Port:       32583,
	}, nil
}

func (cm DummyContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}
