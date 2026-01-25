package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSafetyRisksCmd = &cobra.Command{
	Use:   "job-production-plan-safety-risks",
	Short: "Manage job production plan safety risks",
	Long: `Create, update, and delete job production plan safety risks.

Safety risks capture potential hazards associated with a job production plan.

Commands:
  create    Create a job production plan safety risk
  update    Update a job production plan safety risk
  delete    Delete a job production plan safety risk`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSafetyRisksCmd)
}
