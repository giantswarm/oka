package client

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/llms"
)

// GetTools returns the list of available tools.
func (c *Clients) GetTools() []llms.Tool {
	return c.tools
}

// GetToolClient returns the MCP client for a given tool.
func (c *Clients) GetToolClient(toolName string) *client.Client {
	toolInfo := c.toolsClients[toolName]
	if toolInfo == nil {
		return nil
	}
	return toolInfo.Client
}

// GetToolInfo returns the full ToolInfo for a given tool.
func (c *Clients) GetToolInfo(toolName string) *ToolInfo {
	return c.toolsClients[toolName]
}

// CallTool calls a tool with the given name and arguments.
func (c *Clients) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	toolInfo := c.toolsClients[name]
	if toolInfo == nil {
		return "", fmt.Errorf("no client found for tool %s", name)
	}

	// Create a proper CallToolRequest.
	req := mcp.CallToolRequest{}
	// Set the original tool name (stored in ToolInfo) and arguments in the params field.
	req.Params.Name = toolInfo.OriginalToolName
	req.Params.Arguments = args

	// Call the tool using the official client.
	result, err := toolInfo.Client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to call tool %s: %w", name, err)
	}

	// Check if the tool call resulted in an error.
	if result.IsError {
		// Attempt to extract error message if available.
		var errMsgText string
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				errMsgText = textContent.Text
			} else {
				errMsgText = "Unknown error"
			}
		} else {
			errMsgText = "Unknown error"
		}

		return "", fmt.Errorf(errMsgText)
	}

	// Extract text content from the result.
	var resultText string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			resultText += textContent.Text
		}
	}

	return resultText, nil
}

// convertToolsResultToLLMtools converts a slice of MCP tools to a slice of
// LangChainGo tools.
func convertToolsResultToLLMtools(mcpTools []mcp.Tool, clientName string) []llms.Tool {
	var llmsTools []llms.Tool

	for _, mcpTool := range mcpTools {
		llmTool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        fmt.Sprintf("%s_%s", clientName, mcpTool.Name),
				Description: mcpTool.Description,
			},
		}

		if mcpTool.InputSchema.Type != "" {
			parameters := map[string]any{
				"type": mcpTool.InputSchema.Type,
			}
			if mcpTool.InputSchema.Properties != nil {
				parameters["properties"] = mcpTool.InputSchema.Properties
			}
			if mcpTool.InputSchema.Required != nil {
				parameters["required"] = mcpTool.InputSchema.Required
			}
			llmTool.Function.Parameters = parameters
		}

		llmsTools = append(llmsTools, llmTool)
	}

	return llmsTools
}
