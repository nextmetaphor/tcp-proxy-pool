package cntrpool

import (
	"net"
	"errors"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"sync"
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
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
	Settings struct {
		InitialSize    int
		MaximumSize    int
		TargetFreeSize int
	}

	ContainerPool struct {
		sync.RWMutex

		// TotalContainersInUse can be calculated from Containers but included here for speed purposes
		TotalContainersInUse int

		Containers map[string]*cntr.Container
	}
)

// CreateContainerPool creates a connection pool
func CreateContainerPool(cm cntrmgr.ContainerManager, ps Settings, l logrus.Logger) *ContainerPool {
	pool := ContainerPool{
		Containers: make(map[string]*cntr.Container),
	}

	// TODO create containers in parallel? this could take a while...
	for i := 0; i < ps.InitialSize; i++ {
		c, err := CreateContainer(&pool, cm)
		if err != nil {
			log.Error(logErrorCreatingContainer, err, &l)
			break
		}
		l.Infof(logCreatedContainer, c.ExternalID)
	}

	return &pool
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
	pool.Lock()
	defer pool.Unlock()
	pool.Containers[c.ExternalID] = c

	return c, nil
}

func DestroyContainer(containerID string, pool *ContainerPool, cm cntrmgr.ContainerManager) (err error) {
	if pool == nil {
		return errors.New(errorContainerPoolNilCannotDestroy)
	}

	// TODO errors?
	cm.DestroyContainer(containerID)

	pool.Lock()
	defer pool.Unlock()
	delete(pool.Containers, containerID)

	// TODO - add monitoring here
	return nil
}

func AssociateClientWithContainer(conn net.Conn, pool *ContainerPool, mon monitor.Client) (*cntr.Container, error) {
	for _, container := range pool.Containers {
		// find the first container with no current connection from the client
		if container.ConnectionFromClient == nil {
			container.Lock()
			if container.ConnectionFromClient == nil {
				container.ConnectionFromClient = conn

				pool.Lock()
				pool.TotalContainersInUse++
				mon.WriteConnectionPoolStats(conn, pool.TotalContainersInUse, len(pool.Containers))
				pool.Unlock()

				mon.WriteConnectionAccepted(conn)

				container.Unlock()

				return container, nil
			}

			// ...otherwise another thread has beat us to it - try and find another one
			container.Unlock()
		}
	}

	mon.WriteConnectionRejected(conn)
	return nil, errors.New(errorContainerPoolFull)
}

func DissociateClientWithContainer(serverConn net.Conn, pool *ContainerPool, c *cntr.Container, mon monitor.Client, logger logrus.Logger) {
	if c == nil {
		logger.Warnf(logNilContainerToDisassociate)
		return
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = nil
	c.ConnectionFromClient = nil

	pool.Lock()
	pool.TotalContainersInUse--
	mon.WriteConnectionPoolStats(serverConn, pool.TotalContainersInUse, len(pool.Containers))
	pool.Unlock()
}

func ConnectClientToContainer(c *cntr.Container) (error) {
	conn, err := net.Dial("tcp", c.IPAddress+":"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}

	// no need to lock here
	c.ConnectionToContainer = conn

	return nil
}
