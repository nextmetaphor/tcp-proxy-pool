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

	containerStatus struct {
		sync.RWMutex

		usedContainers   map[string]*cntr.Container
		unusedContainers map[string]*cntr.Container
	}

	ContainerPool struct {
		//sync.RWMutex

		// master map of all containers
		containers map[string]*cntr.Container

		status containerStatus

		logger   *logrus.Logger
		settings Settings
		manager  cntrmgr.ContainerManager
		monitor  monitor.Client

		// totalContainersInUse can be calculated from containers but included here for speed purposes
		//totalContainersInUse int
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
		status: containerStatus{
			unusedContainers: make(map[string]*cntr.Container),
			usedContainers:   make(map[string]*cntr.Container),
		},
		logger:   l,
		settings: s,
		manager:  cm,
		monitor:  m,
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
		//cp.Lock()
		cp.containers[c.ExternalID] = c
		//cp.Unlock()
	}

	return errors
}

func (cp *ContainerPool) removeContainersFromPool(numContainers int) (errors []error) {
	containersToRemove := make(map[string]*cntr.Container, numContainers)

	cp.status.Lock()
	containersRemoved := 0
	for cID, c := range cp.status.unusedContainers {
		containersRemoved++

		containersToRemove[cID] = c

		delete(cp.status.unusedContainers, cID)
		delete(cp.containers, cID)
		if containersRemoved >= numContainers {
			break
		}
	}
	cp.status.Unlock()

	for _, c := range containersToRemove {
		errors = append(errors, cp.DestroyContainer(c))
	}

	//
	//for _, container := range cp.containers {
	//	if numContainers < len(containersToRemove) {
	//		// find the first container with no current connection from the client which is not in the process
	//		// of being removed from the pool
	//		if (container.ConnectionFromClient == nil) && !container.IsBeingRemoved {
	//			container.Lock()
	//			if (container.ConnectionFromClient == nil) && !container.IsBeingRemoved {
	//				container.IsBeingRemoved = true
	//				containersToRemove = append(containersToRemove, container.ExternalID)
	//			}
	//			container.Unlock()
	//		}
	//	}
	//}

	//cp.Lock()
	//defer cp.Unlock()
	//for _, cId := range containersToRemove {
	//	errors = append(errors, cp.DestroyContainer(cId))
	//}

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

	cp.status.Lock()
	cp.status.unusedContainers[c.ExternalID] = c
	cp.status.Unlock()

	return c, nil
}

func (cp *ContainerPool) DestroyContainer(c *cntr.Container) (err error) {
	err = cp.manager.DestroyContainer(c.ExternalID)

	cp.status.Lock()
	cp.status.unusedContainers[c.ExternalID] = c
	delete(cp.status.usedContainers, c.ExternalID)
	cp.status.Unlock()

	// TODO - add monitoring here
	return err
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
		amountToScale := min(numNewContainersRequired, maxSizePool-sizePool)
		if amountToScale > 0 {
			numContainers = amountToScale
		}
	}

	return numContainers
}

func getOldContainersNoLongerRequired(freePool, targetFreePool int) (numContainers int) {
	numContainers = freePool - targetFreePool
	if numContainers < 0 {
		numContainers = 0
	}

	return numContainers
}

// scaleUpPoolIfRequired is called when a successful connection has been made, and will increase the size of the
// pool should it be required.
func (cp *ContainerPool) scaleUpPoolIfRequired() (errors []error) {
	amountToScale := getNewContainersRequired(len(cp.containers), cp.settings.MaximumSize, len(cp.status.unusedContainers), cp.settings.TargetFreeSize)
	if amountToScale > 0 {
		return cp.addContainersToPool(amountToScale)
	}

	return nil
}

func (cp *ContainerPool) scaleDownPoolIfRequired() (errors []error) {
	amountToScale := getOldContainersNoLongerRequired(len(cp.status.usedContainers), cp.settings.TargetFreeSize)
	if amountToScale > 0 {
		return cp.removeContainersFromPool(amountToScale)
	}

	return nil
}

// AssociateClientWithContainer is called whenever a client connection is made requiring a container to
// service it. This is essentially one of the 'core' function handling both associating connections with containers,
// but also scaling the up pool when new connection requests are made.
func (cp *ContainerPool) AssociateClientWithContainer(conn net.Conn) (*cntr.Container, error) {
	var cID = ""
	var c *cntr.Container = nil

	cp.status.Lock()

	// use a loop to simply get a single element in the map
	for cID, c = range cp.status.unusedContainers {
		// associate the connection with the container
		c.ConnectionFromClient = conn

		// add this container to the "used" map and remove from the "unused" map
		cp.status.usedContainers[cID] = c
		delete(cp.status.unusedContainers, cID)

		cp.monitor.WriteConnectionPoolStats(conn, len(cp.status.usedContainers), len(cp.containers))
		break
	}
	cp.status.Unlock()

	if c != nil {
		cp.monitor.WriteConnectionAccepted(conn)
		cp.scaleUpPoolIfRequired()
		return c, nil
	}

	//for _, container := range cp.containers {
	//	// find the first container with no current connection from the client which is not in the process
	//	// of being removed from the pool
	//	if (container.ConnectionFromClient == nil) && !container.IsBeingRemoved {
	//		container.Lock()
	//		if (container.ConnectionFromClient == nil) && !container.IsBeingRemoved {
	//			container.ConnectionFromClient = conn
	//
	//			cp.Lock()
	//			cp.totalContainersInUse++
	//			cp.monitor.WriteConnectionPoolStats(conn, cp.totalContainersInUse, len(cp.containers))
	//			cp.Unlock()
	//
	//			cp.monitor.WriteConnectionAccepted(conn)
	//
	//			container.Unlock()
	//
	//			// TODO errors?
	//			cp.scaleUpPoolIfRequired()
	//			return container, nil
	//		}
	//
	//		// ...otherwise another thread has beat us to it - try and find another one
	//		container.Unlock()
	//	}
	//}

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

	cp.status.Lock()

	cp.status.unusedContainers[c.ExternalID] = c
	delete(cp.status.usedContainers, c.ExternalID)

	//c.ConnectionToContainer = nil
	c.ConnectionFromClient = nil

	cp.monitor.WriteConnectionPoolStats(serverConn, len(cp.status.usedContainers), len(cp.containers))

	cp.status.Unlock()

	//c.Lock()
	//defer c.Unlock()

	//cp.Lock()
	//cp.totalContainersInUse--
	//cp.monitor.WriteConnectionPoolStats(serverConn, cp.totalContainersInUse, len(cp.containers))
	//cp.Unlock()

	//cp.scaleDownPoolIfRequired()
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
