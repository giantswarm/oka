package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"

	"github.com/giantswarm/oka/pkg/config"
	"github.com/giantswarm/oka/pkg/llm"
	"github.com/giantswarm/oka/pkg/logger"
	"github.com/giantswarm/oka/pkg/mcp/client"
	"github.com/giantswarm/oka/pkg/opsgenie"
	"github.com/giantswarm/oka/pkg/service"
	"github.com/giantswarm/oka/pkg/session"
)

var (
	// name is the name of the application.
	name = "oka"
)

// cmd defines the root command for the MCP time server.
var cmd = &cobra.Command{
	Use:   name,
	Short: "Oncall Kubernetes Assistant (OKA)",
	Long:  `Oncall Kubernetes Assistant (OKA) is a tool to help you manage your oncall duties with OpsGenie by leveraging LLM capabilities.`,
	RunE:  runner,
}

// runner is the main execution function for the MCP server.
// It sets up logging, creates the MCP server, and starts it.
func runner(c *cobra.Command, args []string) (err error) {
	// If the version flag is set, print the version and exit.
	if versionFlag {
		fmt.Println(version.Print(name))
		return nil
	}

	// Load the configuration file.
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		return err
	}
	slog.Info("Loaded config", "file", configFile)
	if conf.LogLevel == "debug" {
		conf.Print()
	}

	// Set up logging.
	logCloser, err := logger.Setup(conf.LogLevel, conf.LogFile)
	if err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}
	defer logCloser()

	// Load environment variables from .env file if it exists.
	err = godotenv.Load(".env")
	if err == nil {
		slog.Info("Loaded environment variables", "file", ".env")
	}

	// Create the sessions log directory if it does not exist.
	err = os.MkdirAll(conf.SessionsLogDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create sessions log directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		signal := <-sigChan
		slog.Info(fmt.Sprintf("Received %s signal", signal))
		cancel()
	}()

	// Initialize MCP servers.
	mcpClients := client.New()
	err = mcpClients.RegisterServersConfig(ctx, conf.GetMCPServers(true))
	if err != nil {
		return err
	}
	defer mcpClients.Close()

	//runbookServer := runbook.NewServer(name, version.Version, conf)
	//service.Run(func() { runbookServer.Start(ctx) })
	//err = mcpClients.RegisterServer(ctx, runbookServer.MCPServer, "runbook")
	//if err != nil {
	//	return fmt.Errorf("failed to register runbook server: %w", err)
	//}

	// Initialize the LLM model.
	llmModel, err := llm.New(conf)
	if err != nil {
		return err
	}
	slog.Info("LLM model initialized", "provider", conf.LLM.Provider)

	// Run initialization commands.
	for _, initCommand := range conf.InitCommands {
		c := exec.Command(initCommand.Command, initCommand.Args...)
		slog.Info("Running init command", "command", c.String())
		_, err = c.Output()
		if err != nil {
			return fmt.Errorf("failed to run init command %s: %w", c.String(), err)
		}
	}

	// Initialize the OpsGenie service.
	opsgenieService, err := opsgenie.NewService(conf)
	if err != nil {
		return fmt.Errorf("failed to create OpsGenie service: %w", err)
	}

	// Start the OpsGenie service and session services.
	alertsChan := make(chan any, 1)
	service.Run(func() { opsgenieService.Start(ctx, alertsChan) })
	service.Run(func() { session.Listen(ctx, alertsChan, llmModel, mcpClients, conf) })

	service.Wait()

	return nil
}
