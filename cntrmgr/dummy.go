package cntrmgr

import (
	"time"
	"math/rand"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
)

type (
	DummyContainerManager struct{}
)

func (cm DummyContainerManager) CreateContainer() (*cntr.Container, error) {
	// TODO
	return &cntr.Container{
		ExternalID: strconv.Itoa(rand.Int()),
		StartTime:  time.Now(),
		IPAddress:  "192.168.64.30",
		Port:       32523,
	}, nil
}

func (cm DummyContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}
