package config

import (
	"time"
)

// Config represents the application's configuration. It holds settings for
// logging, LLM, OpsGenie, MCP servers, and other operational parameters.
type Config struct {
	LogLevel         string           `mapstructure:"log_level"`         // Log level for the application (e.g., "debug", "info", "error")
	LogFile          string           `mapstructure:"log_file"`          // Path to the log file, if empty logging is disabled
	MaxCalls         int              `mapstructure:"max_calls"`         // Maximum number of calls to the LLM per session
	RunbookDir       string           `mapstructure:"runbook_dir"`       // Directory containing runbooks for the application
	RunbookContainer RunbookContainer `mapstructure:"runbook_container"` // Configuration for the runbook container, including image and port
	SessionsLogDir   string           `mapstructure:"sessions_log_dir"`  // Directory to store session logs
	SlackHandle      string           `mapstructure:"slack_handle"`      // Slack handle to use for notifications

	InitCommands  []Command     `mapstructure:"init_commands"`  // Commands to run during initialization
	LLM           LLM           `mapstructure:"llm"`            // LLM configuration for the application
	MCPServers    MCPServers    `mapstructure:"mcp_servers"`    // MCP servers to configure
	MCPKubernetes MCPKubernetes `mapstructure:"mcp_kubernetes"` // Kubernetes configuration for MCP servers
	OpsGenie      OpsGenie      `mapstructure:"opsgenie"`       // OpsGenie configuration for fetching alerts
}

// OpsGenie holds the configuration for the OpsGenie integration, including API
// settings, alert filtering, and polling interval.
type OpsGenie struct {
	APIUrl      string        `mapstructure:"api_url"`      // API URL is the OpsGenie API endpoint URL, defaults to the official API URL
	EnvVar      string        `mapstructure:"env_var"`      // Environment variable for the OpsGenie API token
	Interval    time.Duration `mapstructure:"interval"`     // Interval for fetching alerts
	QueryString string        `mapstructure:"query_string"` // Query string to filter alerts, e.g., "status:open AND tags:team"
	Team        string        `mapstructure:"team"`         // Team name to filter alerts
}

// MCPServers is a map of MCP server configurations, where the key is the server
// name.
type MCPServers map[string]MCPServer

// MCPServer represents the configuration for an MCP server, including the
// command to run, arguments, environment variables, and other settings.
type MCPServer struct {
	Args                     []string `mapstructure:"args"`                                 // Arguments for the MCP server command
	Command                  string   `mapstructure:"command"`                              // Command to run the MCP server
	Disabled                 bool     `mapstructure:"disabled,omitempty"`                   // Whether this server is disabled
	Env                      []string `mapstructure:"env"`                                  // Environment variables for the MCP server command
	InitializeTimeoutSeconds *int     `mapstructure:"initialize_timeout_seconds,omitempty"` // Timeout for server initialization in seconds
	Shared                   *bool    `mapstructure:"shared,omitempty"`                     // Whether this server is shared across sessions
	URL                      string   `mapstructure:"url"`                                  // URL of the MCP server
}

type MCPKubernetes struct {
	MCPServer      `mapstructure:",squash"` // Embed MCPServer to inherit its fields
	AdditionalArgs []string                 `mapstructure:"additional_args"` // Additional arguments for the Kubernetes MCP server
	KubeConfig     MCPKubernetesKubeConfig  `mapstructure:"kube_config"`     // KubeConfig settings for the Kubernetes MCP server
}

type MCPKubernetesKubeConfig struct {
	FromEnv  string `mapstructure:"from_env"`  // Environment variable to read the kubeconfig from
	FromFlag string `mapstructure:"from_flag"` // Command-line flag to read the kubeconfig from
}

// LLM holds the configuration for the Large Language Model, including the
// provider, model name, and API token.
type LLM struct {
	Model    string `mapstructure:"model"`    // Model name (e.g., "gpt-3.5-turbo", "claude-2")
	Provider string `mapstructure:"provider"` // LLM provider (e.g., "openai", "anthropic")
	Token    string `mapstructure:"token"`    // API token for the LLM provider
}

// Command represents a command to be executed, including its arguments and
// environment variables.
type Command struct {
	Args    []string `mapstructure:"args"`    // Arguments for the command
	Command string   `mapstructure:"command"` // Command to run
	Env     []string `mapstructure:"env"`     // Environment variables for the command
}

type RunbookContainer struct {
	Image string `mapstructure:"image"` // Docker image for the runbook container
	Port  string `mapstructure:"port"`  // Port to expose for the runbook container
}
