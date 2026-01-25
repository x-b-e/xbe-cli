package cli

import "github.com/spf13/cobra"

var doJobProductionPlanDuplicationsCmd = &cobra.Command{
	Use:   "job-production-plan-duplications",
	Short: "Duplicate job production plan templates",
	Long: `Create job production plan duplications to copy templates into new plans or templates.

Commands:
  create    Duplicate a job production plan template`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanDuplicationsCmd)
}
