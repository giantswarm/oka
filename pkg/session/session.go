package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"

	"github.com/giantswarm/oka/pkg/mcp/client"
)

const (
	motivationText = "Provide next steps to continue the investigation."
)

// Session represents an AI assistant session for processing a single alert.
type Session struct {
	ID string

	alert      any
	llm        llms.Model
	logFile    *os.File
	maxCalls   int
	mcpClients *client.Clients
	messages   []llms.MessageContent
}

// New creates a new session for processing an alert.
func New(alert any, llm llms.Model, mcpClients *client.Clients, maxToolCalls int, logDir string) (*Session, error) {
	id := uuid.New().String()

	logFile := fmt.Sprintf("%s/session-%s.log", logDir, id)
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open session log file: %w", err)
	}

	s := &Session{
		ID:         id,
		alert:      alert,
		llm:        llm,
		logFile:    f,
		maxCalls:   maxToolCalls,
		mcpClients: mcpClients,
		messages:   make([]llms.MessageContent, 0),
	}

	return s, nil
}

// Run starts the session and processes the alert.
func (s *Session) Run(ctx context.Context) {
	var finalErr error

	slog.Info("Starting session", "session.id", s.ID, "logFile", s.logFile.Name())
	defer slog.Info("Stopping session", "session.id", s.ID)
	defer s.logFile.Close()
	defer func() {
		if finalErr != nil {
			s.log("\n## Error\n%s\n", finalErr.Error())
		}
		s.log("\n# Session end")
	}()

	// Add the alert to the session context.
	alertBytes, err := json.Marshal(s.alert)
	if err != nil {
		slog.Error("Failed to marshal alert to json", "error", err, "session.id", s.ID)
		finalErr = fmt.Errorf("failed to marshal alert: %w", err)
		return
	}
	s.addToContext(llms.ChatMessageTypeHuman, llms.TextPart(string(alertBytes)))

	// Add system prompt instructions.
	s.addToContext(llms.ChatMessageTypeSystem, llms.TextPart(systemPrompt))

	s.log("# Session initialized: %s\n", s.ID)
	s.log("\n## Alert\n%s\n", string(alertBytes))
	s.log("\n## Prompt\n%s\n", systemPrompt)
	s.log("\n## Tools\n")
	for _, tool := range s.mcpClients.GetTools() {
		s.log("- %s: %s\n", tool.Function.Name, tool.Function.Description)
	}

	s.log("\n# Session start LLM\n")
	for i := 0; i < s.maxCalls; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Continue if context is not done
		}

		lastCall := i == (s.maxCalls - 1)
		if lastCall {
			s.addToContext(llms.ChatMessageTypeSystem, llms.TextPart("You must now complete your investigation and provide a final response."))
		}

		slog.Info("Calling LLM", "session.id", s.ID)
		llmResponse, toolCalls, err := s.callLLM(ctx, lastCall)
		if err != nil {
			slog.Error("Failed to call LLM", "error", err, "session.id", s.ID)
			finalErr = fmt.Errorf("failed to call LLM: %w", err)
			return
		}

		if llmResponse.ReasoningContent != "" {
			s.log("\n## LLM reasoning\n%s\n", llmResponse.ReasoningContent)
		}

		// Check if the investigation is complete (considers both content and tool calls)
		if s.isInvestigationComplete(llmResponse, lastCall) {
			return
		}

		// Trim whitespace from response content to avoid API rejection
		trimmedContent := strings.TrimSpace(llmResponse.Content)
		if trimmedContent != "" {
			s.addToContext(llms.ChatMessageTypeAI, llms.TextPart(trimmedContent))
		}
		s.log("\n## LLM response\n%s\n", trimmedContent)

		// Insist in providing next steps if no tool calls are suggested.
		if len(toolCalls) == 0 {
			s.addToContext(llms.ChatMessageTypeHuman, llms.TextPart(motivationText))
			s.log("\n## Insist LLM to provide next steps\n%s\n", motivationText)
		}

		// Create a context with timeout for tool processing.
		toolCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()

		for _, toolCall := range toolCalls {
			s.addToContext(llms.ChatMessageTypeAI, toolCall)

			slog.Info("Tool call", "session.id", s.ID, "tool", toolCall.FunctionCall.Name)
			s.log("\n## Tool call\ntool: %s\nargs: %s\n", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

			args := make(map[string]interface{})
			err = json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
			if err != nil {
				slog.Error("Failed to unmarshal tool call arguments", "error", err, "session.id", s.ID, "arguments", toolCall.FunctionCall.Arguments)
				finalErr = fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
				return
			}

			toolResponse, err := s.mcpClients.CallTool(toolCtx, toolCall.FunctionCall.Name, args)
			if err != nil {
				slog.Error("Failed to process tool response", "error", err, "session.id", s.ID, "toolCall", toolCall.FunctionCall.Name)
				toolResponse = fmt.Sprintf("Error: %s", err.Error())
			}

			slog.Info("Tool response", "session.id", s.ID, "tool", toolCall.FunctionCall.Name, "response", len(toolResponse))
			s.log("\n## Tool response\ntool: %s\n%s\n", toolCall.FunctionCall.Name, toolResponse)

			// Add history.
			toolResponsePart := llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.FunctionCall.Name,
				Content:    strings.TrimSpace(toolResponse),
			}
			s.addToContext(llms.ChatMessageTypeTool, toolResponsePart)
		}
	}
}

// addToContext adds a message to the session's context.
func (s *Session) addToContext(role llms.ChatMessageType, parts ...llms.ContentPart) {
	message := llms.MessageContent{
		Role:  role,
		Parts: parts,
	}
	s.messages = append(s.messages, message)
}

// callLLM generates a text completion using the specified provider from the registry.
func (s Session) callLLM(ctx context.Context, lastCall bool) (*llms.ContentChoice, []llms.ToolCall, error) {
	// Create a context with appropriate timeout.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	options := []llms.CallOption{
		llms.WithTools(s.mcpClients.GetTools()),
	}

	resp, err := s.llm.GenerateContent(ctx, s.messages, options...)
	if err != nil {
		return nil, nil, err
	}

	// Merge ToolCalls from all choices and return them
	var allToolCalls []llms.ToolCall
	for _, choice := range resp.Choices {
		allToolCalls = append(allToolCalls, choice.ToolCalls...)
	}
	return resp.Choices[0], allToolCalls, nil
}

// log writes a message to the session's log file.
func (s Session) log(format string, args ...any) {
	fmt.Fprintf(s.logFile, format, args...)
}

// isInvestigationComplete checks if the LLM response indicates the investigation is finished
func (s *Session) isInvestigationComplete(response *llms.ContentChoice, isLastCall bool) bool {
	// Check for completion phrases
	if strings.Contains(strings.ToLower(response.Content), endSessionPhrase) {
		slog.Info("LLM indicated investigation is complete", "session.id", s.ID, "phrase", endSessionPhrase)
		return true
	}

	// If we haven't reached the limit of calls, continue the investigation
	if !isLastCall {
		return false
	}

	// If there are still tool call suggestions after the last call, log it
	if len(response.ToolCalls) > 0 {
		slog.Info("LLM call limit reached, but there are still tool call suggestions. Ending session prematurely.", "session.id", s.ID)
	}

	return true
}
