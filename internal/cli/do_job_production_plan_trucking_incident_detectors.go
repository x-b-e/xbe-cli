package cli

import "github.com/spf13/cobra"

var doJobProductionPlanTruckingIncidentDetectorsCmd = &cobra.Command{
	Use:   "job-production-plan-trucking-incident-detectors",
	Short: "Run job production plan trucking incident detectors",
	Long: `Run job production plan trucking incident detectors.

Commands:
  create    Run a trucking incident detector`,
	Example: `  # Run detector for a job production plan
  xbe do job-production-plan-trucking-incident-detectors create --job-production-plan 123

  # Run detector as of a timestamp
  xbe do job-production-plan-trucking-incident-detectors create \
    --job-production-plan 123 \
    --as-of "2026-01-23T00:00:00Z"

  # Persist detected incident changes
  xbe do job-production-plan-trucking-incident-detectors create \
    --job-production-plan 123 \
    --persist-changes`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanTruckingIncidentDetectorsCmd)
}
