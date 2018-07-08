package application

import (
	"github.com/alecthomas/kingpin"
	"github.com/nextmetaphor/tcp-proxy-pool/log"
)

const (
	cAppName            = "tcp-proxy-pool"
	cAppNameDescription = "TODO"

	// HostNameFlag is the command-line flag which allows the listen host to be specified
	HostNameFlag          = "hostname"
	// PortNameFlag is the command-line flag which allows the listen port to be specified
	PortNameFlag          = "port"
	// CertFileFlag is the command-line flag which allows the TLS certificate file is specified 
	CertFileFlag          = "certFile"
	// KeyFileFlag is the command-line flag which allows the TLS key file to be specified
	KeyFileFlag           = "keyFile"
	// TransportProtocolFlag is the command-line flag which allows the TCP transport to be specified
	TransportProtocolFlag = "transport"
	// LogLevelFlag is the command-line flag which allows the log level to be specified
	LogLevelFlag          = "logLevel"
)

type (
	commandLineFlag struct {
		ShortName    rune
		Help         string
		DefaultValue string
		FlagValue    *string
	}

	commandLineFlags map[string]*commandLineFlag
)

// CreateFlags creates the appropriate struct representing the application command line arguments
func CreateFlags() commandLineFlags {
	return commandLineFlags{
		HostNameFlag: &commandLineFlag{
			ShortName:    'h',
			DefaultValue: "",
			Help:         "hostname to bind to",
		},
		PortNameFlag: &commandLineFlag{
			ShortName:    'p',
			DefaultValue: "8443",
			Help:         "port to bind to",
		},
		CertFileFlag: &commandLineFlag{
			ShortName:    'c',
			DefaultValue: cAppName + ".crt",
			Help:         "TLS certificate file",
		},
		KeyFileFlag: &commandLineFlag{
			ShortName:    'k',
			DefaultValue: cAppName + ".key",
			Help:         "TLS key file",
		},
		TransportProtocolFlag: &commandLineFlag{
			ShortName:    't',
			DefaultValue: "tcp4",
			Help:         "transport protocol to use",
		},
		LogLevelFlag: &commandLineFlag{
			ShortName:    'l',
			DefaultValue: log.LevelWarning,
			Help:         "log level: debug, info, warn or error",
		},
	}
}

// SetFlagsWithArguments sets the value of the flags struct with the actual command-line arguments
func (flags commandLineFlags) SetFlagsWithArguments(arguments []string) {
	app := kingpin.New(cAppName, cAppNameDescription)
	for flagName := range flags {
		flag := flags[flagName]
		flag.FlagValue = app.Flag(flagName, flag.Help).Short(flag.ShortName).Default(flag.DefaultValue).String()
	}

	kingpin.MustParse(app.Parse(arguments))
}
