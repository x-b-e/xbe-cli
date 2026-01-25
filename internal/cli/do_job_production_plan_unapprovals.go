package cli

import "github.com/spf13/cobra"

var doJobProductionPlanUnapprovalsCmd = &cobra.Command{
	Use:   "job-production-plan-unapprovals",
	Short: "Manage job production plan unapprovals",
	Long: `Create job production plan unapprovals.

Unapprovals move approved or scrapped job production plans back to a rejected status.

Commands:
  create    Unapprove a job production plan`,
	Example: `  # Unapprove a job production plan
  xbe do job-production-plan-unapprovals create --job-production-plan 123 --comment "Need revisions"`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanUnapprovalsCmd)
}
