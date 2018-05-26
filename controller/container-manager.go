package controller

import "time"

type (
	ECSContainerManager struct {}
)

func (cm ECSContainerManager) CreateContainer() Container {
	return Container{
		ExternalId: "42",
		StartTime:  time.Now(),
	}
}


