package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xbe-inc/xbe-cli/internal/cli"
	"github.com/xbe-inc/xbe-cli/internal/telemetry"
)

func main() {
	// Create cancellable context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize telemetry (no-op if disabled)
	tp, err := telemetry.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: telemetry init failed: %v\n", err)
	}
	if tp != nil {
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer shutdownCancel()
			if err := tp.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: telemetry shutdown failed: %v\n", err)
			}
		}()
	}

	if err := cli.ExecuteContext(ctx, tp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
