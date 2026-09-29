package oka

import "github.com/giantswarm/oka/pkg/config"

var (
	configFile  = "oka.yaml"
	versionFlag = false
	alertID     = ""
)

// init initializes command line flags for the application.
func init() {
	Cmd.Flags().StringVar(&configFile, "config", configFile, "Path to configuration file, flag values take precedence over config file values")
	Cmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information and exit")
	Cmd.Flags().StringVar(&alertID, "alert-id", "", "Process a single alert by ID and exit (instead of running continuous mode)")

	config.BindFlags(Cmd)
}
