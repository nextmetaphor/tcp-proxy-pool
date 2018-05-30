package controller

import (
	"time"
	"math/rand"
	"strconv"
)

type (
	ECSContainerManager struct {}
)

func (cm ECSContainerManager) CreateContainer() (*Container, error) {
	// TODO
	return &Container{
		ExternalID: strconv.Itoa(rand.Int()),
		StartTime:  time.Now(),
		IPAddress:  "192.168.64.26",
		Port:       32583,
	}, nil
}

func (cm ECSContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}


