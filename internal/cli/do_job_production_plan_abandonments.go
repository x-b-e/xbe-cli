package cli

import "github.com/spf13/cobra"

var doJobProductionPlanAbandonmentsCmd = &cobra.Command{
	Use:   "job-production-plan-abandonments",
	Short: "Abandon job production plans",
	Long: `Abandon job production plans on the XBE platform.

Abandonments transition job production plans to abandoned status. Only plans
in editing, submitted, or rejected status can be abandoned.

Commands:
  create    Abandon a job production plan`,
	Example: `  # Abandon a job production plan
  xbe do job-production-plan-abandonments create --job-production-plan 123 --comment "No longer needed"

  # Abandon and suppress notifications
  xbe do job-production-plan-abandonments create --job-production-plan 123 --suppress-status-change-notifications`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanAbandonmentsCmd)
}
