package cli

import "github.com/spf13/cobra"

var doJobProductionPlanCancellationReasonTypesCmd = &cobra.Command{
	Use:   "job-production-plan-cancellation-reason-types",
	Short: "Manage job production plan cancellation reason types",
	Long: `Create, update, and delete job production plan cancellation reason types.

Cancellation reason types define the reasons why a job production plan can be cancelled.

Note: Only admin users can create, update, or delete cancellation reason types.

Commands:
  create  Create a new cancellation reason type
  update  Update an existing cancellation reason type
  delete  Delete a cancellation reason type`,
	Example: `  # Create a cancellation reason type
  xbe do job-production-plan-cancellation-reason-types create --name "Weather" --slug "weather"

  # Update a cancellation reason type
  xbe do job-production-plan-cancellation-reason-types update 123 --name "Updated Name"

  # Delete a cancellation reason type
  xbe do job-production-plan-cancellation-reason-types delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanCancellationReasonTypesCmd)
}
