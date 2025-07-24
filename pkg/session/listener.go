package session

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/tmc/langchaingo/llms"

	"github.com/giantswarm/oka/pkg/config"
	"github.com/giantswarm/oka/pkg/mcp/client"
)

const endSessionPhrase = "investigation complete"

//go:embed system-prompt.tmpl
var systemPromptTmpl string
var systemPromptTemplate *template.Template
var systemPrompt string

func init() {
	systemPromptTemplate = template.Must(template.New("system-prompt").Funcs(sprig.FuncMap()).Parse(systemPromptTmpl))
}

// Listen listens for incoming alerts and starts a new session for each one.
func Listen(ctx context.Context, c <-chan any, llmModel llms.Model, mcpClients *client.Clients, conf *config.Config) error {
	// Add system prompt instructions
	systemPromptData := struct {
		SlackHandle      string
		EndSessionPhrase string
	}{
		SlackHandle:      conf.SlackHandle,
		EndSessionPhrase: endSessionPhrase,
	}

	var systemPromptBuilder strings.Builder
	err := systemPromptTemplate.Execute(&systemPromptBuilder, systemPromptData)
	if err != nil {
		return fmt.Errorf("failed to execute system prompt template: %w", err)
	}
	systemPrompt = systemPromptBuilder.String()

	slog.Info("Session service started")

	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case alert := <-c:
				wg.Add(1)
				go func(alert any, llmModel llms.Model, mcpClients *client.Clients, conf *config.Config) {
					defer wg.Done()
					run(ctx, alert, llmModel, mcpClients, conf)
				}(alert, llmModel, mcpClients, conf)
			}
		}
	}()

	<-ctx.Done()
	slog.Info("Waiting for sessions to complete")
	wg.Wait()
	slog.Info("Session service stopped")

	return nil
}

// run starts a new session for the given alert.
func run(ctx context.Context, alert any, llmModel llms.Model, mcpClients *client.Clients, conf *config.Config) {
	sessionClients := mcpClients.Clone()
	// TODO: close non-shared clients
	err := sessionClients.RegisterServersConfig(ctx, conf.GetMCPServers(false))
	if err != nil {
		slog.Error("Failed to register MCP servers", "error", err)
		return
	}

	s, err := New(alert, llmModel, sessionClients, conf.MaxCalls, conf.SessionsLogDir)
	if err != nil {
		slog.Error("Failed to create new session", "error", err)
		return
	}

	s.Run(ctx)
}

// ProcessSingleAlert processes a single alert directly without continuous listening.
// This is used for the single alert processing mode.
func ProcessSingleAlert(ctx context.Context, alert any, llmModel llms.Model, mcpClients *client.Clients, conf *config.Config) error {
	// Initialize system prompt with configuration data
	systemPromptData := struct {
		SlackHandle      string
		EndSessionPhrase string
	}{
		SlackHandle:      conf.SlackHandle,
		EndSessionPhrase: endSessionPhrase,
	}

	var systemPromptBuilder strings.Builder
	err := systemPromptTemplate.Execute(&systemPromptBuilder, systemPromptData)
	if err != nil {
		return fmt.Errorf("failed to execute system prompt template: %w", err)
	}
	systemPrompt = systemPromptBuilder.String()

	// Clone and configure MCP clients for this session
	sessionClients := mcpClients.Clone()
	err = sessionClients.RegisterServersConfig(ctx, conf.GetMCPServers(false))
	if err != nil {
		return fmt.Errorf("failed to register MCP servers: %w", err)
	}

	// Create and run the session
	s, err := New(alert, llmModel, sessionClients, conf.MaxCalls, conf.SessionsLogDir)
	if err != nil {
		return fmt.Errorf("failed to create new session: %w", err)
	}

	s.Run(ctx)
	return nil
}
