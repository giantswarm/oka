package main

import (
	"log/slog"
	"os"

	"github.com/giantswarm/oka/cmd/oka"
)

// main is the entry point of the application.
// It sets up the command-line interface and executes the root command.
// If any errors occur during execution, it logs the error and exits with a
// non-zero status code.
func main() {
	// Execute the root command and handle any errors.
	err := oka.Cmd.Execute()
	if err != nil {
		slog.Error("execution error", "error", err)
		os.Exit(1)
	}
}
