package cntrpool

import (
	"errors"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const (
	errorInitialiseError  = "some error"
	errorDestroyContainer = "error destroying container"
)

type (
	Test42ContainerManager struct{}

	TestNilContainerManager struct{}

	TestIncrementContainerManager struct{}

	TestCreateErrContainerManager struct{}

	TestDestroyErrContainerManager struct{}
)

var (
	testContainer1 = &cntr.Container{
		ExternalID: "1",
	}

	testContainer2 = &cntr.Container{
		ExternalID: "2",
	}

	testContainer42 = &cntr.Container{
		ExternalID: "42",
	}

	nextContainerID = 0
)

func (cm TestNilContainerManager) CreateContainer() (*cntr.Container, error) {
	return nil, nil
}

func (cm TestNilContainerManager) DestroyContainer(externalID string) error {
	return nil
}

func (cm Test42ContainerManager) CreateContainer() (*cntr.Container, error) {
	return testContainer42, nil
}

func (cm Test42ContainerManager) DestroyContainer(externalID string) error {
	return nil
}

func (cm TestIncrementContainerManager) CreateContainer() (*cntr.Container, error) {
	nextContainerID++
	return &cntr.Container{ExternalID: strconv.Itoa(nextContainerID)}, nil
}

func (cm TestIncrementContainerManager) DestroyContainer(externalID string) error {
	return nil
}

func (cm TestCreateErrContainerManager) CreateContainer() (*cntr.Container, error) {
	return nil, errors.New(errorInitialiseError)
}

func (cm TestCreateErrContainerManager) DestroyContainer(externalID string) error {
	return nil
}

func (cm TestDestroyErrContainerManager) CreateContainer() (*cntr.Container, error) {
	nextContainerID++
	return &cntr.Container{ExternalID: strconv.Itoa(nextContainerID)}, nil
}

func (cm TestDestroyErrContainerManager) DestroyContainer(externalID string) error {
	return errors.New(errorDestroyContainer)
}

func Test_CreateContainer(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, logger)

	tcm := Test42ContainerManager{}
	cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)

	t.Run("EmptyPool", func(t *testing.T) {
		cp.containers = make(map[string]*cntr.Container)
		c, err := cp.createContainer()
		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		// check we don't add the created container to the pool
		assert.Equal(t, 0, len(cp.containers), "pool size incorrect")
	})

	t.Run("ExistingPoolNewContainer", func(t *testing.T) {
		cp.containers = make(map[string]*cntr.Container)
		cp.containers[testContainer1.ExternalID] = testContainer1
		cp.containers[testContainer2.ExternalID] = testContainer2

		c, err := cp.createContainer()

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 2, len(cp.containers), "pool size incorrect")
		// check we don't add the created container to the pool
		assert.Equal(t, testContainer1, cp.containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, cp.containers[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolExistingContainer", func(t *testing.T) {
		pool := ContainerPool{
			containers: make(map[string]*cntr.Container),
		}
		pool.containers[testContainer1.ExternalID] = testContainer1
		pool.containers[testContainer2.ExternalID] = testContainer2
		pool.containers[testContainer42.ExternalID] = testContainer42

		c, err := cp.createContainer()

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool.containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.containers[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNilContainer", func(t *testing.T) {
		pool := ContainerPool{
			containers: make(map[string]*cntr.Container),
		}
		pool.containers[testContainer1.ExternalID] = testContainer1
		pool.containers[testContainer2.ExternalID] = testContainer2
		pool.containers[testContainer42.ExternalID] = testContainer42

		tcm := TestNilContainerManager{}
		cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)
		c, err := cp.createContainer()

		assert.NotNil(t, err, "error expected")
		assert.Nil(t, c, "nil container expected")
		assert.Equal(t, 3, len(pool.containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.containers[testContainer2.ExternalID], "incorrect container in pool")
	})
}

func Test_DestroyContainer(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, logger)

	t.Run("ContainerInPool", func(t *testing.T) {
		pool := ContainerPool{
			containers: make(map[string]*cntr.Container),
		}
		pool.containers[testContainer1.ExternalID] = testContainer1
		pool.containers[testContainer2.ExternalID] = testContainer2
		pool.containers[testContainer42.ExternalID] = testContainer42

		tcm := TestDestroyErrContainerManager{}
		cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)
		err := cp.destroyContainer(pool.containers[testContainer42.ExternalID])

		assert.NotNil(t, err, "error expected")
		assert.Equal(t, errors.New(errorDestroyContainer), err)
		assert.Equal(t, 3, len(pool.containers), "pool size incorrect")
	})
}

