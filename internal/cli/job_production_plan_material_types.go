package cli

import "github.com/spf13/cobra"

var jobProductionPlanMaterialTypesCmd = &cobra.Command{
	Use:   "job-production-plan-material-types",
	Short: "View job production plan material types",
	Long: `View job production plan material types on the XBE platform.

Job production plan material types describe the materials, quantities, and
units planned for a job production plan.

Commands:
  list    List job production plan material types
  show    Show job production plan material type details`,
	Example: `  # List job production plan material types
  xbe view job-production-plan-material-types list

  # Show details for a material type on a plan
  xbe view job-production-plan-material-types show 123

  # Output JSON
  xbe view job-production-plan-material-types list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanMaterialTypesCmd)
}
