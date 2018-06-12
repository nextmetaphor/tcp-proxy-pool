package application

import (
	"os"
	"encoding/json"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrpool"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
)

type (
	ListenerSettings struct {
		Host      string
		Port      string
		Transport string
		CertFile  string
		KeyFile   string
	}

	Settings struct {
		Listener ListenerSettings
		Pool     cntrpool.Settings
		Monitor  monitor.Settings
		ECS      cntrmgr.Settings
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
