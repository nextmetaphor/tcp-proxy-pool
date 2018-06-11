package application

import (
	"os"
	"encoding/json"
)

type (
	ListenerSettings struct {
		Host      string
		Port      string
		Transport string
		CertFile  string
		KeyFile   string
	}

	PoolSettings struct {
		InitialSize    int
		MaximumSize    int
		TargetFreeSize int
	}

	MonitorSettings struct {
		Address  string
		Database string
	}

	ECSSettings struct {
		Profile                      string
		Region                       string
		Cluster                      string
		TaskDefinition               string
		LaunchType                   string
		AssignPublicIP               string
		Subnets                      []string
		SecurityGroups               []string
		MaximumContainerStartTimeSec int
	}

	Settings struct {
		Listener ListenerSettings
		Pool     PoolSettings
		Monitor  MonitorSettings
		ECS      ECSSettings
	}
)

func LoadSettings(file string) (settings *Settings, err error) {
	config, err := os.Open(file)
	defer config.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(config)
	err = jsonParser.Decode(&settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
