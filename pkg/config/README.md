# Config reference

Here is the reference for the OKA configuration file. You can use this as a guide to set up your `oka.yaml` file.

```yaml
# OKA Configuration File
# Allowed levels: debug, info, warn, error
log_level: info
# File used to log OKA's output, stderr is used if not specified
log_file: ""
# Maximum number of iterations for LLM calls
max_calls: 20
# Directory used to store session logs
session_log_dir: "sessions"
# Slack handle for notifications, find it in your Slack profile > Copy member ID
slack_handle: ""
# Commands to run at startup
init_commands:
  - command: tsh
    args:
    - kube
    - login
    - --all
    env:
    - KEY=value
# LLM configuration
llm:
  # LLM model to use, e.g., "gpt-4", "gpt-3.5-turbo"
  model: ""
  # LLM provider, supported values: "openai", "anthropic", "google".
  provider: ""
  # LLM token for authentication
  token: ""
# List of MCP servers providing additional functionality to the LLM
# Recommended: Use Muster as an MCP aggregator for intelligent tool discovery
mcp_servers:
  # Muster MCP aggregator - provides meta-tools for intelligent tool discovery
  muster:
    # URL for the Muster MCP aggregator running locally
    url: "http://localhost:8099"
    # Muster should be shared across sessions for efficiency
    shared: true
  # Alternative: Direct MCP server configuration (legacy)
  # example_server:
    # Command to run the MCP server
    # command: ""
    # Arguments for the MCP server command
    # args: []
    # URL for the MCP server, mutually exclusive with command, takes precedence if both are provided
    # url: ""
    # Is the MCP server enabled?
    # disabled: false
    # Environment variables provided to the MCP server
    # env:
    #   KEY: value
    # Timeout for the MCP server to initialize
    # initialize_timeout_seconds: 15s
    # Optional: If true, a new MCP server will be started for each session
    # shared: false
# OpsGenie configuration
opsgenie:
  # API URL for OpsGenie, default is the v2 API
  api_url: "https://api.opsgenie.com/v2"
  # Environment variable containing the OpsGenie API key
  envVar: "OPSGENIE_API_KEY"
  # Query string to filter alerts, {{ .Team }} and {{ .Today }} placeholders are available
  query_string: 'responder: "{{ .Team }}" AND status: open'
  # Team name to use for the {{ .Team }} placeholder
  team: ""
  # Interval for fetching alerts, e.g., "1m", "30s"
  interval: 30s
```
