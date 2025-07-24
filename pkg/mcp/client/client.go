// Package client provides functionality for interacting with MCP (Model-Context
// Protocol) servers.
package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tmc/langchaingo/llms"

	"github.com/giantswarm/oka/pkg/config"
	"github.com/giantswarm/oka/pkg/kubernetes"
)

// ToolInfo contains metadata about an MCP tool.
type ToolInfo struct {
	Client           *client.Client
	ClientName       string
	OriginalToolName string
}

var (
	// defaultInitTimeout is the default timeout for initializing an MCP client.
	defaultInitTimeout = 15 * time.Second
)

// Clients manages a collection of MCP clients and their associated tools.
type Clients struct {
	tools         []llms.Tool
	toolsClients  map[string]*ToolInfo
	uniqueClients []*client.Client
}

// New creates a new Clients instance.
func New() *Clients {
	c := &Clients{
		tools:         make([]llms.Tool, 0),
		toolsClients:  make(map[string]*ToolInfo),
		uniqueClients: make([]*client.Client, 0),
	}

	return c
}

// Clone creates a new Clients instance with the same tools and clients.
func (c Clients) Clone() *Clients {
	newClients := &Clients{
		tools:        make([]llms.Tool, len(c.tools)),
		toolsClients: make(map[string]*ToolInfo, len(c.toolsClients)),
	}

	newClients.toolsClients = maps.Clone(c.toolsClients)
	newClients.tools = slices.Clone(c.tools)

	return newClients
}

// RegisterServersConfig registers MCP servers from the provided configuration.
func (c *Clients) RegisterServersConfig(ctx context.Context, mcpServers config.MCPServers) error {
	serverCount := 0
	for name, server := range mcpServers {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Continue if context is not done
		}

		if server.Disabled {
			slog.Info("Skipping disabled MCP server", "server", name)
			continue
		}

		// Create a new MCP client.
		sc, err := newClient(server)
		if err != nil {
			return err
		}

		err = c.RegisterClient(ctx, sc, name, server.InitializeTimeoutSeconds)
		if err != nil {
			// If the client failed to initialize, close it and continue.
			return err
		}

		serverCount++
	}

	slog.Info("Finished initializing all MCP clients", "count", serverCount)

	return nil
}

// RegisterServer registers a new MCP server.
func (c *Clients) RegisterServer(ctx context.Context, mcpServer *server.MCPServer, name string) error {
	// Create a new MCP client.
	sc, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		return err
	}

	err = c.RegisterClient(ctx, sc, name, nil)
	if err != nil {
		// If the client failed to initialize, close it and continue.
		return err
	}

	slog.Info("Finished initializing all MCP clients")

	return nil
}

// RegisterClient registers a new MCP client.
func (c *Clients) RegisterClient(ctx context.Context, sc *client.Client, name string, initializeTimeoutSeconds *int) error {
	slog.Info("Initializing MCP client", "server", name)

	// Start the client.
	err := sc.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start client for %s: %w", name, err)
	}

	timeout := defaultInitTimeout
	if initializeTimeoutSeconds != nil {
		timeout = time.Duration(*initializeTimeoutSeconds) * time.Second
	}

	// Create a context with timeout for initialization.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Initialize the client.
	_, err = sc.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		t := sc.GetTransport()
		stdioTransport, ok := t.(*transport.Stdio)
		if ok {
			stdioErr, readErr := io.ReadAll(stdioTransport.Stderr())
			if readErr != nil {
				slog.Warn("Failed to read stderr from MCP client", "server", name, "error", readErr)
			} else if len(stdioErr) > 0 {
				return fmt.Errorf("failed to initialize client for %s: Stderr: %s", name, string(stdioErr))
			}
		}
		return fmt.Errorf("failed to initialize client for %s: %w", name, err)
	}

	// List tools from the MCP server.
	toolsResult, err := sc.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools for %s: %w", name, err)
	}

	if len(toolsResult.Tools) == 0 {
		slog.Warn("No tools found for MCP client", "server", name)
		err = sc.Close()
		if err != nil {
			return fmt.Errorf("failed to close client for %s: %w", name, err)
		}
		// TODO: define and handle no tool error
		return nil
	}

	c.uniqueClients = append(c.uniqueClients, sc)

	// Register tools' client with prefixed names.
	llmTools := convertToolsResultToLLMtools(toolsResult.Tools, name)
	toolsCount := 0
	for i, tool := range llmTools {
		_, exists := c.toolsClients[tool.Function.Name]
		if exists {
			slog.Warn("Tool already exists, skipping", "server", name, "tool", tool.Function.Name)
			continue
		}

		// Store with the prefixed name (mcp_clientname_toolname)
		c.toolsClients[tool.Function.Name] = &ToolInfo{
			Client:           sc,
			ClientName:       name,
			OriginalToolName: toolsResult.Tools[i].Name, // Store the original MCP tool name
		}
		c.tools = append(c.tools, tool)
		toolsCount++
	}

	slog.Info("Initialized MCP client", "server", name, "tools", toolsCount)

	return nil
}

// newClient creates a new MCP client from the provided configuration.
func newClient(mcpServer config.MCPServer) (c *client.Client, err error) {
	var t transport.Interface

	switch {
	case mcpServer.URL != "":
		t, err = transport.NewStreamableHTTP(mcpServer.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create transport: %w", err)
		}
	case mcpServer.Command != "":
		fallthrough
	default:
		mcpEnv := mcpServer.Env
		// Create temporary kubeconfig file if the command is for Kubernetes.
		// This is a hack to isolate the kubeconfig file and avoid changing the
		// current user's context.
		if strings.Contains(mcpServer.Command, "kubernetes") {
			// Create a temporary kubeconfig file.
			kubeConfigFile, err := kubernetes.CreateTmpKubeConfigFile()
			if err != nil {
				return nil, err
			}
			// Add the kubeconfig file to the environment variables.
			if mcpEnv == nil {
				mcpEnv = make([]string, 0)
			}

			mcpEnv = append(mcpEnv, fmt.Sprintf("KUBECONFIG=%s", kubeConfigFile))

			slog.Info("Using temporary kubeconfig file", "file", kubeConfigFile)
			// TODO: handle cleanup of the temporary file properly
			//defer os.Remove(kubeConfigFile) // Clean up the temporary file after use
		}
		t = transport.NewStdio(mcpServer.Command, mcpEnv, mcpServer.Args...)
	}

	c = client.NewClient(t)

	return c, nil
}

// Close closes all unique MCP clients.
func (c *Clients) Close() error {
	var errs []error

	for name, client := range c.uniqueClients {
		err := client.Close()
		if err != nil {
			err := fmt.Errorf("failed to close client for %s: %w", name, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
