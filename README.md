# OKA: Oncall Kubernetes Assistant

OKA (Oncall Kubernetes Assistant) is a powerful tool designed to assist on-call engineers in managing and troubleshooting Kubernetes-related alerts. By leveraging the capabilities of Large Language Models (LLMs), OKA can analyze alerts, retrieve relevant runbooks, and provide context-aware assistance to streamline the incident response process.

## Features

- **OpsGenie Integration**: OKA seamlessly integrates with OpsGenie to fetch and process alerts from your designated team.
- **LLM-Powered Assistance**: By using LLMs, OKA can understand the context of an alert and provide intelligent suggestions and insights.
- **Runbook Retrieval**: OKA can automatically retrieve and display relevant runbooks for a given alert, ensuring that you have the necessary information at your fingertips.
- **MCP Tooling**: OKA's functionality can be extended through the use of MCP (Model-Context Protocol) servers, allowing you to integrate with other tools and services (e.g. Kubernetes, Slack, etc.).
- **Intelligent Tool Discovery**: Uses [Muster MCP aggregator](https://github.com/giantswarm/muster) for smart tool discovery and dynamic tool loading, reducing AI token costs and improving context efficiency.

## Getting Started

To get started with OKA, you will need to create a configuration file (e.g., `oka.yaml`) and provide it to the application. The configuration file allows you to specify your OpsGenie API key, LLM provider, and other settings.

### Configuration

Here is an example of a basic configuration file (both YAML and JSON formats are supported):

```yaml
log_level: debug
runbook_dir: /path/to/your/runbooks
slack_handle: your-slack-handle
opsgenie:
  team: your-team-name
  interval: 1m
  query_string: 'status: open'
llm:
  provider: openai
  token: your-openai-api-key
  model: gpt-4
mcp_servers:
  muster:
    url: http://localhost:8099
```

See the [reference configuration file](./pkg/config) for more details on available options.

### Prerequisites

OKA uses [Muster](https://github.com/giantswarm/muster) as an MCP aggregator for intelligent tool discovery. You'll need to set it up first:

1. **Install and start Muster**:
```bash
# Install Muster
go install github.com/giantswarm/muster@latest

# Start Muster aggregator on port 8099
muster serve --port 8099
```

2. **Configure Muster with your tools** (create `.muster/config.yaml`):
```yaml
servers:
  kubernetes:
    command: mcp-server-kubernetes
  prometheus:
    command: mcp-server-prometheus
  # Add other MCP servers as needed
```

### Running OKA

You can run OKA using the following command:

```bash
go install -v github.com/giantswarm/oka@latest
curl -Lo oka.yaml https://raw.githubusercontent.com/giantswarm/oka/refs/heads/main/pkg/config/oka.yaml.example
# Edit oka.yaml to suit your environment
oka
```

## How It Works

OKA operates by periodically fetching alerts from OpsGenie. When a new, unacknowledged alert is found, OKA initiates a new session to process it. During the session, OKA uses an LLM to analyze the alert and determine the best course of action.

**Intelligent Tool Discovery with Muster:**
1. **Dynamic Tool Loading**: Instead of loading all tools upfront, OKA uses Muster's meta-tools (`list_tools`, `filter_tools`, `call_tool`) for intelligent discovery
2. **Context Optimization**: Tools are discovered and loaded only when needed, reducing AI token costs
3. **Flexible Workflows**: The AI can discover Kubernetes, monitoring, and other tools dynamically based on the alert context
4. **Efficient Investigation**: Use `filter_tools(pattern="kubernetes")` to find relevant tools, then `call_tool` to execute them

**Session Flow:**
- **Alert Analysis**: LLM examines the alert data
- **Tool Discovery**: Uses `filter_tools` to find relevant investigation tools
- **Investigation**: Executes tools via `call_tool` to gather information
- **Resolution**: Takes action if confident, or escalates with detailed findings
- **Reporting**: Provides structured summary via Slack

Each session is logged to a file in the `sessions` directory, allowing you to review the entire interaction between OKA and the LLM.
