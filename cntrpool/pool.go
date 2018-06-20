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
	return cp.addContainersToPool(cp.settings.InitialSize)
}

func (cp *ContainerPool) addContainersToPool(numContainers int) (errors []error) {
	// TODO obvs better to create containers in parallel
	for i := 0; i < numContainers; i++ {
		c, err := cp.CreateContainer()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		cp.Lock()
		cp.containers[c.ExternalID] = c
		cp.Unlock()
	}

	return errors
}

// CreateContainer creates a new Container, returning a pointer to the container or the error that occurred.
// It does not associate it with the connection pool due to locking reasons: we don't want to lock the pool
// whilst the container is being created. We create the container first; only locking the pool when we want to
// add the container pointer.
func (cp *ContainerPool) CreateContainer() (c *cntr.Container, err error) {
	c, err = cp.manager.CreateContainer()
	if err != nil {
		log.Error(logErrorCreatingContainer, err, cp.logger)

		// TODO - add monitoring here
		return c, err
	}
	if c == nil {
		return c, errors.New(errorCreatedContainerCannotBeNil)
	}
	cp.logger.WithFields(logrus.Fields{logFieldContainerId: c.ExternalID}).Infof(logCreatedContainer)

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getNewContainersRequired(sizePool, maxSizePool, freePool, targetFreePool int) (numContainers int) {
	numContainers = 0
	if targetFreePool > freePool {
		numNewContainersRequired := targetFreePool - freePool
		amountToScale := min(numNewContainersRequired, maxSizePool - sizePool)
		if amountToScale > 0 {
			numContainers = amountToScale
		}
	}

	return numContainers
}

// scaleUpPoolIfRequired is called when a successful connection has been made, and will increase the size of the
// pool should it be required.
func (cp *ContainerPool) scaleUpPoolIfRequired() (errors []error) {
	unusedCapacity := len(cp.containers) - cp.totalContainersInUse
	if cp.settings.TargetFreeSize > unusedCapacity {
		// check to see whether we can scale
		numNewContainersRequired := cp.settings.TargetFreeSize - unusedCapacity

		amountToScale := min(numNewContainersRequired, cp.settings.MaximumSize - len(cp.containers))
		if amountToScale > 0 {
			return cp.addContainersToPool(amountToScale)
		}
	}

	return nil
}

// AssociateClientWithContainer is called whenever a client connection is made requiring a container to
// service it. This is essentially one of the 'core' function handling both associating connections with containers,
// but also scaling the up pool when new connection requests are made.
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

				// TODO errors?
				cp.scaleUpPoolIfRequired()
				return container, nil
			}

			// ...otherwise another thread has beat us to it - try and find another one
			container.Unlock()
		}
	}

	cp.monitor.WriteConnectionRejected(conn)
	return nil, errors.New(errorContainerPoolFull)
}

// DissociateClientWithContainer is called whenever a client connection disconnected.
// This is essentially one of the 'core' function handling both disassociating connections with containers,
// but also scaling the down pool when new connection requests are made.
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
