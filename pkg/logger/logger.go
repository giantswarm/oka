// Package logger provides functionality for setting up and managing logging.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
)

// levels is a map of log level names to their corresponding slog.Level values.
var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// GetLevel returns the slog.Level for a given log level name.
func GetLevel(name string) (slog.Level, error) {
	if level, ok := levels[name]; ok {
		return level, nil
	}

	return -1, fmt.Errorf("unknown log level: %s", name)
}

// GetLevels returns a slice of available log level names.
func GetLevels() []string {
	return slices.Collect(maps.Keys(levels))
}

// Setup initializes the global logger with the specified log level and output
// file. If no log file is provided, it defaults to stderr. It returns a closer
// function to be called on application shutdown.
func Setup(logLevel, logFile string) (closer func(), err error) {
	var logWriter io.Writer
	if logFile != "" {
		// If a log file is specified, create/open it and use it for logging.
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644) // nolint:gosec
		if err != nil {
			return nil, err
		}
		closer = func() { file.Close() } // nolint:errcheck
		logWriter = file
	} else {
		// Set up logging. Default to stderr if no log file is specified.
		logWriter = os.Stderr
		closer = func() {}
	}

	level, err := GetLevel(logLevel)
	if err != nil {
		return nil, err
	}

	// Set up the logger with a custom format and level.
	logger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
		Level: level,
	}))

	// Set the global logger.
	slog.SetDefault(logger)

	return closer, nil
}
