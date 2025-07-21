# OKA: Oncall Kubernetes Assistant

OKA (Oncall Kubernetes Assistant) is a powerful tool designed to assist on-call engineers in managing and troubleshooting Kubernetes-related alerts. By leveraging the capabilities of Large Language Models (LLMs), OKA can analyze alerts, retrieve relevant runbooks, and provide context-aware assistance to streamline the incident response process.

## Features

- **OpsGenie Integration**: OKA seamlessly integrates with OpsGenie to fetch and process alerts from your designated team.
- **LLM-Powered Assistance**: By using LLMs, OKA can understand the context of an alert and provide intelligent suggestions and insights.
- **Runbook Retrieval**: OKA can automatically retrieve and display relevant runbooks for a given alert, ensuring that you have the necessary information at your fingertips.
- **MCP Tooling**: OKA's functionality can be extended through the use of MCP (Model-Context Protocol) servers, allowing you to integrate with other tools and services (e.g. Kubernetes, Slack, etc.).

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
  opsgenie:
    command: mcp-server-kubernetes
```

See the [reference configuration file](./pkg/config) for more details on available options.

### Running OKA

You can run OKA using the following command:

```bash
go install -v github.com/giantswarm/oka@latest
curl -Lo oka.yaml https://raw.githubusercontent.com/giantswarm/oka/refs/heads/main/pkg/config/oka.yaml.example
# Edit oka.yaml to suit your environment
oka
```

## How It Works

OKA operates by periodically fetching alerts from OpsGenie. When a new, unacknowledged alert is found, OKA initiates a new session to process it. During the session, OKA uses an LLM to analyze the alert and determine the best course of action. This may involve retrieving a runbook, executing a command, or interacting with other tools via MCP servers.

Each session is logged to a file in the `sessions` directory, allowing you to review the entire interaction between OKA and the LLM.
