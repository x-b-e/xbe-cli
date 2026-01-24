package cli

import "github.com/spf13/cobra"

var doProductionIncidentDetectorsCmd = &cobra.Command{
	Use:   "production-incident-detectors",
	Short: "Run production incident detection",
	Long: `Run production incident detection for a job production plan.

Production incident detectors analyze actual production versus plan and
return detected incident windows based on configurable thresholds.

Commands:
  create    Run detection and return incidents`,
	Example: `  # Run detection for a job production plan
  xbe do production-incident-detectors create --job-production-plan 123

  # Run with custom thresholds
  xbe do production-incident-detectors create \
    --job-production-plan 123 \
    --lookahead-offset 30 \
    --minutes-threshold 45 \
    --quantity-threshold 50`,
}

func init() {
	doCmd.AddCommand(doProductionIncidentDetectorsCmd)
}
