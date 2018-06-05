package controller

import (
	"net"
	"errors"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
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
	ContainerPool struct {
		// TotalContainersInPool can be calculated from Containers but included here for speed purposes
		TotalContainersInPool int
		// TotalContainersInUse can be calculated from Containers but included here for speed purposes
		TotalContainersInUse int

		Containers map[string]*cntr.Container
	}
)

// InitialiseContainerPool creates a connection pool
func (ctx *Context) InitialiseContainerPool(cm cntrmgr.ContainerManager) {
	pool := ContainerPool{
		Containers: make(map[string]*cntr.Container),
	}
	ctx.ContainerPool = &pool

	poolSize := ctx.Settings.Pool.InitialSize

	// TODO create containers in parallel? this could take a while...
	for i := 0; i < poolSize; i++ {
		c, err := CreateContainer(ctx.ContainerPool, cm)
		if err != nil {
			log.Error(logErrorCreatingContainer, err, ctx.Logger)
			break
		}
		ctx.Logger.Infof(logCreatedContainer, c.ExternalID)
	}
}

// CreateContainer creates a new Container and adds it to the ContainerPool, indexed by the ExternalID of the
// created container.
func CreateContainer(pool *ContainerPool, cm cntrmgr.ContainerManager) (c *cntr.Container, err error) {
	if pool == nil {
		return c, errors.New(errorContainerPoolNilCannotCreate)
	}

	c, err = cm.CreateContainer()
	if err != nil {
		// TODO - add monitoring here
		return c, err
	}

	if (c == nil) {
		return c, errors.New(errorCreatedContainerCannotBeNil)
	}
	pool.Containers[c.ExternalID] = c
	pool.TotalContainersInPool++

	return c, nil
}

func DestroyContainer(containerID string, pool *ContainerPool, cm cntrmgr.ContainerManager) (err error) {
	if pool == nil {
		return errors.New(errorContainerPoolNilCannotDestroy)
	}

	cm.DestroyContainer(containerID)
	delete(pool.Containers, containerID)
	pool.TotalContainersInPool--

	// TODO - add monitoring here
	return nil
}

func (ctx *Context) AssociateClientWithContainer(conn net.Conn) (*cntr.Container, error) {
	// TODO - would it be better to lock the whole pool?
	for _, container := range ctx.ContainerPool.Containers {
		// find the first container with no current connection from the client
		if container.ConnectionToContainer == nil {
			container.Lock()
			if container.ConnectionFromClient == nil {
				container.ConnectionFromClient = conn
				ctx.ContainerPool.TotalContainersInUse++
				container.Unlock()

				ctx.MonitorClient.WriteConnectionAccepted(conn)


				return container, nil
			}

			// ...otherwise another thread has beat us to it - try and find another one
			container.Unlock()
		}
	}

	ctx.MonitorClient.WriteConnectionRejected(conn)
	return nil, errors.New(errorContainerPoolFull)
}

func (ctx *Context) DissociateClientWithContainer(c *cntr.Container) {
	if c == nil {
		ctx.Logger.Warnf(logNilContainerToDisassociate)
		return
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = nil
	c.ConnectionFromClient = nil
	ctx.ContainerPool.TotalContainersInUse--
}

func (ctx *Context) ConnectClientToContainer(c *cntr.Container) (error) {
	conn, err := net.Dial("tcp", c.IPAddress+":"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = conn

	return nil
}
