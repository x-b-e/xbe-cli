package cli

import "github.com/spf13/cobra"

var jobProductionPlanCancellationsCmd = &cobra.Command{
	Use:     "job-production-plan-cancellations",
	Aliases: []string{"job-production-plan-cancellation"},
	Short:   "View job production plan cancellations",
	Long: `View job production plan cancellations.

Cancellations record a status change to cancelled for a job production plan and
may include a cancellation reason type and comment.

Commands:
  list    List job production plan cancellations
  show    Show job production plan cancellation details`,
	Example: `  # List cancellations
  xbe view job-production-plan-cancellations list

  # Show a cancellation
  xbe view job-production-plan-cancellations show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanCancellationsCmd)
}
