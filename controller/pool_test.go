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
	testContainer = &Container{
		ExternalID: "42",
		//StartTime:  12,
	}
)

func (tcm TestContainerManager) CreateContainer() *Container {
	return testContainer
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
		assert.Equal(t, testContainer, c, "returned container incorrect")
	})
}
