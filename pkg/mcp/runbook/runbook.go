// Package runbook provides an MCP server for retrieving runbooks.
package runbook

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/giantswarm/oka/pkg/config"
)

// Server wraps the core MCP server and provides runbook-specific functionality.
type Server struct {
	*server.MCPServer
}

// NewServer creates a new MCP server with the runbook tool registered.
// It initializes the underlying MCP server and registers the `get_runbook` tool.
func NewServer(name, version string, conf *config.Config) *Server {
	mcpServer := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer,
	}

	registerHandlers(s)

	return s
}

// Start starts the server, listening on standard input/output.
// It blocks until the context is canceled or an error occurs.
func (s Server) Start(ctx context.Context) error {
	stdioServer := server.NewStdioServer(s.MCPServer)

	// Set the logger for the stdio server.
	stdioServer.SetErrorLogger(log.Default())

	err := stdioServer.Listen(ctx, os.Stdin, os.Stdout)
	// Don't return an error if the context was canceled, as it's an expected shutdown.
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

// registerHandlers registers the tool handlers for the runbook server.
func registerHandlers(s *Server) {
	getRunbook := mcp.NewTool("get_runbook",
		mcp.WithDescription("Get the runbook for a specific alert"),
		mcp.WithString("url",
			mcp.Description("URL of the runbook"),
			mcp.Required(),
		),
	)
	s.AddTool(getRunbook, s.GetRunbook)

}

// GetRunbook is the tool implementation for retrieving a runbook.
// It takes a URL as input and returns the content of the corresponding runbook
// file.
func (s *Server) GetRunbook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	requestURL := request.GetString("url", "")
	if requestURL == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
	}

	_, err := url.Parse(requestURL)
	if err != nil {
		return mcp.NewToolResultError("invalid URL format: " + err.Error()), nil
	}

	// TODO: implement the logic to retrieve the runbook content from the URL.
	content := ""

	return mcp.NewToolResultText(string(content)), nil
}
