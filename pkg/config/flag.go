package config

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/giantswarm/oka/pkg/logger"
)

func BindFlags(cmd *cobra.Command) {
	configFlags := pflag.NewFlagSet("config", pflag.ContinueOnError)

	configFlags.String("log-file", defaultConfig().LogFile, "Path to log file (logs is disabled if not specified)")
	configFlags.String("log-level", defaultConfig().LogLevel, "Log level to use. Available levels: "+strings.Join(logger.GetLevels(), ", "))
	configFlags.String("sessions-log-dir", defaultConfig().SessionsLogDir, "Directory to store session logs")

	configFlags.VisitAll(func(flag *pflag.Flag) {
		key := strings.ReplaceAll(flag.Name, "-", "_")
		if err := viper.BindPFlag(key, flag); err != nil {
			panic("failed to bind flag: " + flag.Name + ": " + err.Error())
		}
	})

	cmd.Flags().AddFlagSet(configFlags)
}
