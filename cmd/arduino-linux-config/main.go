package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arduino/arduino-linux-config/cmd/arduino-linux-config/carrier"
	"github.com/arduino/arduino-linux-config/cmd/feedback"

	"github.com/spf13/cobra"
	"go.bug.st/cleanup"
)

// Version will be set a build time with -ldflags
var Version string = "0.0.0-dev"
var format string
var logLevelStr string

func run() error {
	rootCmd := &cobra.Command{
		Use:   "arduino-linux-config",
		Short: "A CLI to manage Arduino Linux Configurations",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			format, ok := feedback.ParseOutputFormat(format)
			if !ok {
				feedback.Fatal(fmt.Sprintf("Invalid output format: %s", format), feedback.ErrBadArgument)
			}
			feedback.SetFormat(format)

			logLevel, err := ParseLogLevel(logLevelStr)
			if err != nil {
				feedback.FatalError(err, feedback.ErrBadArgument)
			}
			slog.SetLogLoggerLevel(logLevel)
		},
	}

	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "Output format (text, json)")
	rootCmd.PersistentFlags().StringVar(&logLevelStr, "log-level", "error", "Set the log level (debug, info, warn, error)")

	rootCmd.AddCommand(
		carrier.NewCarrierCmd(),
	)

	ctx := context.Background()
	ctx, _ = cleanup.InterruptableContext(ctx)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		feedback.FatalError(err, 1)
	}
}

func ParseLogLevel(level string) (slog.Level, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		return 0, fmt.Errorf("invalid log level: %w", err)
	}
	return l, nil
}
