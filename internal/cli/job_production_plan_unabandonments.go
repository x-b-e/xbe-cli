package cli

import "github.com/spf13/cobra"

var jobProductionPlanUnabandonmentsCmd = &cobra.Command{
	Use:     "job-production-plan-unabandonments",
	Aliases: []string{"job-production-plan-unabandonment"},
	Short:   "View job production plan unabandonments",
	Long: `View job production plan unabandonments.

Unabandonments record a status change from abandoned back to the previous
status for a job production plan and may include a comment.

Commands:
  list    List job production plan unabandonments
  show    Show job production plan unabandonment details`,
	Example: `  # List unabandonments
  xbe view job-production-plan-unabandonments list

  # Show an unabandonment
  xbe view job-production-plan-unabandonments show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanUnabandonmentsCmd)
}
