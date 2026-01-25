package cli

import "github.com/spf13/cobra"

var jobProductionPlanTruckingIncidentDetectorsCmd = &cobra.Command{
	Use:     "job-production-plan-trucking-incident-detectors",
	Aliases: []string{"job-production-plan-trucking-incident-detector"},
	Short:   "Browse job production plan trucking incident detectors",
	Long: `Browse job production plan trucking incident detectors on the XBE platform.

Trucking incident detectors analyze material transactions for a job production
plan and identify potential trucking incidents based on ordering inconsistencies.

Commands:
  list    List trucking incident detectors with filtering and pagination
  show    Show trucking incident detector details`,
	Example: `  # List trucking incident detectors
  xbe view job-production-plan-trucking-incident-detectors list

  # Filter by job production plan
  xbe view job-production-plan-trucking-incident-detectors list --job-production-plan 123

  # Filter by performed status
  xbe view job-production-plan-trucking-incident-detectors list --is-performed true

  # Show detector details
  xbe view job-production-plan-trucking-incident-detectors show 456

  # Output as JSON
  xbe view job-production-plan-trucking-incident-detectors list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanTruckingIncidentDetectorsCmd)
}
