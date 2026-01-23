package cli

import "github.com/spf13/cobra"

var doJobProductionPlanMaterialSiteChangesCmd = &cobra.Command{
	Use:   "job-production-plan-material-site-changes",
	Short: "Manage job production plan material site changes",
	Long: `Create job production plan material site changes to swap material sites
(and optionally material types or mix designs) on a job production plan.

Commands:
  create    Create a material site change`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanMaterialSiteChangesCmd)
}
