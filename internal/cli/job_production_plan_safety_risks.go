package cli

import "github.com/spf13/cobra"

var jobProductionPlanSafetyRisksCmd = &cobra.Command{
	Use:   "job-production-plan-safety-risks",
	Short: "View job production plan safety risks",
	Long: `View job production plan safety risks on the XBE platform.

Safety risks capture potential hazards associated with a job production plan.

Commands:
  list    List job production plan safety risks with filtering
  show    Show job production plan safety risk details`,
	Example: `  # List job production plan safety risks
  xbe view job-production-plan-safety-risks list

  # Filter by job production plan
  xbe view job-production-plan-safety-risks list --job-production-plan 123

  # Show details
  xbe view job-production-plan-safety-risks show 456`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSafetyRisksCmd)
}