func Test_CreateContainerPool(t *testing.T) {
	l, _ := test.NewNullLogger()
	l.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, l)
	tcm := Test42ContainerManager{}
	s := Settings{}

	t.Run("NilLogger", func(t *testing.T) {
		cp, err := CreateContainerPool(tcm, s, nil, *m)
		assert.Equal(t, errors.New(errorLoggerNil), err)
		assert.Nil(t, cp)
	})

	t.Run("NilContainerManager", func(t *testing.T) {
		cp, err := CreateContainerPool(nil, s, l, *m)
		assert.Equal(t, errors.New(errorContainerManagerNil), err)
		assert.Nil(t, cp)
	})

	t.Run("ValidCall", func(t *testing.T) {
		cp, err := CreateContainerPool(tcm, s, l, *m)
		assert.Equal(t, l, cp.logger)
		assert.Equal(t, s, cp.settings)
		assert.Equal(t, m, &cp.monitor)
		assert.Equal(t, tcm, cp.manager)

		assert.Nil(t, err)
	})
}

func Test_InitialisePool(t *testing.T) {
	l, h := test.NewNullLogger()
	l.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, l)
	tcm := TestIncrementContainerManager{}

	t.Run("PoolSizeOf0", func(t *testing.T) {
		s := Settings{InitialSize: 0, MaximumSize: 10}
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		err := cp.InitialisePool()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(cp.containers))
	})

	t.Run("PoolSizeOf1", func(t *testing.T) {
		s := Settings{InitialSize: 1, MaximumSize: 10}
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		err := cp.InitialisePool()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(cp.containers))
	})

	t.Run("PoolSizeOf10", func(t *testing.T) {
		h.Reset()
		s := Settings{InitialSize: 10, MaximumSize: 10}
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		err := cp.InitialisePool()
		assert.Nil(t, err)
		assert.Equal(t, 10, len(cp.containers))
		assert.Equal(t, 10, len(h.AllEntries()))
		for _, tl := range h.AllEntries() {
			assert.Equal(t, logrus.InfoLevel, tl.Level)
			assert.Contains(t, logMsgCreatedContainer, tl.Message)
		}
	})

	t.Run("ErrorCreatingContainer", func(t *testing.T) {
		h.Reset()
		s := Settings{InitialSize: 3, MaximumSize: 10}
		cp, _ := CreateContainerPool(TestCreateErrContainerManager{}, s, l, *m)
		err := cp.InitialisePool()
		assert.Equal(t, []error{errors.New(errorInitialiseError), errors.New(errorInitialiseError), errors.New(errorInitialiseError)}, err)
		assert.Equal(t, 0, len(cp.containers))
		assert.Equal(t, 3, len(h.AllEntries()))
		for _, tl := range h.AllEntries() {
			assert.Equal(t, logrus.ErrorLevel, tl.Level)
			assert.Contains(t, logErrorCreatingContainer, tl.Message)
		}
	})
}

