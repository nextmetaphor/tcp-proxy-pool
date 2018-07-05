package cntrpool

import (
	"errors"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
	"time"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
)

const (
	logMsgCreatedContainer         = "created container"
	logMsgDestroyedContainer       = "destroyed container"
	logMsgNewContainersRequired    = "calculating new containers required"
	logMsgOldContainersNotRequired = "calculating old containers not required"
	logMsgAlreadyScaling           = "already scaling; not considering scale-up event"
	logMsgScaleDownStatus          = "scale down status"

	logFieldContainerID              = "container-id"
	logFieldSizePool                 = "size-pool"
	logFieldMaxSizePool              = "max-size-pool"
	logFieldFreePool                 = "free-pool"
	logFieldUsedPool                 = "used-pool"
	logFieldTargetFreePool           = "target-free-pool"
	logFieldNewContainersRequired    = "new-containers-required"
	logFieldOldContainersNotRequired = "old-containers-not-required"
	logFieldLastScaleDownTime        = "last-scale-down-time"
	logFieldNextScaleDownTime        = "next-scale-down-time"
	logFieldCurrentTime              = "current-time"

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
		ScaleDownDelay int
	}

	containerStatus struct {
		sync.RWMutex
		isScaling     bool
		lastScaleDown time.Time

		usedContainers   map[string]*cntr.Container
		unusedContainers map[string]*cntr.Container
	}

	ContainerPool struct {
		// master map of all containers - do not iterate over this, it is not synchronised, use the status field
		containers map[string]*cntr.Container

		status containerStatus

		logger   *logrus.Logger
		settings Settings
		manager  cntrmgr.ContainerManager
		monitor  monitor.Client
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
			lastScaleDown:    time.Now(),
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

func (cp *ContainerPool) addContainersToPool(numContainers int) (e []error) {
	// TODO obvs better to create containers in parallel
	for i := 0; i < numContainers; i++ {
		c, err := cp.CreateContainer()
		if err != nil {
			e = append(e, err)
			continue
		}

		cp.status.Lock()
		{
			// there is a chance that the number of used containers in the pool has changed which would mean that
			// we'd exceed the maximum size of the pool by adding our new container to it.
			// now we've got the lock, check if this is the case, and destroy the container if necessary
			if len(cp.containers) < cp.settings.MaximumSize {
				cp.status.unusedContainers[c.ExternalID] = c
				cp.containers[c.ExternalID] = c
			} else {
				err := cp.DestroyContainer(c)
				if err != nil {
					e = append(e, err)
				}
			}

		}
		cp.status.Unlock()
	}

	return e
}

func (cp *ContainerPool) removeContainersFromPool(numContainers int) (errors []error) {
	containersToRemove := make(map[string]*cntr.Container, numContainers)

	cp.status.Lock()
	{
		for cID, c := range cp.status.unusedContainers {
			if len(containersToRemove) >= numContainers {
				break
			}

			containersToRemove[cID] = c

			delete(cp.status.unusedContainers, cID)
			delete(cp.containers, cID)

			// shouldn't be possible but just in case...
			delete(cp.status.usedContainers, cID)
		}
	}
	cp.status.Unlock()

	// at this point these containers are no longer referenced from the pool so can be destroyed
	// without a lock
	for _, c := range containersToRemove {
		e := cp.DestroyContainer(c)
		if e != nil {
			errors = append(errors, e)
		}
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
	cp.logger.WithFields(logrus.Fields{logFieldContainerID: c.ExternalID}).Infof(logMsgCreatedContainer)

	return c, nil
}

func (cp *ContainerPool) DestroyContainer(c *cntr.Container) (err error) {
	err = cp.manager.DestroyContainer(c.ExternalID)

	cp.logger.WithFields(logrus.Fields{logFieldContainerID: c.ExternalID}).Infof(logMsgDestroyedContainer)

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
	amountToScale := 0
	cp.status.RLock()
	if cp.status.isScaling {
		cp.logger.Debug(logMsgAlreadyScaling)
		cp.status.RUnlock()
		return errors
	}

	cp.status.isScaling = true
	amountToScale = getNewContainersRequired(len(cp.containers), cp.settings.MaximumSize, len(cp.status.unusedContainers), cp.settings.TargetFreeSize)
	cp.logger.WithFields(logrus.Fields{
		logFieldSizePool:              len(cp.containers),
		logFieldMaxSizePool:           cp.settings.MaximumSize,
		logFieldFreePool:              len(cp.status.unusedContainers),
		logFieldTargetFreePool:        cp.settings.TargetFreeSize,
		logFieldNewContainersRequired: amountToScale,
	}).Debugf(logMsgNewContainersRequired)
	cp.status.RUnlock()

	if amountToScale > 0 {
		errors = cp.addContainersToPool(amountToScale)
	}

	cp.status.RLock()
	cp.status.isScaling = false
	cp.status.RUnlock()

	return errors
}

func (cp *ContainerPool) scaleDownPoolIfRequired() (errors []error) {
	lastScaleDownTime := cp.status.lastScaleDown
	nextScaleDownTime := lastScaleDownTime.Add(time.Duration(cp.settings.ScaleDownDelay) * time.Second)
	currentTime := time.Now()

	cp.logger.WithFields(logrus.Fields{
		logFieldLastScaleDownTime: lastScaleDownTime,
		logFieldNextScaleDownTime: nextScaleDownTime,
		logFieldCurrentTime:       currentTime,
	}).Debugf(logMsgScaleDownStatus)

	// only consider scaling down if necessary
	if (cp.status.lastScaleDown.IsZero()) || (currentTime.Before(nextScaleDownTime)) {
		return nil
	}

	amountToScale := 0
	cp.status.RLock()
	{
		if !cp.status.isScaling {
			cp.status.isScaling = true
			amountToScale = getOldContainersNoLongerRequired(len(cp.status.usedContainers), cp.settings.TargetFreeSize)
			cp.logger.WithFields(logrus.Fields{
				logFieldUsedPool:                 len(cp.status.usedContainers),
				logFieldTargetFreePool:           cp.settings.TargetFreeSize,
				logFieldOldContainersNotRequired: amountToScale,
			}).Debugf(logMsgOldContainersNotRequired)

		}
	}
	cp.status.RUnlock()

	if amountToScale > 0 {
		errors = cp.removeContainersFromPool(amountToScale)
		cp.status.lastScaleDown = time.Now()
	}

	cp.status.RLock()
	cp.status.isScaling = false
	cp.status.RUnlock()

	return errors
}

// AssociateClientWithContainer is called whenever a client connection is made requiring a container to
// service it. This is essentially one of the 'core' function handling both associating connections with containers,
// but also scaling the up pool when new connection requests are made.
func (cp *ContainerPool) AssociateClientWithContainer(conn net.Conn) (*cntr.Container, error) {
	var cID = ""
	var c *cntr.Container

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
	{
		c.ConnectionFromClient = nil

		cp.status.unusedContainers[c.ExternalID] = c
		delete(cp.status.usedContainers, c.ExternalID)

		cp.monitor.WriteConnectionPoolStats(serverConn, len(cp.status.usedContainers), len(cp.containers))
	}
	cp.status.Unlock()

	cp.scaleDownPoolIfRequired()
}

func ConnectClientToContainer(c *cntr.Container) error {
	conn, err := net.Dial("tcp", c.IPAddress+":"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}

	c.ConnectionToContainer = conn

	return nil
}
