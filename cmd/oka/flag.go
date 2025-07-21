package oka

import "github.com/giantswarm/oka/pkg/config"

var (
	configFile  = "oka.yaml"
	versionFlag = false
)

// init initializes command line flags for the application.
func init() {
	Cmd.Flags().StringVar(&configFile, "config", configFile, "Path to configuration file, config values takes precedence over flags")
	Cmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information and exit")

	config.BindFlags(Cmd)
}
