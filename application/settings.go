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
		InitialSize int
	}

	MonitorSettings struct {
		Address  string
		Database string
	}

	ECSSettings struct {
		Cluster        string
		TaskDefinition string
		Subnets        []string
		SecurityGroups []string
		AssignPublicIP string
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
