package cli

import "github.com/spf13/cobra"

var jobProductionPlanMaterialSitesCmd = &cobra.Command{
	Use:   "job-production-plan-material-sites",
	Short: "View job production plan material sites",
	Long: `View job production plan material sites on the XBE platform.

Job production plan material sites link a job production plan to the
material sites that supply materials for that plan.

Commands:
  list    List job production plan material sites with filtering
  show    Show job production plan material site details`,
	Example: `  # List job production plan material sites
  xbe view job-production-plan-material-sites list

  # Filter by job production plan
  xbe view job-production-plan-material-sites list --job-production-plan 123

  # Show details
  xbe view job-production-plan-material-sites show 456`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanMaterialSitesCmd)
}