func Test_GetNewContainersRequired(t *testing.T) {
	t.Run("ZeroSizeCreateAllTarget", func(t *testing.T) {
		i := getNewContainersRequired(0, 10, 0, 5)
		assert.Equal(t, 5, i)
	})
	t.Run("ZeroSizeCreatePartialTarget", func(t *testing.T) {
		i := getNewContainersRequired(0, 10, 3, 5)
		assert.Equal(t, 2, i)
	})
	t.Run("ZeroSizeCreateZeroTarget", func(t *testing.T) {
		i := getNewContainersRequired(0, 10, 5, 5)
		assert.Equal(t, 0, i)
	})

	t.Run("NonZeroSizeCreateAllTarget", func(t *testing.T) {
		i := getNewContainersRequired(3, 10, 0, 5)
		assert.Equal(t, 5, i)
	})
	t.Run("NonZeroSizeCreatePartialTarget", func(t *testing.T) {
		i := getNewContainersRequired(3, 10, 3, 5)
		assert.Equal(t, 2, i)
	})
	t.Run("NonZeroSizeCreateZeroTarget", func(t *testing.T) {
		i := getNewContainersRequired(3, 10, 5, 5)
		assert.Equal(t, 0, i)
	})

	t.Run("NonZeroSizeCreateAllTargetWithMaxRestriction", func(t *testing.T) {
		i := getNewContainersRequired(7, 10, 0, 5)
		assert.Equal(t, 3, i)
	})
	t.Run("NonZeroSizeCreatePartialTargetWithMaxRestriction", func(t *testing.T) {
		i := getNewContainersRequired(7, 10, 1, 5)
		assert.Equal(t, 3, i)
	})

	t.Run("IdenticalNegatives", func(t *testing.T) {
		i := getNewContainersRequired(-10, -10, -10, -10)
		assert.Equal(t, 0, i)
	})
	t.Run("DifferentNegatives", func(t *testing.T) {
		i := getNewContainersRequired(-7, -10, -2, -5)
		assert.Equal(t, 0, i)
	})
}

func Test_GetOldContainersNoLongerRequired(t *testing.T) {
	t.Run("BothNegative", func(t *testing.T) {
		i := getOldContainersNoLongerRequired(-15, -7)
		assert.Equal(t, 0, i)
	})

	t.Run("BothZero", func(t *testing.T) {
		i := getOldContainersNoLongerRequired(0, 0)
		assert.Equal(t, 0, i)
	})

	t.Run("FreeLessThanTarget", func(t *testing.T) {
		i := getOldContainersNoLongerRequired(2, 5)
		assert.Equal(t, 0, i)
	})

	t.Run("FreeEqualToTarget", func(t *testing.T) {
		i := getOldContainersNoLongerRequired(2, 2)
		assert.Equal(t, 0, i)
	})

	t.Run("FreeGreaterThanTarget", func(t *testing.T) {
		i := getOldContainersNoLongerRequired(5, 2)
		assert.Equal(t, 3, i)
	})
}

func Test_AddContainersToPool(t *testing.T) {
	l, _ := test.NewNullLogger()
	l.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, l)
	tcm := TestIncrementContainerManager{}
	s := Settings{InitialSize: 0, MaximumSize: 10}

	t.Run("AddZeroContainers", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		errors := cp.addContainersToPool(0)
		assert.Nil(t, errors)
		assert.Equal(t, 0, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 0, len(cp.status.unusedContainers))
	})

	t.Run("AddSingleContainer", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		errors := cp.addContainersToPool(1)
		assert.Nil(t, errors)
		assert.Equal(t, 1, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 1, len(cp.status.unusedContainers))
	})

	t.Run("AddMultipleContainers", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		errors := cp.addContainersToPool(9)
		assert.Nil(t, errors)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))
	})

	t.Run("AddMultipleCreateErroringContainers", func(t *testing.T) {
		tcm := TestCreateErrContainerManager{}
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		errors := cp.addContainersToPool(9)
		assert.NotNil(t, errors)
		assert.Equal(t, 9, len(errors))
		for _, e := range errors {
			assert.Equal(t, errorInitialiseError, e.Error())
		}
		assert.Equal(t, 0, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 0, len(cp.status.unusedContainers))
	})

	t.Run("AddMultipleDestroyErroringContainers", func(t *testing.T) {
		tcm := TestDestroyErrContainerManager{}
		s := Settings{InitialSize: 0, MaximumSize: 0}
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		errors := cp.addContainersToPool(9)
		assert.NotNil(t, errors)
		assert.Equal(t, 9, len(errors))
		for _, e := range errors {
			assert.Equal(t, errorDestroyContainer, e.Error())
		}
		assert.Equal(t, 0, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 0, len(cp.status.unusedContainers))
	})
}

