package controller

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type (
	TestContainerManager struct {

	}
)

var (
	testContainer1 = &Container{
		ExternalID: "1",
		//StartTime:  12,
	}

	testContainer2 = &Container{
		ExternalID: "2",
		//StartTime:  12,
	}

	testContainer42 = &Container{
		ExternalID: "42",
		//StartTime:  12,
	}
)

func (tcm TestContainerManager) CreateContainer() *Container {
	return testContainer42
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
		pool["1"] = testContainer1
		pool["2"] = testContainer2

		c, err := CreateContainer(&pool, tcm)

		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, testContainer42, c, "returned container incorrect")
		assert.Equal(t, 3, len(pool), "pool size incorrect")
		assert.Equal(t, testContainer42, pool[testContainer42.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer1, pool[testContainer1.ExternalID], "incorrect container in pool")
		assert.Equal(t, testContainer2, pool[testContainer2.ExternalID], "incorrect container in pool")
	})

}
