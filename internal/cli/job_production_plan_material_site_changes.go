package cli

import "github.com/spf13/cobra"

var jobProductionPlanMaterialSiteChangesCmd = &cobra.Command{
	Use:   "job-production-plan-material-site-changes",
	Short: "Browse job production plan material site changes",
	Long: `Browse job production plan material site changes on the XBE platform.

Material site changes record swaps of material sites (and optional material types
or mix designs) for a job production plan.

Commands:
  list    List material site changes
  show    Show material site change details`,
	Example: `  # List recent material site changes
  xbe view job-production-plan-material-site-changes list

  # Show material site change details
  xbe view job-production-plan-material-site-changes show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanMaterialSiteChangesCmd)
}
