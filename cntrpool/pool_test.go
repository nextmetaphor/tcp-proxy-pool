package cntrpool

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
)

type (
	TestContainerManager struct {}

	TestNilContainerManager struct {}
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
)

func (cm TestNilContainerManager) CreateContainer() (*cntr.Container, error) {
	return nil, nil
}

func (cm TestNilContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}

func (cm TestContainerManager) CreateContainer() (*cntr.Container, error) {
	return testContainer42, nil
}

func (cm TestContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}


func Test_CreateContainer(t *testing.T) {
	logger, _ := test.NewNullLogger()
	m := monitor.CreateMonitor(monitor.Settings{Address: "something"}, logger)

	tcm  := TestContainerManager{}
	cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)

	t.Run("EmptyPool", func(t *testing.T) {
		cp.Containers = make(map[string]*cntr.Container)
		c, err := cp.CreateContainer()
		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 1, len(cp.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, cp.Containers[testContainer42.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNewContainer", func(t *testing.T) {
		cp.Containers = make(map[string]*cntr.Container)
		cp.Containers[testContainer1.ExternalID] = testContainer1
		cp.Containers[testContainer2.ExternalID] = testContainer2

		c, err := cp.CreateContainer()

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(cp.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, cp.Containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, cp.Containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, cp.Containers[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolExistingContainer", func(t *testing.T) {
		pool := ContainerPool{
			Containers: make(map[string]*cntr.Container),
		}
		pool.Containers[testContainer1.ExternalID] = testContainer1
		pool.Containers[testContainer2.ExternalID] = testContainer2
		pool.Containers[testContainer42.ExternalID] = testContainer42

		c, err := cp.CreateContainer()

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.Containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.Containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.Containers[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNilContainer", func(t *testing.T) {
		pool := ContainerPool{
			Containers: make(map[string]*cntr.Container),
		}
		pool.Containers[testContainer1.ExternalID] = testContainer1
		pool.Containers[testContainer2.ExternalID] = testContainer2
		pool.Containers[testContainer42.ExternalID] = testContainer42

		tcm  := TestNilContainerManager{}
		cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)
		c, err := cp.CreateContainer()

		assert.NotNil(t, err, "error expected")
		assert.Nil(t, c, "nil container expected")
		assert.Equal(t, 3, len(pool.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.Containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.Containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.Containers[testContainer2.ExternalID], "incorrect container in pool")
	})
}

func Test_DestroyContainer(t *testing.T) {
	// TODO
	//tcm  := TestContainerManager{}
	//logger, _ := test.NewNullLogger()
	//m := monitor.CreateMonitor(monitor.Settings{Address: "Something"}, logger)
	//
	//cp, _ := CreateContainerPool(tcm, Settings{}, logger, *m)
}

func Test_CreateContainerPool(t *testing.T) {
}
