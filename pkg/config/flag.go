package config

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/oka/pkg/logger"
)

func BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&config.LogFile, "log-file", config.LogFile, "Path to log file (logs is disabled if not specified)")
	cmd.Flags().StringVar(&config.LogLevel, "log-level", config.LogLevel, "Log level to use. Available levels: "+strings.Join(logger.GetLevels(), ", "))
	cmd.Flags().StringVar(&config.SessionsLogDir, "sessions-log-dir", config.SessionsLogDir, "Directory to store session logs")
}
