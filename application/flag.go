package application

import (
	"github.com/alecthomas/kingpin"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
)

const (
	cAppName            = "tcp-proxy-pool"
	cAppNameDescription = "TODO"

	HostNameFlag          = "hostname"
	PortNameFlag          = "port"
	CertFileFlag          = "certFile"
	KeyFileFlag           = "keyFile"
	TransportProtocolFlag = "transport"
	LogLevelFlag          = "logLevel"
)

type (
	CommandLineFlag struct {
		ShortName    rune
		Help         string
		DefaultValue string
		FlagValue    *string
	}

	CommandLineFlags map[string]*CommandLineFlag
)

func CreateFlags() CommandLineFlags {
	return CommandLineFlags{
		HostNameFlag: &CommandLineFlag{
			ShortName:    'h',
			DefaultValue: "",
			Help:         "hostname to bind to",
		},
		PortNameFlag: &CommandLineFlag{
			ShortName:    'p',
			DefaultValue: "8443",
			Help:         "port to bind to",
		},
		CertFileFlag: &CommandLineFlag{
			ShortName:    'c',
			DefaultValue: cAppName + ".crt",
			Help:         "TLS certificate file",
		},
		KeyFileFlag: &CommandLineFlag{
			ShortName:    'k',
			DefaultValue: cAppName + ".key",
			Help:         "TLS key file",
		},
		TransportProtocolFlag: &CommandLineFlag{
			ShortName:    't',
			DefaultValue: "tcp4",
			Help:         "transport protocol to use",
		},
		LogLevelFlag: &CommandLineFlag{
			ShortName:    'l',
			DefaultValue: log.LevelWarn,
			Help:         "log level: debug, info, warn or error",
		},
	}
}

func (flags CommandLineFlags) SetFlagsWithArguments(arguments []string) {
	app := kingpin.New(cAppName, cAppNameDescription)
	for flagName := range flags {
		flag := flags[flagName]
		flag.FlagValue = app.Flag(flagName, flag.Help).Short(flag.ShortName).Default(flag.DefaultValue).String()
	}

	kingpin.MustParse(app.Parse(arguments))
}
