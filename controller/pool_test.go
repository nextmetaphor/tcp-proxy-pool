package controller

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type (
	TestContainerManager struct {

	}
)

func (tcm TestContainerManager) CreateContainer() Container {
	return Container{
		ExternalID: "42",
		//StartTime:  12,
	}
}

func Test_CreateContainer(t *testing.T) {
	tcm  := TestContainerManager{}

	t.Run("NilPool", func(t *testing.T) {
		id, err := CreateContainer(nil, tcm)
		assert.Empty(t, id, "empty id string should have been returned")
		assert.NotNil(t, err, "error should have been returned")
	})

	t.Run("EmptyPool", func(t *testing.T) {
		pool := make(ContainerPool, 0)
		id, err := CreateContainer(&pool, tcm)
		assert.Nil(t, err, "nil error should have been returned")
		assert.Equal(t, "42", id, "container id incorrect")
	})

}
