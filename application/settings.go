package application

import (
	"os"
	"encoding/json"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrpool"
	"github.com/nextmetaphor/tcp-proxy-pool/cntrmgr"
	"github.com/nextmetaphor/tcp-proxy-pool/monitor"
)

type (
	// ListenerSettings represents the command-line settings that can additionally be passed
	ListenerSettings struct {
		Host      string
		Port      string
		Transport string
		CertFile  string
		KeyFile   string
	}

	// Settings represents the various different parameters that can be configured using an appropriate configuration
	// file
	Settings struct {
		Listener ListenerSettings
		Pool     cntrpool.Settings
		Monitor  monitor.Settings
		ECS      cntrmgr.Settings
	}
)

// LoadSettings loads the settings file from the pathname provided. It returns the pointer of a populated Settings
// struct if this file is valid; a nil pointer and the error that occurred if this is not the case
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
