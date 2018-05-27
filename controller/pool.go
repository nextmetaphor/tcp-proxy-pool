package controller

import (
	"net"
	"time"
	"errors"
	"sync"
	"strconv"
)

const (
	logErrorCreatingContainer     = "Error creating container"
	logCreatedContainer           = "Created container with ID [%s]"
	logNilContainerToDisassociate = "Nil container to disassociate from the container pool"
	logContainerDoesNotExist      = "The container with ID [%s] to disassociate from the client does not exist in the pool"

	errorContainerPoolNilCannotCreate  = "Pool is nil; cannot create container"
	errorCreatedContainerCannotBeNil   = "Created container cannot be nil"
	errorContainerPoolNilCannotDestroy = "Pool is nil; cannot destroy container"
	errorContainerPoolFull             = "Pool is full; cannot allocate connection to container"
)

type (
	Container struct {
		sync.RWMutex
		ExternalID            string
		StartTime             time.Time
		IPAddress             string
		Port                  int
		IsReady               bool
		ConnectionFromClient  net.Conn
		ConnectionToContainer net.Conn
	}

	ContainerPool map[string]*Container

	ContainerManager interface {
		CreateContainer() *Container
	}
)

// InitialiseContainerPool creates a connection pool
func (ctx *Context) InitialiseContainerPool(cm ContainerManager) {
	pool := make(ContainerPool)
	ctx.ContainerPool = &pool

	// TODO - this needs to be a configuration parameter
	poolSize := 10

	for i := 0; i < poolSize; i++ {
		c, err := CreateContainer(ctx.ContainerPool, cm)
		if err != nil {
			ctx.Log.Error(logErrorCreatingContainer, err)
			break
		}
		ctx.Log.Infof(logCreatedContainer, c.ExternalID)
	}
}

// CreateContainer creates a new Container and adds it to the ContainerPool, indexed by the ExternalID of the
// created container.
func CreateContainer(pool *ContainerPool, cm ContainerManager) (c *Container, err error) {
	if pool == nil {
		return c, errors.New(errorContainerPoolNilCannotCreate)
	}

	c = cm.CreateContainer()
	if c == nil {
		return c, errors.New(errorCreatedContainerCannotBeNil)
	}

	(*pool)[c.ExternalID] = c

	return c, nil
}

func (ctx *Context) DestroyContainer(containerID string, pool *ContainerPool) (err error) {
	if pool == nil {
		return errors.New(errorContainerPoolNilCannotDestroy)
	}

	// TODO - make external call to remove container
	delete((*pool), containerID)
	return nil
}



func (ctx *Context) AssociateClientWithContainer(conn net.Conn) (*Container, error) {
	// TODO - would it be better to lock the whole pool?
	for _, container := range *ctx.ContainerPool {
		// find the first container with no current connection from the client
		if container.ConnectionToContainer == nil {
			container.Lock()
			if container.ConnectionFromClient == nil {
				container.ConnectionFromClient = conn
				container.Unlock()
				return container, nil
			}

			// ...otherwise another thread has beat us to it - try and find another one
			container.Unlock()
		}
	}

	return nil, errors.New(errorContainerPoolFull)
}

func (ctx *Context) DissociateClientWithContainer(c *Container) {
	if c == nil {
		ctx.Log.Warnf(logNilContainerToDisassociate)
		return
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = nil
	c.ConnectionFromClient = nil
}

func (ctx *Context) ConnectClientToContainer(c *Container) (error) {
	conn, err := net.Dial("tcp", c.IPAddress+":"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = conn

	return nil
}
