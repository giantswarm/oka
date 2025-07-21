package config

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/spf13/viper"
)

var (
	config = Config{
		LogLevel:       "info",
		MaxCalls:       20,
		SessionsLogDir: "sessions",

		InitCommands: []Command{
			{
				Command: "tsh",
				Args:    []string{"kube", "login", "--all"},
			},
		},
		MCPServers: make(map[string]MCPServer),
		OpsGenie: &OpsGenie{
			APIUrl:      string(client.API_URL),
			EnvVar:      "OPSGENIE_TOKEN",
			Interval:    30 * time.Second,
			QueryString: `responders: "{{ .Team }}" AND status: open`,
		},
	}
)

func LoadConfig(cfgFile string) (*Config, error) {
	if cfgFile == "" {
		return &config, nil
	}

	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = viper.UnmarshalExact(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (conf *Config) Print() {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)

	fmt.Fprintf(w, "log_level:\t%s\n", conf.LogLevel)
	fmt.Fprintf(w, "log_file:\t%s\n", conf.LogFile)
	fmt.Fprintf(w, "max_calls:\t%d\n", conf.MaxCalls)
	fmt.Fprintf(w, "runbook_dir:\t%s\n", conf.RunbookDir)
	fmt.Fprintf(w, "slack_handle:\t%s\n", conf.SlackHandle)
	fmt.Fprintf(w, "sessions_log_directory:\t%s\n", conf.SessionsLogDir)
	fmt.Fprintf(w, "init_commands:\t%d\n", len(conf.InitCommands))
	for _, initCmd := range conf.InitCommands {
		fmt.Fprintf(w, "\t- %s %s\n", initCmd.Command, strings.Join(initCmd.Args, " "))
	}
	fmt.Fprintf(w, "opsgenie.api_url:\t%s\n", conf.OpsGenie.APIUrl)
	fmt.Fprintf(w, "opsgenie.query_string:\t%s\n", conf.OpsGenie.QueryString)
	fmt.Fprintf(w, "opsgenie.environment_variable:\t%s\n", conf.OpsGenie.EnvVar)
	fmt.Fprintf(w, "opsgenie.interval:\t%s\n", conf.OpsGenie.Interval)
	fmt.Fprintf(w, "opsgenie.team:\t%s\n", conf.OpsGenie.Team)
	fmt.Fprintf(w, "llm.model:\t%s\n", conf.LLM.Model)
	fmt.Fprintf(w, "llm.provider:\t%s\n", conf.LLM.Provider)
	fmt.Fprintf(w, "mcp_servers:\t%d\n", len(conf.MCPServers))
	for name, server := range conf.MCPServers {
		if server.Command != "" {
			fmt.Fprintf(w, "\t- %s: %s %s\n", name, server.Command, strings.Join(server.Args, " "))
		} else {
			fmt.Fprintf(w, "\t- %s: %s\n", name, server.URL)
		}
	}

	w.Flush()
}
