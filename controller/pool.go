package controller

import (
	"net"
	"time"
	"errors"
)

const (
	logErrorProxyingConnection = "Error proxying connection"
	logErrorCreatingContainer  = "Error creating container"
	logCreatedContainer = "Created container with ID [%s]"

	errorContainerPoolNilCannotCreate  = "Pool is nil; cannot create container"
	errorContainerPoolNilCannotDestroy = "Pool is nil; cannot destroy container"
)

type (
	Container struct {
		ExternalId string
		StartTime  time.Time
		IPAddress  string
		Port       int
		IsReady    bool
		ServerConn net.Conn
		ClientConn net.Conn
	}

	ContainerPool map[string]Container

	ContainerManager interface {
		CreateContainer() Container
	}
)

func CreateContainer(pool *ContainerPool, cm ContainerManager) (containerId string, err error) {
	if pool == nil {
		return "", errors.New(errorContainerPoolNilCannotCreate)
	}

	if pool != nil {
		// TODO - make call to create container

		(*pool)[containerId] = cm.CreateContainer()
	}

	return (*pool)[containerId].ExternalId, nil
}

func (ctx Context) DestroyContainer(containerId string, pool *ContainerPool) (err error) {
	if pool == nil {
		return errors.New(errorContainerPoolNilCannotDestroy)
	}

	// TODO - make external call to remove container
	delete((*pool), containerId)
	return nil
}

func (ctx Context) InitialiseContainerPool() (pool ContainerPool) {
	cm  := ECSContainerManager{}

	// TODO - pool size needs to be a parameter
	poolSize := 4
	pool = make(ContainerPool, poolSize)

	for i := 0; i < poolSize; i++ {
		id, err := CreateContainer(&pool, cm)
		if err != nil {
			ctx.Log.Error(logErrorCreatingContainer, err)
			break
		}
		ctx.Log.Infof(logCreatedContainer, id)
	}

	return pool
}

func (ctx Context) GetUpstreamConnection() net.Conn {
	conn, err := net.Dial("tcp", "192.168.64.26:32583")
	if err != nil {
		ctx.Log.Error(logErrorProxyingConnection, err)
		return nil
	}

	return conn
}
