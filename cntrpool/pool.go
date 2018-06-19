package cntrpool

import (
	"net"
	"errors"
	"strconv"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"sync"
	"github.com/sirupsen/logrus"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
)

const (
	logCreatedContainer = "created container"
	logFieldContainerId = "container-id"

	logErrorCreatingContainer     = "Error creating container"
	logNilContainerToDisassociate = "Nil container to disassociate from the container pool"
	logContainerDoesNotExist      = "The container with ID [%s] to disassociate from the client does not exist in the pool"

	errorContainerManagerNil         = "error creating container pool: container manager cannot be nil"
	errorLoggerNil                   = "error creating container pool: logger cannot be nil"
	errorCreatedContainerCannotBeNil = "created container cannot be nil"
	errorContainerPoolFull           = "pool is full; cannot allocate connection to container"
)

type (
	Settings struct {
		InitialSize    int
		MaximumSize    int
		TargetFreeSize int
	}

	ContainerPool struct {
		sync.RWMutex

		containers map[string]*cntr.Container

		logger   *logrus.Logger
		settings Settings
		manager  cntrmgr.ContainerManager
		monitor  monitor.Client

		// totalContainersInUse can be calculated from containers but included here for speed purposes
		totalContainersInUse int
	}
)

// CreateContainerPool creates a connection pool
func CreateContainerPool(cm cntrmgr.ContainerManager, s Settings, l *logrus.Logger, m monitor.Client) (pool *ContainerPool, err error) {
	if cm == nil {
		return nil, errors.New(errorContainerManagerNil)
	}
	if l == nil {
		return nil, errors.New(errorLoggerNil)
	}

	pool = &ContainerPool{
		containers: make(map[string]*cntr.Container),
		logger:     l,
		settings:   s,
		manager:    cm,
		monitor:    m,
	}

	return pool, nil
}

func (cp *ContainerPool) InitialisePool() (errors []error) {
	// TODO better to create containers in parallel
	for i := 0; i < cp.settings.InitialSize; i++ {
		c, err := cp.CreateContainer()
		if err != nil {
			log.Error(logErrorCreatingContainer, err, cp.logger)
			errors = append(errors, err)
			continue
		}
		cp.logger.WithFields(logrus.Fields{logFieldContainerId: c.ExternalID}).Infof(logCreatedContainer)
	}

	return errors
}

// CreateContainer creates a new Container and adds it to the ContainerPool, indexed by the ExternalID of the
// created container.
func (cp *ContainerPool) CreateContainer() (c *cntr.Container, err error) {
	c, err = cp.manager.CreateContainer()
	if err != nil {
		// TODO - add monitoring here
		return c, err
	}

	if c == nil {
		return c, errors.New(errorCreatedContainerCannotBeNil)
	}
	cp.Lock()
	defer cp.Unlock()
	cp.containers[c.ExternalID] = c

	return c, nil
}

func (cp *ContainerPool) DestroyContainer(containerID string) (err error) {
	// TODO errors?
	cp.manager.DestroyContainer(containerID)

	cp.Lock()
	defer cp.Unlock()
	delete(cp.containers, containerID)

	// TODO - add monitoring here
	return nil
}

func (cp *ContainerPool) AssociateClientWithContainer(conn net.Conn) (*cntr.Container, error) {
	for _, container := range cp.containers {
		// find the first container with no current connection from the client
		if container.ConnectionFromClient == nil {
			container.Lock()
			if container.ConnectionFromClient == nil {
				container.ConnectionFromClient = conn

				cp.Lock()
				cp.totalContainersInUse++
				cp.monitor.WriteConnectionPoolStats(conn, cp.totalContainersInUse, len(cp.containers))
				cp.Unlock()

				cp.monitor.WriteConnectionAccepted(conn)

				container.Unlock()

				return container, nil
			}

			// ...otherwise another thread has beat us to it - try and find another one
			container.Unlock()
		}
	}

	cp.monitor.WriteConnectionRejected(conn)
	return nil, errors.New(errorContainerPoolFull)
}

func (cp *ContainerPool) DissociateClientWithContainer(serverConn net.Conn, c *cntr.Container) {
	if c == nil {
		cp.logger.Warnf(logNilContainerToDisassociate)
		return
	}

	c.Lock()
	defer c.Unlock()
	c.ConnectionToContainer = nil
	c.ConnectionFromClient = nil

	cp.Lock()
	cp.totalContainersInUse--
	cp.monitor.WriteConnectionPoolStats(serverConn, cp.totalContainersInUse, len(cp.containers))
	cp.Unlock()
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
