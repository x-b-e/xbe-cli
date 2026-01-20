package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/telemetry"
	"github.com/xbe-inc/xbe-cli/internal/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var rootCmd = &cobra.Command{
	Use:   "xbe",
	Short: "XBE CLI - Access XBE platform data and services",
	Long: `XBE CLI - Access XBE platform data and services

The XBE command-line interface provides programmatic access to the XBE platform,
enabling you to browse newsletters, manage broker data, and integrate XBE
capabilities into your workflows.

This CLI is designed for both interactive use and automation. All commands
support JSON output (--json) for easy parsing and integration with other tools.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Context keys for telemetry data
type contextKey string

const (
	spanKey      contextKey = "telemetry_span"
	startTimeKey contextKey = "telemetry_start_time"
)

// telemetryProvider holds the global telemetry provider
var telemetryProvider *telemetry.Provider

// lastExecutedCmd tracks the last command for telemetry finalization
var lastExecutedCmd *cobra.Command

func init() {
	initHelp(rootCmd)
	rootCmd.AddCommand(versionCmd)

	// Set up telemetry hook for span creation
	rootCmd.PersistentPreRunE = telemetryPreRun
}

// Execute runs the root command (for backward compatibility).
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext runs the root command with context and telemetry support.
func ExecuteContext(ctx context.Context, tp *telemetry.Provider) error {
	telemetryProvider = tp
	api.SetTelemetryProvider(tp)

	// Execute the command and capture the error
	err := rootCmd.ExecuteContext(ctx)

	// Finalize telemetry regardless of success/failure
	// This ensures spans are always closed and metrics recorded
	finalizeTelemetry(err)

	return err
}

func telemetryPreRun(cmd *cobra.Command, args []string) error {
	if telemetryProvider == nil || !telemetryProvider.Enabled() {
		return nil
	}

	ctx := cmd.Context()

	// Start a span for this command
	ctx, span := telemetryProvider.Tracer().Start(ctx,
		"xbe.command."+cmd.Name(),
		trace.WithAttributes(
			attribute.String("command.name", cmd.Name()),
			attribute.String("command.path", cmd.CommandPath()),
		),
	)

	// Store span and start time in context
	ctx = context.WithValue(ctx, spanKey, span)
	ctx = context.WithValue(ctx, startTimeKey, time.Now())
	cmd.SetContext(ctx)

	// Track this command for finalization
	lastExecutedCmd = cmd

	return nil
}

func finalizeTelemetry(cmdErr error) {
	if telemetryProvider == nil || !telemetryProvider.Enabled() {
		return
	}

	if lastExecutedCmd == nil {
		return
	}

	ctx := lastExecutedCmd.Context()

	// Get span from context
	span, ok := ctx.Value(spanKey).(trace.Span)
	if !ok || span == nil {
		return
	}

	// Get start time from context
	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	if !ok {
		startTime = time.Now() // Fallback
	}

	// Set span status based on command result
	success := cmdErr == nil
	if success {
		span.SetStatus(codes.Ok, "")
	} else {
		span.SetStatus(codes.Error, cmdErr.Error())
		span.SetAttributes(attribute.String("error.message", cmdErr.Error()))
	}
	span.End()

	// Record command metrics with correct success status
	telemetryProvider.RecordCommand(ctx, lastExecutedCmd.Name(), lastExecutedCmd.CommandPath(), success, time.Since(startTime))

	// Clear the tracked command
	lastExecutedCmd = nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Long: `Print the CLI version.

Displays the current version of the XBE CLI. Useful for debugging,
reporting issues, or verifying you have the latest version installed.`,
	Example: `  # Show version
  xbe version`,
	Annotations: map[string]string{"group": GroupUtility},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version.String())
		return nil
	},
}
