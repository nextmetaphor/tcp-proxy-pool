package controller

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/nextmetaphor/tcp-proxy-pool/container"
)

type (
	TestContainerManager struct {}

	TestNilContainerManager struct {}
)

var (
	testContainer1 = &container.Container{
		ExternalID: "1",
	}

	testContainer2 = &container.Container{
		ExternalID: "2",
	}

	testContainer42 = &container.Container{
		ExternalID: "42",
	}
)

func (tcm TestNilContainerManager) CreateContainer() (*container.Container, error) {
	return nil, nil
}

func (cm TestNilContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}

func (tcm TestContainerManager) CreateContainer() (*container.Container, error) {
	return testContainer42, nil
}

func (tcm TestContainerManager) DestroyContainer(externalID string) (error) {
	return nil
}


func Test_CreateContainer(t *testing.T) {
	tcm  := TestContainerManager{}

	t.Run("NilPool", func(t *testing.T) {
		c, err := CreateContainer(nil, tcm)
		assert.Nil(t, c, "nil container should have been returned")
		assert.NotNil(t, err, "error should have been returned")
	})

	t.Run("EmptyPool", func(t *testing.T) {
		pool := make(ContainerPool, 0)
		c, err := CreateContainer(&pool, tcm)
		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 1, len(pool), "pool size incorrect")
		assert.Equal(t, testContainer42, pool[testContainer42.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNewContainer", func(t *testing.T) {
		pool := make(ContainerPool, 0)
		pool[testContainer1.ExternalID] = testContainer1
		pool[testContainer2.ExternalID] = testContainer2

		c, err := CreateContainer(&pool, tcm)

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool), "pool size incorrect")
		assert.Equal(t, testContainer42, pool[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolExistingContainer", func(t *testing.T) {
		pool := make(ContainerPool, 0)
		pool[testContainer1.ExternalID] = testContainer1
		pool[testContainer2.ExternalID] = testContainer2
		pool[testContainer42.ExternalID] = testContainer42

		c, err := CreateContainer(&pool, tcm)

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool), "pool size incorrect")
		assert.Equal(t, testContainer42, pool[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool[testContainer2.ExternalID], "incorrect container in pool")
	})

	t.Run("ExistingPoolNilContainer", func(t *testing.T) {
		pool := make(ContainerPool, 0)
		pool[testContainer1.ExternalID] = testContainer1
		pool[testContainer2.ExternalID] = testContainer2
		pool[testContainer42.ExternalID] = testContainer42

		c, err := CreateContainer(&pool, TestNilContainerManager{})

		assert.NotNil(t, err, "error expected")
		assert.Nil(t, c, "nil container expected")
		assert.Equal(t, 3, len(pool), "pool size incorrect")
		assert.Equal(t, testContainer42, pool[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool[testContainer2.ExternalID], "incorrect container in pool")
	})
}

func Test_DestroyContainer(t *testing.T) {
	tcm  := TestContainerManager{}

	t.Run("NilPool", func(t *testing.T) {
		err := DestroyContainer("42", nil, tcm)
		assert.NotNil(t, err, "error should have been returned")
	})

}
