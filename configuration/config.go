package configuration

type (
	ListenerSettings struct {
		Host string
		Port string
		Transport string
		CertFile string
		KeyFile string
	}

	PoolSettings struct {
		InitialSize int
		DesiredFreeCapacity int
		IncrementAmount int

	}

	MonitorSettings struct {
		Database string

	}

	ApplicationSettings struct {
		Listener ListenerSettings
		Pool     PoolSettings
		Monitor  MonitorSettings
	}
)
