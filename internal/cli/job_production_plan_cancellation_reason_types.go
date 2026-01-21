package cli

import "github.com/spf13/cobra"

var jobProductionPlanCancellationReasonTypesCmd = &cobra.Command{
	Use:   "job-production-plan-cancellation-reason-types",
	Short: "View job production plan cancellation reason types",
	Long: `View job production plan cancellation reason types on the XBE platform.

Cancellation reason types define the reasons why a job production plan can be cancelled.

Commands:
  list    List job production plan cancellation reason types`,
	Example: `  # List cancellation reason types
  xbe view job-production-plan-cancellation-reason-types list

  # Output as JSON
  xbe view job-production-plan-cancellation-reason-types list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanCancellationReasonTypesCmd)
}
