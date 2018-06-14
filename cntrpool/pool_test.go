package cntrpool

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/nextmetaphor/tcp-proxy-pool/cntr"
	"github.com/sirupsen/logrus/hooks/test"
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
	tcm  := TestContainerManager{}
	logger, hook := test.NewNullLogger()
	cp := CreateContainerPool(tcm, )

	t.Run("NilPool", func(t *testing.T) {
		c, err := cp.CreateContainer(nil, tcm)
		assert.Nil(t, c, "nil container should have been returned")
		assert.NotNil(t, err, "error should have been returned")
	})

	t.Run("EmptyPool", func(t *testing.T) {
		pool := ContainerPool{
			Containers: make(map[string]*cntr.Container),
		}
		c, err := cp.CreateContainer()
		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 1, len(pool.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.Containers[testContainer42.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNewContainer", func(t *testing.T) {
		pool := ContainerPool{
			Containers: make(map[string]*cntr.Container),
		}
		pool.Containers[testContainer1.ExternalID] = testContainer1
		pool.Containers[testContainer2.ExternalID] = testContainer2

		c, err := cp.CreateContainer()

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.Containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.Containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.Containers[testContainer2.ExternalID], "incorrect container in pool")
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

		c, err := cp.CreateContainer(&pool, TestNilContainerManager{})

		assert.NotNil(t, err, "error expected")
		assert.Nil(t, c, "nil container expected")
		assert.Equal(t, 3, len(pool.Containers), "pool size incorrect")
		assert.Equal(t, testContainer42, pool.Containers[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool.Containers[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool.Containers[testContainer2.ExternalID], "incorrect container in pool")
	})
}

func Test_DestroyContainer(t *testing.T) {
	tcm  := TestContainerManager{}

	t.Run("NilPool", func(t *testing.T) {
		err := DestroyContainer("42", nil, tcm)
		assert.NotNil(t, err, "error should have been returned")
	})
}

func Test_CreateContainerPool(t *testing.T) {
}