func Test_RemoveContainersFromPool(t *testing.T) {
	l, _ := test.NewNullLogger()
	l.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, l)
	tcm := TestIncrementContainerManager{}
	s := Settings{InitialSize: 0, MaximumSize: 10}

	t.Run("RemoveZeroContainers", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)

		// First add several containers and check they are created as expected
		errors := cp.addContainersToPool(9)
		assert.Nil(t, errors)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))

		err := cp.removeContainersFromPool(0)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))
		assert.Nil(t, err)
	})

	t.Run("RemoveSingleContainer", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)

		// First add several containers and check they are created as expected
		errors := cp.addContainersToPool(9)
		assert.Nil(t, errors)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))

		err := cp.removeContainersFromPool(1)
		assert.Equal(t, 8, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 8, len(cp.status.unusedContainers))
		assert.Nil(t, err)
	})

	t.Run("RemoveMultipleContainers", func(t *testing.T) {
		cp, _ := CreateContainerPool(tcm, s, l, *m)

		// First add several containers and check they are created as expected
		errors := cp.addContainersToPool(9)
		assert.Nil(t, errors)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))

		err := cp.removeContainersFromPool(5)
		assert.Equal(t, 4, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 4, len(cp.status.unusedContainers))
		assert.Nil(t, err)
	})

	t.Run("RemoveMultipleDestroyErroringContainers", func(t *testing.T) {
		tcm := TestDestroyErrContainerManager{}
		cp, _ := CreateContainerPool(tcm, s, l, *m)

		// First add several containers and check they are created as expected
		errors := cp.addContainersToPool(9)
		assert.Nil(t, errors)
		assert.Equal(t, 9, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 9, len(cp.status.unusedContainers))

		err := cp.removeContainersFromPool(5)
		assert.Equal(t, 4, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 4, len(cp.status.unusedContainers))
		assert.NotNil(t, err)
		assert.Equal(t, 5, len(err))
		for _, e := range errors {
			assert.Equal(t, errorDestroyContainer, e.Error())
		}
	})

	t.Run("RemoveSingleContainerFromUsed", func(t *testing.T) {
		tcm := Test42ContainerManager{}
		cp, _ := CreateContainerPool(tcm, s, l, *m)

		// First add container and check it is created as expected
		errors := cp.addContainersToPool(1)

		//this situation should never occur under normal operation
		cp.status.usedContainers["42"] = cp.containers["42"]

		assert.Nil(t, errors)
		assert.Equal(t, 1, len(cp.containers))
		assert.Equal(t, 1, len(cp.status.usedContainers))
		assert.Equal(t, 1, len(cp.status.unusedContainers))

		err := cp.removeContainersFromPool(1)
		assert.Equal(t, 0, len(cp.containers))
		assert.Equal(t, 0, len(cp.status.usedContainers))
		assert.Equal(t, 0, len(cp.status.unusedContainers))
		assert.Nil(t, err)
	})
}

func Test_scaleUpPoolIfRequired(t *testing.T) {
	l, h := test.NewNullLogger()
	l.Level = logrus.DebugLevel
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, l)
	tcm := TestIncrementContainerManager{}
	s := Settings{InitialSize: 0, MaximumSize: 10}

	t.Run("AlreadyScaling", func(t *testing.T) {
		h.Reset()
		cp, _ := CreateContainerPool(tcm, s, l, *m)
		cp.status.isScaling = true

		assert.Equal(t, len(cp.containers), 0)
		err := cp.scaleUpPoolIfRequired()
		assert.Nil(t, err)
		assert.Equal(t, true, cp.status.isScaling)

		assert.Equal(t, logrus.DebugLevel, l.Level)
		assert.Equal(t, 1, len(h.AllEntries()))
		assert.Contains(t, logMsgAlreadyScaling, h.LastEntry().Message)
	})
}