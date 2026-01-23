package cli

import "github.com/spf13/cobra"

var doJobProductionPlanMaterialSitesCmd = &cobra.Command{
	Use:   "job-production-plan-material-sites",
	Short: "Manage job production plan material sites",
	Long: `Create, update, and delete job production plan material sites.

Job production plan material sites link a job production plan to the
material sites that supply materials for that plan.

Commands:
  create    Create a job production plan material site
  update    Update a job production plan material site
  delete    Delete a job production plan material site`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanMaterialSitesCmd)
}
