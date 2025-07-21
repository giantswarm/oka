package main

import "github.com/giantswarm/oka/pkg/config"

var (
	configFile  = "oka.yaml"
	versionFlag = false
)

// init initializes command line flags for the application.
func init() {
	cmd.Flags().StringVar(&configFile, "config", configFile, "Path to configuration file, config values takes precedence over flags")
	cmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information and exit")

	config.BindFlags(cmd)
}
